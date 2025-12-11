package main

import (
	"context"
	"sync"
	"time"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

type Room struct {
	ID string
	listLock sync.RWMutex
	clients map[string]*RTCClient
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
	trackRemotes map[string]*webrtc.TrackRemote
	cancelFunc context.CancelFunc
}

func addTrack(id string, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	room, ok := rooms.getRoom(id);if !ok {
		return nil
	}
	room.listLock.Lock()
	defer func() {
		room.listLock.Unlock()
		signalPeerConnections(id)
	}()
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}
	room.trackLocals[t.ID()] = trackLocal
	return trackLocal
}

func removeTrack(id string, t *webrtc.TrackLocalStaticRTP) {
	room, ok := rooms.getRoom(id);if !ok {
		return
	}
	room.listLock.Lock()
	defer func() {
		room.listLock.Unlock()
		signalPeerConnections(id)
	}()

	delete(room.trackLocals, t.ID())
}

func addRemoteTrack(id string, t *webrtc.TrackRemote) {
	room, ok := rooms.getRoom(id);if !ok {
		return
	}
	room.listLock.Lock()
	defer room.listLock.Unlock()
	room.trackRemotes[t.ID()] = t
}

func removeRemoteTrack(id string, t *webrtc.TrackRemote) {
	room, ok := rooms.getRoom(id);if !ok {
		return
	}
	room.listLock.Lock()
	defer room.listLock.Unlock()
	delete(room.trackRemotes, t.ID())
}

// dispatchKeyFrame sends a keyframe to all PeerConnections, used everytime a new user joins the call.
func dispatchKeyFrame(id string) {
    room, ok := rooms.getRoom(id)
    if !ok { return }

    // 収集だけロック下で
    type target struct{ pc *webrtc.PeerConnection; ssrc uint32 }
    var targets []target

    room.listLock.RLock()
    for _, c := range room.clients {
        for _, recv := range c.Peer.GetReceivers() {
            if tr := recv.Track(); tr != nil {
                targets = append(targets, target{pc: c.Peer, ssrc: uint32(tr.SSRC())})
            }
        }
    }
    room.listLock.RUnlock()

    // ロック外でRTCP送信
    for _, t := range targets {
        _ = t.pc.WriteRTCP([]rtcp.Packet{
            &rtcp.PictureLossIndication{MediaSSRC: t.ssrc},
        })
    }
}

func signalPeerConnections(id string) {
	room, ok := rooms.getRoom(id);if !ok {
        return
    }
	room.listLock.Lock()
	defer func() {
		room.listLock.Unlock()
		dispatchKeyFrame(id)
	}()
	attemptSync := func() (tryAgain bool) {
		for i := range room.clients {
			if room.clients[i].Peer.ConnectionState() == webrtc.PeerConnectionStateClosed {
				delete(room.clients, i)
				if len(room.clients) == 0 {
					delete(rooms.item, id)
				}
				return true // We modified the slice, start from the beginning
			}

			// map of sender we already are seanding, so we don't double send
			existingSenders := map[string]bool{}

			for _, sender := range room.clients[i].Peer.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := room.trackLocals[sender.Track().ID()]; !ok {
					if err := room.clients[i].Peer.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range room.clients[i].Peer.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range room.trackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := room.clients[i].Peer.AddTrack(room.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := room.clients[i].Peer.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = room.clients[i].Peer.SetLocalDescription(offer); err != nil {
				return true
			}

			if err = room.clients[i].WS.Send("offer", offer); err != nil {
				return true
			}
		}

		return tryAgain
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 10 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				signalPeerConnections(id)
			}()

			return
		}

		if !attemptSync() {
			break
		}
	}
}

type Rooms struct {
	item map[string]*Room
	lock sync.RWMutex
}

func (r *Rooms) getOrCreate(id string) *Room {
    r.lock.Lock()
    defer r.lock.Unlock()

    room := r.item[id]
    if room == nil {
        room = &Room{
            ID:          id,
            listLock:    sync.RWMutex{},
            clients:     make(map[string]*RTCClient),
            trackLocals: make(map[string]*webrtc.TrackLocalStaticRTP),
            trackRemotes: make(map[string]*webrtc.TrackRemote),
        }
        r.item[id] = room

        // ticker は作成時に一度だけ
        ctx, cancel := context.WithCancel(context.Background())
        room.cancelFunc = cancel
        go func() {
            t := time.NewTicker(3 * time.Second)
            defer t.Stop()
            for {
                select {
                case <-ctx.Done():
                    return
                case <-t.C:
                    dispatchKeyFrame(id) // ← この中も lock 外でPC操作するよう修正(後述)
                }
            }
        }()
    }
    return room
}

func (r *Rooms) getRoom(id string) (*Room, bool) {
    r.lock.RLock()
    room, ok := r.item[id]
    r.lock.RUnlock()
    if !ok {
        return nil, false
    }
    return room, true
}

func (r *Rooms) deleteRoom(id string) {
    r.lock.Lock()
    room, ok := r.item[id]
    if ok {
        if room.cancelFunc != nil { room.cancelFunc() }
        delete(r.item, id)
    }
    r.lock.Unlock()
}
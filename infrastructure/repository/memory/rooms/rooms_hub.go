package rooms_hub

import (
	"errors"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/webrtc/broadcast"
	"streaming-server.com/infrastructure/ws"
)

type Hub struct {
	rooms map[int]*Room // roomID -> runtime
	mu    sync.RWMutex
}

var (
	// _ live_video_hub.Interface = (*Hub)(nil)
	log = *logger.Log
)

func New() *Hub {
	rooms := make(map[int]*Room)
	room := NewRoom()
	rooms[1] = room
	return &Hub{ rooms, sync.RWMutex{} }
}

func (h *Hub) RoomExists(roomID int) bool {
	_, ok := h.rooms[roomID]
	return ok
}

func (h *Hub) getOrCreate(roomID int) *Room {
	rt := h.rooms[roomID]
	if rt == nil {
		rt = NewRoom()
		h.rooms[roomID] = rt
	}
	return rt
}

func (h *Hub) getRoom(roomID int) (*Room, error) {
	room, ok := h.rooms[roomID];if !ok {
		return nil, errors.New("room not found")
	}
	return room, nil
}

func (h *Hub) getClient(roomID, userID int) (*broadcast.PeerClient, error) {
	room, ok := h.rooms[roomID]; if !ok {
		return nil, errors.New("client not found")
	}
	client, err := room.getClient(userID);if err != nil {
		return nil, err
	}
	return client, nil
}

func (h *Hub) DeleteRoom(roomID int) {
	delete(h.rooms, roomID)
}

func (h *Hub) AddPeerConnection(roomId, userId int, peerClient *broadcast.PeerClient) error {
	room := h.getOrCreate(roomId)
	room.listLock.Lock()
	room.clients[userId] = peerClient
	room.listLock.Unlock()
	log.Debug("AddPeerConnection: %v", room.clients[userId].Peer.ConnectionState())
	return nil
}

func (h *Hub) AddICECandidate(roomId, userId int, candidate webrtc.ICECandidateInit) error {
	room, err := h.getRoom(roomId);if err != nil {
		return err
	}
	if err := room.clients[userId].Peer.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (h *Hub) SetRemoteDescription(roomId, userId int, sdp string) error {
	room, err := h.getRoom(roomId);if err != nil {
		return err
	}
	log.Debug("AddPeerConnection: %v", room.clients)
	client, ok := room.clients[userId];if !ok {
		return errors.New("no client")
	}
	if err := client.Peer.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}); err != nil {
		return err
	}
	return nil
}

// Add to list of tracks and fire renegotation for all PeerConnections.
func (h *Hub) AddTrack(roomId int, t *webrtc.TrackLocalStaticRTP) error {
	room, err := h.getRoom(roomId);if err != nil {
		return err
	}
	room.listLock.Lock()
	defer func() {
		room.listLock.Unlock()
		h.SignalPeerConnections(roomId)
	}()
	room.trackLocals[t.ID()] = t
	return nil
}

// Remove from list of tracks and fire renegotation for all PeerConnections.
func (h *Hub) RemoveTrack(roomId int, t *webrtc.TrackLocalStaticRTP) error {
	room, err := h.getRoom(roomId);if err != nil {
		return err
	}
	room.listLock.Lock()
	defer func() {
		room.listLock.Unlock()
		h.SignalPeerConnections(roomId)
	}()
	delete(room.trackLocals, t.ID())
	return nil
}

// signalPeerConnections updates each PeerConnection so that it is getting all the expected media tracks.
func (h *Hub) SignalPeerConnections(roomId int) error {
	room, err := h.getRoom(roomId);if err != nil {
		return err
	}
	room.listLock.Lock()
	defer func() {
		room.listLock.Unlock()
		h.dispatchKeyFrame(room)
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range room.clients {
			if room.clients[i].Peer.ConnectionState() == webrtc.PeerConnectionStateClosed {
				room.clients[i].Peer.Close()
				log.Error("delete peer: %v", room.clients[i].Peer.ConnectionState())
				delete(room.clients, i)
				return true
			}

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

			if err = room.clients[i].WS.WriteJSON(&ws.WebsocketMessage{
				Event: "offer",
				Data:  offer,
			}); err != nil {
				return true
			}
		}

		return tryAgain
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				h.SignalPeerConnections(roomId)
			}()

			return nil
		}

		if !attemptSync() {
			break
		}
	}
	return nil
}

// dispatchKeyFrame sends a keyframe to all PeerConnections, used everytime a new user joins the call.
func (h *Hub) dispatchKeyFrame(room *Room) {
	room.listLock.Lock()
	defer room.listLock.Unlock()

	for i := range room.clients {
		for _, receiver := range room.clients[i].Peer.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = room.clients[i].Peer.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

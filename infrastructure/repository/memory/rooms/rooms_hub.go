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
	return &Hub{rooms: make(map[int]*Room)}
}

func (h *Hub) RoomExists(roomID int) bool {
	_, ok := h.rooms[roomID]
	return ok
}

func (h *Hub) getOrCreate(roomID int) *Room {
	rt := h.rooms[roomID]
	if rt == nil {
		rt = &Room{
			clients: make(map[int]*RtcClient),
			tracks:  make(map[string]*webrtc.TrackLocalStaticRTP),
		}
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

// func (h *Hub) AddPeerConnection(roomID, userID int, pc *webrtc.PeerConnection) error {
// 	client, err := h.getClient(roomID, userID); if err != nil {
// 		return err
// 	}
// 	client.PeerConn = pc
// 	return nil
// }

func (h *Hub) getClient(roomID, userID int) (*RtcClient, error) {
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

// func (h *Hub) AddICECandidate(roomID, userID int, candidate webrtc.ICECandidateInit) error {
// 	client, err := h.getClient(roomID, userID); if err != nil {
// 		return err
// 	}
// 	if client.PeerConn == nil {
// 		return errors.New("no peer conn")
// 	}
// 	if err := client.PeerConn.AddICECandidate(candidate); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (h *Hub) SetRemoteDescription(roomID, userID int, sdp string) error {
// 	client, err := h.getClient(roomID, userID); if err != nil {
// 		return err
// 	}
// 	if err := client.PeerConn.SetRemoteDescription(webrtc.SessionDescription{
// 		Type: webrtc.SDPTypeAnswer,
// 		SDP:  sdp,
// 	}); err != nil {
// 		return err
// 	}
// 	return nil
// }

var (
	listLock = sync.RWMutex{}
	peerConnections = map[int]*broadcast.PeerConnectionState{}
	trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}
)


type peerConnectionState struct {
	peerConnection *webrtc.PeerConnection
	websocket      *ws.ThreadSafeWriter
}

func (h *Hub) AddPeerConnection(userId int, peerConnection *broadcast.PeerConnectionState) {
	listLock.Lock()
	peerConnections[userId] = peerConnection
	listLock.Unlock()
	log.Debug("AddPeerConnection: %v", peerConnections[userId].Peer.ConnectionState())
}

func (h *Hub) AddICECandidate(roomID, userID int, candidate webrtc.ICECandidateInit) error {
	log.Debug("add ice")
	if err := peerConnections[userID].Peer.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (h *Hub) SetRemoteDescription(roomID, userID int, sdp string) error {
	log.Debug("AddPeerConnection: %v", peerConnections)
	peer, ok := peerConnections[userID];if !ok {
		return errors.New("no peer")
	}
	if err := peer.Peer.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}); err != nil {
		return err
	}
	return nil
}

// Add to list of tracks and fire renegotation for all PeerConnections.
func (h *Hub) AddTrack(t *webrtc.TrackLocalStaticRTP) {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
		h.SignalPeerConnections()
	}()
	trackLocals[t.ID()] = t
}

// Remove from list of tracks and fire renegotation for all PeerConnections.
func (h *Hub) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
		h.SignalPeerConnections()
	}()

	delete(trackLocals, t.ID())
}

// signalPeerConnections updates each PeerConnection so that it is getting all the expected media tracks.
func (h *Hub) SignalPeerConnections() {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
		h.dispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range peerConnections {
			if peerConnections[i].Peer.ConnectionState() == webrtc.PeerConnectionStateClosed {
				peerConnections[i].Peer.Close()
				log.Error("delete peer: %v", peerConnections[i].Peer.ConnectionState())
				delete(peerConnections, i)
				return true
			}

			existingSenders := map[string]bool{}

			for _, sender := range peerConnections[i].Peer.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := trackLocals[sender.Track().ID()]; !ok {
					if err := peerConnections[i].Peer.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range peerConnections[i].Peer.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range trackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := peerConnections[i].Peer.AddTrack(trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := peerConnections[i].Peer.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = peerConnections[i].Peer.SetLocalDescription(offer); err != nil {
				return true
			}

			if err = peerConnections[i].WS.WriteJSON(&ws.WebsocketMessage{
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
				h.SignalPeerConnections()
			}()

			return
		}

		if !attemptSync() {
			break
		}
	}
}

// dispatchKeyFrame sends a keyframe to all PeerConnections, used everytime a new user joins the call.
func (h *Hub) dispatchKeyFrame() {
	listLock.Lock()
	defer listLock.Unlock()

	for i := range peerConnections {
		for _, receiver := range peerConnections[i].Peer.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = peerConnections[i].Peer.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

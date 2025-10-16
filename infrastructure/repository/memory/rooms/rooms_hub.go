package rooms_hub

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	"streaming-server.com/infrastructure/logger"
)

type Hub struct {
	rooms map[int]*Room // roomID -> runtime
	mu    sync.RWMutex
}

var (
	_ live_video_hub.Interface = (*Hub)(nil)
	log = *logger.Log
)

func New() *Hub {
	return &Hub{rooms: make(map[int]*Room)}
}

func (h *Hub) Join(roomID, userID int, conn *websocket.Conn) {
	h.addClient(roomID, userID, conn)
}

func (h *Hub) RemoveClient(roomID, userID int) error {
	room, err := h.getRoom(roomID);if err != nil {
		return err
	}
	client, err := h.getClient(roomID, userID);if err != nil {
		return err
	}
	room.RemoveTracks(userID)
	if client.HasPeerConnection() {
		client.ClosePeerConnection()
	}
	room.removeClient(userID)
	if !room.HasClient() {
		h.DeleteRoom(roomID)
	}
	return nil
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
			tracks:  make(map[int]*Tracks),
		}
		h.rooms[roomID] = rt
	}
	return rt
}

func (h *Hub) addClient(roomID, userID int, conn *websocket.Conn) {
	room := h.getOrCreate(roomID)
	room.addClient(userID, conn)
}

func (h *Hub) getRoom(roomID int) (*Room, error) {
	room, ok := h.rooms[roomID];if !ok {
		return nil, errors.New("room not found")
	}
	return room, nil
}

// TODO VideoとAudioでメソッド分ける必要なさそう
func (h *Hub) setVideoTrack(roomID, userID int, localTrack *webrtc.TrackLocalStaticRTP) error {
	room, err := h.getRoom(roomID);if err != nil {
		return err
	}
	track, err := room.GetTrack(userID);if err != nil {
		room.tracks[userID] = &Tracks{
			localTrack,
			nil,
		}
	} else {
		track.Video = localTrack
	}
	return nil
}

func (h *Hub) setAudioTrack(roomID, userID int, localTrack *webrtc.TrackLocalStaticRTP) error {
	room, err := h.getRoom(roomID);if err != nil {
		return err
	}

	track, err := room.GetTrack(userID);if err != nil {
		room.tracks[userID] = &Tracks{
			nil,
			localTrack,
		}
	} else {
		track.Audio = localTrack
	}
	return nil
}

func (h *Hub) SetTrack(roomID, userID int, localTrack *webrtc.TrackLocalStaticRTP, track *webrtc.TrackRemote) error {
	var err error
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		err = h.setVideoTrack(roomID, userID, localTrack)
	} else if track.Kind() == webrtc.RTPCodecTypeAudio {
		err = h.setAudioTrack(roomID, userID, localTrack)
	}
	if err != nil {
		log.Debug("%v", err)
		return err
	}
    go func() {
        buf := make([]byte, 1500)
        packetCount := 0
        for {
            n, _, err := track.Read(buf)
            if err != nil {
                log.Debug("Track read error: %v", err)
                return
            }
            packetCount++
            if packetCount%1000 == 0 {
                log.Debug("📦 Received %d RTP packets (%d bytes)", packetCount, n) //送信側のストリーミング確認ログ
            }
			if _, err = localTrack.Write(buf[:n]); err != nil {
				break
			}
        }
    }()
	room, _ := h.getRoom(roomID)
	client, _ := room.getClient(userID)
	for _, viewer := range room.clients {
		if viewer.PeerConn == client.PeerConn {
			continue
		}

		// 4-2) Sender が無い場合は AddTrack → Stable の時だけ 1 回だけ再交渉
		if _, err := viewer.PeerConn.AddTrack(localTrack); err != nil {
			log.Error("AddTrack to viewer:", err)
			continue
		}
		log.Debug("AddTrack to Viewer UserID: %v", viewer.UserID)

		//　AddTrackしてもクライアントでイベントが発火するわけではないため再オファーが必要
		offer, err := viewer.PeerConn.CreateOffer(nil)
		if err != nil {
			log.Error("ReOffer error:", err)
			continue
		}
		_ = viewer.PeerConn.SetLocalDescription(offer)

		// WebSocket経由で viewer に送信
		message := struct {
			Type string `json:"type"`
			Data struct {
				RoomID int    `json:"roomId"`
				SDP    string `json:"sdp"`
			} `json:"data"`
		}{
			Type: "offer",
			Data: struct {
				RoomID int    `json:"roomId"`
				SDP    string `json:"sdp"`
			}{
				RoomID: roomID,
				SDP:    offer.SDP,
			},
		}
		_ = viewer.Conn.WriteJSON(message)
			}
			return nil
		}

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

func (h *Hub) AddPeerConnection(roomID, userID int, pc *webrtc.PeerConnection) error {
	client, err := h.getClient(roomID, userID); if err != nil {
		return err
	}
	client.PeerConn = pc
	return nil
}

func (h *Hub) AddICECandidate(roomID, userID int, candidate webrtc.ICECandidateInit) error {
	client, err := h.getClient(roomID, userID); if err != nil {
		return err
	}
	if client.PeerConn == nil {
		return errors.New("no peer conn")
	}
	if err := client.PeerConn.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (h *Hub) SetRemoteDescription(roomID, userID int, sdp string) error {
	client, err := h.getClient(roomID, userID); if err != nil {
		return err
	}
	if err := client.PeerConn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	}); err != nil {
		return err
	}
	return nil
}

func (h *Hub) AddPublisherTracks(roomID, userID int, pc *webrtc.PeerConnection) error {
	// 既に room にトラックがあれば初回 offer に含める
	room, err := h.getRoom(roomID); if err != nil {
		return err
	}
	for _, track := range room.tracks {
		if track.Video != nil {
			pc.AddTrack(track.Video)
		}
		if track.Audio != nil {
			pc.AddTrack(track.Audio)
		}
	}
	return nil
}
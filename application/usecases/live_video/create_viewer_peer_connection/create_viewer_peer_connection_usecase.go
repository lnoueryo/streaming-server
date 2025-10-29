package create_viewer_peer_connection_usecase

import (
	"github.com/pion/webrtc/v4"
	room_memory_repository "streaming-server.com/application/ports/repositories/memory"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	user_entity "streaming-server.com/domain/entities/user"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/webrtc/broadcast"
	"streaming-server.com/infrastructure/ws"
)

var (
	log = logger.Log
)

type CreateViewerPeerConnectionUsecase struct {
	roomRepository room_memory_repository.IRoomRepository
}

func NewCreateViewerPeerConnection(roomRepo room_memory_repository.IRoomRepository) *CreateViewerPeerConnectionUsecase {
	return &CreateViewerPeerConnectionUsecase{
		roomRepo,
	}
}

func (u *CreateViewerPeerConnectionUsecase) Do(
	params *live_video_dto.Params,
	conn *ws.ThreadSafeWriter,
) error {
	pcs := broadcast.NewPeerConnection(conn)
	// defer pcs.Peer.Close()
	room := u.roomRepository.GetOrCreate(params.RoomID)
	logger.Log.Debug("%v", room)
	user := &user_entity.RuntimeUser{
		params.UserID,
		pcs,
	}
	room.AddUser(user)

	pcs.Peer.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		if writeErr := conn.Send("candidate", i.ToJSON()); writeErr != nil {
			log.Error("Failed to write JSON: %v", writeErr)
		}
	})

	pcs.Peer.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		log.Info("Connection state change: %s", p)

		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := pcs.Peer.Close(); err != nil {
				log.Error("Failed to close PeerConnection: %v", err)
			}
		case webrtc.PeerConnectionStateClosed:
			room.SignalPeerConnections()
		default:
		}
	})

	pcs.Peer.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		log.Info("ICE connection state changed: %s", is)
	})
	room.SignalPeerConnections()
	return nil
}

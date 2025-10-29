package close_connection_usecase

import (
	room_memory_repository "streaming-server.com/application/ports/repositories/memory"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/ws"
)

var log = logger.Log

type CloseConnectionUsecase struct {
	roomRepository room_memory_repository.IRoomRepository
}

func NewCloseConnection(
	roomRepo room_memory_repository.IRoomRepository,
) *CloseConnectionUsecase {
	return &CloseConnectionUsecase{
		roomRepo,
	}
}

func (u *CloseConnectionUsecase) Do(
	params *live_video_dto.Params,
	conn *ws.ThreadSafeWriter,
) error {
	log.Debug("🧩 RemoveClient called: room=%d user=%d", params.RoomID, params.UserID)
	room, err := u.roomRepository.GetRoom(params.RoomID);if err != nil {
		return err
	}
	user, err := room.GetClient(params.UserID);if err != nil {
		return err
	}
	user.Peer.Close()
	user.WS.Close()
	if len(room.Users) == 0 {
		u.roomRepository.DeleteRoom(params.RoomID)
	}
	room.SignalPeerConnections()
	return nil
}

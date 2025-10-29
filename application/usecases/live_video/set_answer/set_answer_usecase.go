package set_answer_usecase

import (
	room_memory_repository "streaming-server.com/application/ports/repositories/memory"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/ws"
)

var log = logger.Log

type SetAnswerUsecase struct {
	roomRepository room_memory_repository.IRoomRepository
}

func NewSetAnswer(roomRepo room_memory_repository.IRoomRepository) *SetAnswerUsecase {
	return &SetAnswerUsecase{
		roomRepo,
	}
}

func (u *SetAnswerUsecase) Do(
	params *live_video_dto.Params,
	message *Message,
	conn *ws.ThreadSafeWriter,
) error {
	room, err := u.roomRepository.GetRoom(params.RoomID);if err != nil {
		log.Error("%v", err)
	}
	room.SetRemoteDescription(params.UserID, message.SDP);if err != nil {
		log.Error("%v", err)
	}
	log.Debug("send answered")
	return nil
}

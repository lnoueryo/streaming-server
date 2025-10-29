package set_candidate_usecase

import (
	"github.com/pion/webrtc/v4"
	room_memory_repository "streaming-server.com/application/ports/repositories/memory"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/ws"
)

var log = logger.Log

type SetCandidateUsecase struct {
	roomRepository room_memory_repository.IRoomRepository
}

func NewSetCandidate(roomRepo room_memory_repository.IRoomRepository) *SetCandidateUsecase {
	return &SetCandidateUsecase{
		roomRepo,
	}
}

func (u *SetCandidateUsecase) Do(
	params *live_video_dto.Params,
	message *Message,
	conn *ws.ThreadSafeWriter,
) error {
	cand := webrtc.ICECandidateInit{
		Candidate:     message.Candidate,
		SDPMid:        message.SDPMid,
		SDPMLineIndex: message.SDPMLineIndex,
	}

	room, err := u.roomRepository.GetRoom(params.RoomID);if err != nil {
		log.Error("%v", err)
	}
	err = room.AddICECandidate(params.UserID, cand);if err != nil {
		log.Error("%v", err)
	}
	log.Info("👌 Send: Set Candidate")
	return nil
}

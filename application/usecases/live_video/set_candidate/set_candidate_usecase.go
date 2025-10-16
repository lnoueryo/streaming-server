package set_candidate_usecase

import (
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	"streaming-server.com/infrastructure/logger"
)

var log = logger.Log

type SetCandidateUsecase struct {
	roomRepository live_video_hub.Interface
}

func NewSetCandidate(roomRepo live_video_hub.Interface) *SetCandidateUsecase {
	return &SetCandidateUsecase{
		roomRepo,
	}
}

func (u *SetCandidateUsecase) Do(
	conn *websocket.Conn,
	params *SetCandidateInput,
) error {
	cand := webrtc.ICECandidateInit{
		Candidate:     params.Candidate,
		SDPMid:        params.SDPMid,
		SDPMLineIndex: params.SDPMLineIndex,
	}

	err := u.roomRepository.AddICECandidate(params.RoomID, params.UserID, cand);if err != nil {
		log.Error("%v", err)
	}
	// TODO どこかで共通化
	message := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: "set_candidate",
		Data: struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		}{
			RoomID:      params.RoomID,
			PublisherID: "publisherID",
		},
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Error("WriteJSON error:", err)
		return err
	}
	log.Info("👌 Send: Set Candidate")
	return nil
}

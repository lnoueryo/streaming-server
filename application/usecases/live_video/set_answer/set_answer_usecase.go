package set_answer_usecase

import (
	"github.com/gorilla/websocket"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	"streaming-server.com/infrastructure/logger"
)

var log = logger.Log

type SetAnswerUsecase struct {
	roomRepository live_video_hub.Interface
}

func NewSetAnswer(roomRepo live_video_hub.Interface) *SetAnswerUsecase {
	return &SetAnswerUsecase{
		roomRepo,
	}
}

func (u *SetAnswerUsecase) Do(
	conn *websocket.Conn,
	param *SetAnswerInput,
) error {
	err := u.roomRepository.SetRemoteDescription(param.RoomID, param.UserID, param.SDP);if err != nil {
		log.Error("%v", err)
	}
	// TODO どこかで共通化
	message := struct {
		Type string `json:"type"`
		Data struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		} `json:"data"`
	}{
		Type: "answered",
		Data: struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		}{
			RoomID:      param.RoomID,
			PublisherID: "publisherID",
		},
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Error("WriteJSON error:", err)
		return err
	}
	log.Debug("send answered")
	return nil
}

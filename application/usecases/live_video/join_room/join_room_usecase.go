package join_room_usecase

import (
	"github.com/gorilla/websocket"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	"streaming-server.com/infrastructure/logger"
)

var log = logger.Log

type JoinRoomUsecase struct {
	roomRepository live_video_hub.Interface
}

func NewJoinRoom(
	roomRepo live_video_hub.Interface,
) *JoinRoomUsecase {
	return &JoinRoomUsecase{
		roomRepo,
	}
}

func (u *JoinRoomUsecase) Do(
	conn *websocket.Conn,
	param *JoinRoomInput,
) error {
	// room の認証が必要な場合ここ
	u.roomRepository.Join(param.RoomID, param.UserID, conn)
	conn.SetCloseHandler(func(code int, text string) error {
		log.Info("Request: Close Connection")
		u.roomRepository.RemoveClient(param.RoomID, param.UserID)
		return nil
	})
	// TODO どこかで共通化
	message := struct {
		Type string `json:"type"`
		Data struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		} `json:"data"`
	}{
		Type: "join",
		Data: struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		}{
			RoomID:      param.RoomID,
			PublisherID: "publisherID",
		},
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Error("WriteJSON error: ", err)
		return err
	}
	log.Info("👌 Send: Join")
	return nil
}

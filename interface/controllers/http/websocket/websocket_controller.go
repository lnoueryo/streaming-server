package websocket_controller

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	// "github.com/pion/webrtc/v4"
	live_video_controller "streaming-server.com/interface/controllers/websocket/live_video"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Controller struct {
	LiveVideoController *live_video_controller.Controller
}

func NewController(
	liveVideoController *live_video_controller.Controller,
) *Controller {
	c := &Controller{}
	c.LiveVideoController = liveVideoController
	return c
}

func (c *Controller) CreateLiveVideoWebsocket(w http.ResponseWriter, r *http.Request) {
	// TODO 認証
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer conn.Close()

	var msg struct { Type string  `json:"type"` }
	// var msg struct {
	// 	Type          string  `json:"type"`
	// 	UserID        int     `json:"userId"`
	// 	RoomID        int     `json:"roomId"`
	// 	SDP           string  `json:"sdp"`
	// 	Candidate     string  `json:"candidate"`
	// 	SDPMid        *string `json:"sdpMid"`
	// 	SDPMLineIndex *uint16 `json:"sdpMLineIndex"`
	// }

	for {
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("read:", err)
			// c.LiveVideoController.CloseConnection(conn, msg)
			return
		}
		log.Println(msg.Type)
		switch msg.Type {
		case "join":
			c.LiveVideoController.JoinRoom(conn, msg)
			conn.SetCloseHandler(func(code int, text string) error {
				log.Printf("connection closed: code=%d, text=%s", code, text)
				// c.LiveVideoController.CloseConnection(conn, msg)
				return nil
			})
		case "offer":
			c.LiveVideoController.GetOffer(conn, msg)
		case "answer":
			c.LiveVideoController.SetAnswer(conn, msg)
		case "candidate":
			c.LiveVideoController.SetCandidate(conn, msg)
		case "viewer":
			c.LiveVideoController.CreateViewPeerConnection(conn, msg)
		}
	}
}

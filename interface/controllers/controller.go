package controllers

import (
	websocket_controller "streaming-server.com/interface/controllers/http/websocket"
	live_video_controller "streaming-server.com/interface/controllers/websocket/live_video"
)

type Controllers struct {
	LiveVideoController *live_video_controller.Controller
	WebsocketController *websocket_controller.Controller
}

func NewControllers(
	liveVideoController *live_video_controller.Controller,
	websocketController *websocket_controller.Controller,
) *Controllers {
	return &Controllers{
		liveVideoController,
		websocketController,
	}
}
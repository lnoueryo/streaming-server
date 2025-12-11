package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"
)

func deleteRtcClient(c *gin.Context) {
	user := getUser(c)
	roomId := c.Param("roomId")
	userId := user.ID
	room, ok := rooms.getRoom(roomId); if !ok {
		err := &ErrorResponse{
			"トークルームが既に存在しません",
			http.StatusNotFound,
			"no-target-room",
		}
		err.response(c)
		return
	}
	client, ok := room.clients[userId]; if !ok {
		err := &ErrorResponse{
			"ユーザーは既にトークルームから退出しています",
			http.StatusNotFound,
			"no-target-user",
		}
		err.response(c)
		return
	}
	room.listLock.Lock()
	var data interface{}
	client.WS.Send("close", data)
	client.Peer.Close()
	client.WS.Close()
	delete(room.clients, userId)
	room.listLock.Unlock()
	c.JSON(http.StatusNoContent, gin.H{})
}

func getRoom(c *gin.Context) {
	roomId := c.Param("roomId")
	room, ok := rooms.getRoom(roomId); if !ok {
		err := &ErrorResponse{
			"roomが存在しません",
			404,
			"not-found",
		}
		err.response(c)
		return
	}
	users := make([]gin.H, 0, len(room.clients))

	for _, user := range room.clients {
		if user.Peer.ConnectionState() == webrtc.PeerConnectionStateConnected {
			users = append(users, gin.H{
				"id": user.ID,
				"name": user.Name,
				"email": user.Email,
				"image": user.Image,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": room.ID,
		"users": users,
	})
}

func getUser(c *gin.Context) UserInfo {
	userVal, exists := c.Get("user")
	if !exists {
		return UserInfo{}
	}
	return userVal.(UserInfo)
}

type ErrorResponse struct {
	message string
	statusCode int
	errorCode string
}

func (er *ErrorResponse)response(c *gin.Context) {
	c.JSON(er.statusCode, gin.H{
		"message": er.message,
		"statusCode": er.statusCode,
		"errorCode": er.errorCode,
	})
}
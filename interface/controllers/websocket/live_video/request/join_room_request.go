package live_video_request

import (
	"encoding/json"
	"errors"
	"strconv"

	join_room_usecase "streaming-server.com/application/usecases/live_video/join_room"
)

type joinRoomRaw struct {
	RoomID interface{} `json:"roomId"`
	UserID interface{} `json:"userId"`
}

func JoinRoomRequest(message interface{}) (*join_room_usecase.JoinRoomInput, error) {
	var req joinRoomRaw
	var param join_room_usecase.JoinRoomInput
    raw, err := json.Marshal(message)
    if err != nil {
        return &param, err
    }

	if err := json.Unmarshal(raw, &req); err != nil {
		return &param, err
	}

	return req.createParam()
}

func (raw *joinRoomRaw) createParam() (*join_room_usecase.JoinRoomInput, error) {
	var input join_room_usecase.JoinRoomInput
	roomID, err := raw.getRoomId();if err != nil {
		return &input, err
	}
	userID, err := raw.getUserId();if err != nil {
		return &input, err
	}
	input.RoomID = roomID
	input.UserID = userID
	return &input, nil
}

func (raw *joinRoomRaw) getRoomId() (int, error) {
	var id int
	switch v := raw.RoomID.(type) {
	case float64:
		id = int(v)
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return id, errors.New("invalid roomId format")
		}
	default:
		return id, errors.New("roomId must be a number or numeric string")
	}

	if id <= 0 {
		return id, errors.New("invalid roomId value")
	}
	return id, nil
}

func (raw *joinRoomRaw) getUserId() (int, error) {
	var id int
	switch v := raw.UserID.(type) {
	case float64:
		id = int(v)
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return id, errors.New("invalid userId format")
		}
	default:
		return id, errors.New("userId must be a number or numeric string")
	}

	if id <= 0 {
		return id, errors.New("invalid userId value")
	}
	return id, nil
}
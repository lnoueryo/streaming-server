package live_video_request

import (
	"encoding/json"
	"errors"
	"strconv"

	create_viewer_peer_connection_usecase "streaming-server.com/application/usecases/live_video/create_viewer_peer_connection"
)

type createViewerPeerConnectionRaw struct {
	RoomID interface{} `json:"roomId"`
	UserID interface{} `json:"userId"`
}

func CreateViewerPeerConnectionRequest(message interface{}) (*create_viewer_peer_connection_usecase.CreateViewerPeerConnectionInput, error) {
	var req createViewerPeerConnectionRaw
	var param create_viewer_peer_connection_usecase.CreateViewerPeerConnectionInput
    raw, err := json.Marshal(message)
    if err != nil {
        return &param, err
    }

	if err := json.Unmarshal(raw, &req); err != nil {
		return &param, err
	}

	return req.createParam()
}

func (raw *createViewerPeerConnectionRaw) createParam() (*create_viewer_peer_connection_usecase.CreateViewerPeerConnectionInput, error) {
	var input create_viewer_peer_connection_usecase.CreateViewerPeerConnectionInput
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

func (raw *createViewerPeerConnectionRaw) getRoomId() (int, error) {
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

func (raw *createViewerPeerConnectionRaw) getUserId() (int, error) {
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
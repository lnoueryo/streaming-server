package live_video_request

import (
	"encoding/json"
	"errors"
	"strconv"

	get_offer_usecase "streaming-server.com/application/usecases/live_video/get_offer"
)

type getOfferRaw struct {
	RoomID interface{} `json:"roomId"`
	UserID interface{} `json:"userId"`
	SDP string `json:"sdp"`
}

func GetOfferRequest(message interface{}) (*get_offer_usecase.GetOfferInput, error) {
	var req getOfferRaw
	var param get_offer_usecase.GetOfferInput
    raw, err := json.Marshal(message)
    if err != nil {
        return &param, err
    }

	if err := json.Unmarshal(raw, &req); err != nil {
		return &param, err
	}

	return req.createParam()
}

func (raw *getOfferRaw) createParam() (*get_offer_usecase.GetOfferInput, error) {
	var input get_offer_usecase.GetOfferInput
	roomID, err := raw.getRoomId();if err != nil {
		return &input, err
	}
	userID, err := raw.getUserId();if err != nil {
		return &input, err
	}
	input.RoomID = roomID
	input.UserID = userID
	input.SDP = raw.SDP
	return &input, nil
}

func (raw *getOfferRaw) getRoomId() (int, error) {
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

func (raw *getOfferRaw) getUserId() (int, error) {
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
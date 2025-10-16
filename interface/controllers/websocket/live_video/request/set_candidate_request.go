package live_video_request

import (
	"encoding/json"
	"errors"
	"strconv"

	set_candidate_usecase "streaming-server.com/application/usecases/live_video/set_candidate"
)

type setCandidateRaw struct {
	RoomID interface{} `json:"roomId"`
	UserID interface{} `json:"userId"`
	Candidate     string  `json:"candidate"`
	SDPMid        *string `json:"sdpMid"`
	SDPMLineIndex *uint16 `json:"sdpMLineIndex"`
}

func SetCandidateRequest(message interface{}) (*set_candidate_usecase.SetCandidateInput, error) {
	var req setCandidateRaw
	var param set_candidate_usecase.SetCandidateInput
    raw, err := json.Marshal(message)
    if err != nil {
        return &param, err
    }

	if err := json.Unmarshal(raw, &req); err != nil {
		return &param, err
	}
	return req.createParam()
}

func (raw *setCandidateRaw) createParam() (*set_candidate_usecase.SetCandidateInput, error) {
	var input set_candidate_usecase.SetCandidateInput
	roomID, err := raw.getRoomId();if err != nil {
		return &input, err
	}
	userID, err := raw.getUserId();if err != nil {
		return &input, err
	}
	input.RoomID = roomID
	input.UserID = userID
	input.Candidate = raw.Candidate
	input.SDPMid = raw.SDPMid
	input.SDPMLineIndex = raw.SDPMLineIndex
	return &input, nil
}

func (raw *setCandidateRaw) getRoomId() (int, error) {
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

func (raw *setCandidateRaw) getUserId() (int, error) {
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
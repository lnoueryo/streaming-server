package get_offer_usecase

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	"streaming-server.com/infrastructure/logger"
)

var log = logger.Log

type GetOfferUsecase struct {
	roomRepository live_video_hub.Interface
}

func NewGetOffer(roomRepo live_video_hub.Interface,) *GetOfferUsecase {
	return &GetOfferUsecase{
		roomRepo,
	}
}

func (u *GetOfferUsecase) Do(
	params *live_video_dto.Params,
	message *Message,
	conn *websocket.Conn,
) error {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})
	if err != nil {
		log.Error("pc:", err)
		return err
	}
	pc.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		log.Debug("📡 Track received from publisher:", track.Kind().String())

		// 1) TrackID は必ず Kind().String() を使う！
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			fmt.Sprintf("%s-%d", track.Kind().String(), params.UserID), // ← unique
			"pion",
		)
		if err != nil {
			log.Error("NewTrackLocalStaticRTP error:", err)
			return
		}

		// 2) room へ差し替え（append しない）
		u.roomRepository.SetTrack(params.RoomID, params.UserID, localTrack, track);if err != nil {
			log.Error("%v", err)
		}
	})
	// TODO 順番の確認とmessageのフォーマット変更

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		log.Info("📥 Request: ws/live OnICECandidate")
		if c != nil {
			message := struct {
				Type string                   `json:"type"`
				Data webrtc.ICECandidateInit `json:"data"`
			}{
				Type: "candidate",
				Data: c.ToJSON(),
			}
			if err := conn.WriteJSON(message); err != nil {
				log.Error("send candidate error:", err)
			}
			log.Info("👌 Send: Candidate")
		}
	})

	u.roomRepository.AddPeerConnection(params.RoomID, params.UserID, pc)
	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: message.SDP}
	if err := pc.SetRemoteDescription(offer); err != nil {
		log.Error("setRemote:", err)
		return err
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Error("createAnswer:", err)
		return err
	}
	g := webrtc.GatheringCompletePromise(pc)
	_ = pc.SetLocalDescription(answer)
	<-g

	msg := struct {
		Type string `json:"type"`
		Data struct {
			RoomID int    `json:"roomId"`
			UserID int    `json:"userId"`
			SDP    string `json:"sdp"`
		} `json:"data"`
	}{
		Type: "answer",
		Data: struct {
			RoomID int    `json:"roomId"`
			UserID int    `json:"userId"`
			SDP    string `json:"sdp"`
		}{
			RoomID: params.RoomID,
			UserID: params.UserID,
			SDP:    answer.SDP,
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Info("WriteJSON error:", err)
		return err
	}
	log.Info("👌 Send: Answer")
	return nil
}

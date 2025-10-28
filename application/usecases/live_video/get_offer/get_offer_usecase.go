package get_offer_usecase

import (
	"encoding/json"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/webrtc/broadcast"
	"streaming-server.com/infrastructure/ws"
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
	conn *ws.ThreadSafeWriter,
) error {
	pcs := broadcast.NewPeerConnection(conn)
	defer pcs.Peer.Close()
	u.roomRepository.AddPeerConnection(params.UserID, pcs)

	// Trickle ICE. Emit server candidate to client
	pcs.Peer.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		if writeErr := conn.Send("candidate", i.ToJSON()); writeErr != nil {
			log.Error("Failed to write JSON: %v", writeErr)
		}
	})

	// If PeerConnection is closed remove it from global list
	pcs.Peer.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		log.Info("Connection state change: %s", p)

		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := pcs.Peer.Close(); err != nil {
				log.Error("Failed to close PeerConnection: %v", err)
			}
		case webrtc.PeerConnectionStateClosed:
			u.roomRepository.SignalPeerConnections()
		default:
		}
	})

	pcs.Peer.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		log.Info("ICE connection state changed: %s", is)
	})
	pcs.Peer.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		log.Info("Got remote track: Kind=%s, ID=%s, PayloadType=%d", t.Kind(), t.ID(), t.PayloadType())

		// Create a track to fan out our incoming video to all peers
		trackLocal, err := pcs.Peer.CreateLocalTrack(t);if err != nil {

		}
		u.roomRepository.AddTrack(trackLocal)
		defer u.roomRepository.RemoveTrack(trackLocal)

		buf := make([]byte, 1500)
		rtpPkt := &rtp.Packet{}

		for {
			i, _, err := t.Read(buf)
			if err != nil {
				return
			}

			if err = rtpPkt.Unmarshal(buf[:i]); err != nil {
				log.Error("Failed to unmarshal incoming RTP packet: %v", err)

				return
			}

			rtpPkt.Extension = false
			rtpPkt.Extensions = nil

			if err = trackLocal.WriteRTP(rtpPkt); err != nil {
				return
			}
		}
	})
	// Signal for the new PeerConnection
	u.roomRepository.SignalPeerConnections()

	msg := &ws.WebsocketMessage{}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Error("Failed to read message: %v", err)
			return err
		}

		log.Info("Got message: %s", raw)

		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Error("Failed to unmarshal json to message: %v", err)
			return err
		}

		switch msg.Event {
		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			rawData, err := json.Marshal(msg.Data)
			if err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(rawData), &candidate); err != nil {
				log.Error("Failed to unmarshal json to candidate: %v", err)
				return err
			}

			log.Info("Got candidate: %v", candidate)

			if err := pcs.Peer.AddICECandidate(candidate); err != nil {
				log.Error("Failed to add ICE candidate: %v", err)
				return err
			}
		case "answer":
			answer := webrtc.SessionDescription{}
			rawData, err := json.Marshal(msg.Data)
			if err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(rawData), &answer); err != nil {
				log.Error("Failed to unmarshal json to answer: %v", err)
				return err
			}

			log.Info("Got answer: %v", answer)

			if err := pcs.Peer.SetRemoteDescription(answer); err != nil {
				log.Error("Failed to set remote description: %v", err)
				return err
			}
		default:
			log.Error("unknown message: %+v", msg)
		}
	}
}

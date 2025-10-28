package broadcast

import (
	"github.com/pion/webrtc/v4"
	"streaming-server.com/infrastructure/ws"
)

type PeerClient struct {
	Peer *PeerConnection
	WS *ws.ThreadSafeWriter
}

type PeerConnection struct {
	*webrtc.PeerConnection
}

func NewPeerConnection(conn *ws.ThreadSafeWriter) *PeerClient {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil
	}

	// Accept one audio and one video track incoming
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			return nil
		}
	}
	return &PeerClient{
		&PeerConnection{ pc },
		conn,
	}
}

func (pc *PeerConnection) CreateLocalTrack(t *webrtc.TrackRemote) (*webrtc.TrackLocalStaticRTP, error) {
		trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
		if err != nil {
			return nil, err
		}
	return trackLocal, nil
}

func (pc *PeerConnection) InitLocalOffer() error {
	offer, err := pc.CreateOffer(nil);if err != nil {
		return err
	}

	if err = pc.SetLocalDescription(offer);err != nil {
		return err
	}
	return nil
}

func (pc *PeerConnection) IsConnectionClosed() bool {
	return pc.ConnectionState() == webrtc.PeerConnectionStateClosed
}

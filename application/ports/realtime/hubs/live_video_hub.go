package live_video_hub

import (
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type RtcClient struct {
    UserID   int
    Conn     *websocket.Conn
    PeerConn *webrtc.PeerConnection
}

type Interface interface {
    // RemoveClient(roomID, userID int)
    // Join(roomID, userID int, conn *websocket.Conn)
    // RoomExists(roomID int) bool
    // AddPeerConnection(roomID, userID int, pc *webrtc.PeerConnection)
    // SetVideoTrack(roomID, userID int, localTrack *webrtc.TrackLocalStaticRTP)
    // SetAudioTrack(roomID, userID int, localTrack *webrtc.TrackLocalStaticRTP)
    // AddPublisherTracks(roomID, userID int, pc *webrtc.PeerConnection)
    // SetRemoteDescription(roomID, userID int, sdp string)

    AddPeerConnection(roomID int, userID int, pc *webrtc.PeerConnection) error
    AddPublisherTracks(roomID int, userID int, pc *webrtc.PeerConnection) error
    DeleteRoom(roomID int)
    Join(roomID int, userID int, conn *websocket.Conn)
    RemoveClient(roomID int, userID int) error
    RoomExists(roomID int) bool
    SetTrack(roomID int, userID int, localTrack *webrtc.TrackLocalStaticRTP, track *webrtc.TrackRemote) error
    SetRemoteDescription(roomID int, userID int, sdp string) error
    AddICECandidate(roomID, userID int, cand webrtc.ICECandidateInit) error
}
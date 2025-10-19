package join_room_usecase

import (
	"github.com/gorilla/websocket"
	live_video_hub "streaming-server.com/application/ports/realtime/hubs"
	live_video_dto "streaming-server.com/application/usecases/live_video/dto"
	"streaming-server.com/infrastructure/logger"
)

var log = logger.Log

type JoinRoomUsecase struct {
	roomRepository live_video_hub.Interface
}

func NewJoinRoom(
	roomRepo live_video_hub.Interface,
) *JoinRoomUsecase {
	return &JoinRoomUsecase{
		roomRepo,
	}
}

func (u *JoinRoomUsecase) Do(
	params *live_video_dto.Params,
	conn *websocket.Conn,
) error {
	// room の認証が必要な場合ここ
	u.roomRepository.Join(params.RoomID, params.UserID, conn)
	// TODO どこかで共通化
	msg := struct {
		Type string `json:"type"`
		Data struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		} `json:"data"`
	}{
		Type: "join",
		Data: struct {
			RoomID      int    `json:"roomId"`
			PublisherID string `json:"publisherId"`
		}{
			RoomID:      params.RoomID,
			PublisherID: "publisherID",
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Error("WriteJSON error: ", err)
		return err
	}
	log.Info("👌 Send: Join")
	return nil
}
// [INFO] 2025/10/19 18:29:02 [websocket_router.go:117] [router.go] Registered route: /live/{roomId}/{userId}
// [INFO] 2025/10/19 18:29:02 [router.go:178] /dashboard
// [INFO] 2025/10/19 18:29:02 [router.go:178] /settings
// [INFO] 2025/10/19 18:29:02 [router.go:178] /hello
// [INFO] 2025/10/19 18:29:02 [router.go:178] /api/echo
// [INFO] 2025/10/19 18:29:02 [router.go:178] /api/item
// [INFO] 2025/10/19 18:29:02 [router.go:178] /ws/live/{roomId}/{userId}
// [INFO] 2025/10/19 18:29:02 [router.go:178] /static/
// [INFO] 2025/10/19 18:29:02 [router.go:178] /auth/broadcast
// [INFO] 2025/10/19 18:29:02 [router.go:178] /
// [INFO] 2025/10/19 18:29:02 [main.go:46] ✅ Server listening on :8080
// [INFO] 2025/10/19 18:29:09 [logger_middleware.go:45] 📥 Request: GET /ws/live/1/2
// 2025/10/19 18:29:09 map[roomId:1 userId:2]
// [INFO] 2025/10/19 18:29:09 [auth_middleware.go:9] Auth
// [INFO] 2025/10/19 18:29:09 [join_room_usecase.go:57] 👌 Send: Join
// [INFO] 2025/10/19 18:29:09 [get_offer_usecase.go:60] 📥 Request: ws/live OnICECandidate
// [INFO] 2025/10/19 18:29:09 [get_offer_usecase.go:72] 👌 Send: Candidate
// [INFO] 2025/10/19 18:29:09 [get_offer_usecase.go:60] 📥 Request: ws/live OnICECandidate
// [INFO] 2025/10/19 18:29:09 [get_offer_usecase.go:72] 👌 Send: Candidate
// [INFO] 2025/10/19 18:29:09 [get_offer_usecase.go:60] 📥 Request: ws/live OnICECandidate
// [INFO] 2025/10/19 18:29:09 [get_offer_usecase.go:115] 👌 Send: Answer
// [INFO] 2025/10/19 18:29:09 [set_candidate_usecase.go:56] 👌 Send: Set Candidate
// [INFO] 2025/10/19 18:29:09 [set_candidate_usecase.go:56] 👌 Send: Set Candidate
// [DEBUG] 2025/10/19 18:29:10 [get_offer_usecase.go:39] 📡 Track received from publisher:%!(EXTRA string=audio)
// [DEBUG] 2025/10/19 18:29:10 [get_offer_usecase.go:39] 📡 Track received from publisher:%!(EXTRA string=video)
// [INFO] 2025/10/19 18:29:13 [logger_middleware.go:45] 📥 Request: GET /
// 2025/10/19 18:29:13 map[]
// [INFO] 2025/10/19 18:29:13 [logger_middleware.go:60] 📤 Not Modified GET / (271.458µs)
// [INFO] 2025/10/19 18:29:14 [logger_middleware.go:45] 📥 Request: GET /ws/live/1/1
// 2025/10/19 18:29:14 map[roomId:1 userId:1]
// [INFO] 2025/10/19 18:29:14 [auth_middleware.go:9] Auth
// [INFO] 2025/10/19 18:29:14 [join_room_usecase.go:57] 👌 Send: Join
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:64] 👌 Send: Candidate to Viewer
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:64] 👌 Send: Candidate to Viewer
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:51] ICE gathering complete
// [INFO] 2025/10/19 18:29:14 [create_viewer_peer_connection_usecase.go:98] 👌 Send: Offer To Viewer
// [DEBUG] 2025/10/19 18:29:14 [set_answer_usecase.go:52] send answered
// [INFO] 2025/10/19 18:29:14 [set_candidate_usecase.go:56] 👌 Send: Set Candidate
// [INFO] 2025/10/19 18:29:22 [logger_middleware.go:45] 📥 Request: GET /
// 2025/10/19 18:29:22 map[]
// [INFO] 2025/10/19 18:29:22 [logger_middleware.go:60] 📤 Not Modified GET / (1.394917ms)
// [WARN] 2025/10/19 18:29:22 [websocket_router.go:89] 🔌 WS read error: websocket: close 1001 (going away)
// [INFO] 2025/10/19 18:29:22 [websocket_router.go:78] 🔌 Connection closed for map[roomId:1 userId:1]
// [DEBUG] 2025/10/19 18:29:22 [close_connection_usecase.go:28] 🧩 RemoveClient called: room=1 user=1
// [INFO] 2025/10/19 18:29:22 [room_hub.go:108] 🧹 Removed client: 1
// [INFO] 2025/10/19 18:29:22 [logger_middleware.go:45] 📥 Request: GET /ws/live/1/1
// 2025/10/19 18:29:22 map[roomId:1 userId:1]
// [INFO] 2025/10/19 18:29:22 [auth_middleware.go:9] Auth
// [INFO] 2025/10/19 18:29:22 [join_room_usecase.go:57] 👌 Send: Join
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:64] 👌 Send: Candidate to Viewer
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:64] 👌 Send: Candidate to Viewer
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:51] ICE gathering complete
// [INFO] 2025/10/19 18:29:22 [create_viewer_peer_connection_usecase.go:98] 👌 Send: Offer To Viewer
// [DEBUG] 2025/10/19 18:29:22 [set_answer_usecase.go:52] send answered
// [INFO] 2025/10/19 18:29:22 [set_candidate_usecase.go:56] 👌 Send: Set Candidate
// [INFO] 2025/10/19 18:29:27 [logger_middleware.go:45] 📥 Request: GET /
// 2025/10/19 18:29:27 map[]
// [INFO] 2025/10/19 18:29:27 [logger_middleware.go:60] 📤 Not Modified GET / (209.292µs)
// [WARN] 2025/10/19 18:29:27 [websocket_router.go:89] 🔌 WS read error: websocket: close 1001 (going away)
// [INFO] 2025/10/19 18:29:27 [websocket_router.go:78] 🔌 Connection closed for map[roomId:1 userId:1]
// [DEBUG] 2025/10/19 18:29:27 [close_connection_usecase.go:28] 🧩 RemoveClient called: room=1 user=1
// [INFO] 2025/10/19 18:29:27 [room_hub.go:108] 🧹 Removed client: 1
// [INFO] 2025/10/19 18:29:28 [logger_middleware.go:45] 📥 Request: GET /ws/live/1/1
// 2025/10/19 18:29:28 map[roomId:1 userId:1]
// [INFO] 2025/10/19 18:29:28 [auth_middleware.go:9] Auth
// [INFO] 2025/10/19 18:29:28 [join_room_usecase.go:57] 👌 Send: Join
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:64] 👌 Send: Candidate to Viewer
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:64] 👌 Send: Candidate to Viewer
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:49] 📥 Request: OnICECandidate
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:51] ICE gathering complete
// [INFO] 2025/10/19 18:29:28 [create_viewer_peer_connection_usecase.go:98] 👌 Send: Offer To Viewer
// [DEBUG] 2025/10/19 18:29:28 [set_answer_usecase.go:52] send answered
// [INFO] 2025/10/19 18:29:28 [set_candidate_usecase.go:56] 👌 Send: Set Candidate
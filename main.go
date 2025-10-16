package main

import (
	// room_entity "streaming-server.com/domain/entities/room"
	// close_connection_usecase "streaming-server.com/application/usecases/live_video/close_connection"
	create_viewer_peer_connection_usecase "streaming-server.com/application/usecases/live_video/create_viewer_peer_connection"
	get_offer_usecase "streaming-server.com/application/usecases/live_video/get_offer"
	join_room_usecase "streaming-server.com/application/usecases/live_video/join_room"
	set_answer_usecase "streaming-server.com/application/usecases/live_video/set_answer"
	set_candidate_usecase "streaming-server.com/application/usecases/live_video/set_candidate"
	"streaming-server.com/infrastructure/logger"
	"streaming-server.com/infrastructure/repository/memory/rooms"
	"streaming-server.com/infrastructure/server"
	// broadcaster "streaming-server.com/infrastructure/ws"

	// "streaming-server.com/infrastructure/server"
	"streaming-server.com/interface/controllers"
	websocket_controller "streaming-server.com/interface/controllers/http/websocket"
	live_video_controller "streaming-server.com/interface/controllers/websocket/live_video"
	"streaming-server.com/interface/router"
)

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool { return true },
// }

// var rooms = struct {
//     sync.Mutex
//     m map[int]*room.Entity
// }{m: make(map[int]*room.Entity)}

// func getRoomByID(roomId int) *room.Entity {
//     rooms.Lock()
//     defer rooms.Unlock()
//     r, ok := rooms.m[roomId]
//     if !ok {
//         r = &room.Entity{}
//         rooms.m[roomId] = r
//     }
//     return r
// }
// var clients = sync.Map{} // conn → Client

// WebSocket handler
// func handleWS(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println("upgrade:", err)
// 		return
// 	}
// 	defer conn.Close()
//     client := &rtc_client.Entity{
//         Conn:   conn,
//         PC:   make(chan []byte, 256),
// 		IsPublisher: false,
//         RoomID: someRoomID, // リクエストから抽出するなど
//     }

//     // roomに追加
//     room := getOrCreateRoom(client.roomID)
//     room.addClient(client)
// 	// クリーンアップ
// 	defer func() {
// 		if c, ok := clients.Load(conn); ok {
// 		client := c.(*rtc_client.Entity)
// 		room := rooms.m[client.roomId]

// 		if room.removeClient(client) {
// 			rooms.Lock()
// 			delete(rooms.m, client.roomId)
// 			rooms.Unlock()
// 			log.Printf("Room %s removed", client.roomId)
// 		}

// 		client.pc.Close()
// 		client.conn.Close()
// 		}

// 	}()
// 	// vpc に対応する conn を探す（再交渉で個別送信するため）
// 	getConnByPC := func(target *webrtc.PeerConnection) *websocket.Conn {
// 		var found *websocket.Conn
// 		clients.Range(func(k, v any) bool {
// 			c := v.(*Client)
// 			if c.pc == target {
// 				found = c.conn
// 				return false
// 			}
// 			return true
// 		})
// 		return found
// 	}

// 	var msg struct {
// 		Type          string  `json:"type"`
// 		RoomID        int  `json:"roomId"`   // ←追加
// 		SDP           string  `json:"sdp"`
// 		Candidate     string  `json:"candidate"`
// 		SDPMid        *string `json:"sdpMid"`
// 		SDPMLineIndex *uint16 `json:"sdpMLineIndex"`
// 	}

// 	for {
// 		if err := conn.ReadJSON(&msg); err != nil {
// 			log.Println("read:", err)
// 			return
// 		}

// 		switch msg.Type {
// 		// ─────────────────────────────────────────────
// 		// Publisher（Android）からの Offer
// 		// ─────────────────────────────────────────────
// 		case "offer":
// 			pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
// 				ICEServers: []webrtc.ICEServer{
// 					{URLs: []string{"stun:stun.l.google.com:19302"}},
// 				},
// 			})
// 			if err != nil {
// 				log.Println("pc:", err)
// 				return
// 			}
// 			clients.Store(conn, &Client{conn: conn, pc: pc, isPublisher: true, roomId: msg.RoomID})
// 			room := getRoomByID(msg.RoomID)

// 			// 受信専用トランシーバ（保険）
// 			_, _ = pc.AddTransceiverFromKind(
// 				webrtc.RTPCodecTypeVideo,
// 				webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly},
// 			)
// 			_, _ = pc.AddTransceiverFromKind(
// 				webrtc.RTPCodecTypeAudio,
// 				webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly},
// 			)

// 			// Publisher から受信したリモートトラックを Viewer へ配る
// 			pc.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
// 				log.Println("📡 Track received from publisher:", track.Kind().String())

// 				// 1) TrackID は必ず Kind().String() を使う！
// 				localTrack, err := webrtc.NewTrackLocalStaticRTP(
// 					track.Codec().RTPCodecCapability,
// 					track.Kind().String(), // ← ここが超重要
// 					"pion",
// 				)
// 				if err != nil {
// 					log.Println("NewTrackLocalStaticRTP error:", err)
// 					return
// 				}

// 				// 2) room へ差し替え（append しない）
// 				room := getRoomByID(msg.RoomID)
// 				room.mu.Lock()
// 				if track.Kind() == webrtc.RTPCodecTypeVideo {
// 					room.video = localTrack
// 				} else if track.Kind() == webrtc.RTPCodecTypeAudio {
// 					room.audio = localTrack
// 				}
// 				viewers := append([]*webrtc.PeerConnection(nil), room.peers...) // スナップショット
// 				room.mu.Unlock()

// 				// 3) Publisher→LocalTrack へのパイプ
// 				go func() {
// 					buf := make([]byte, 1500)
// 					for {
// 						n, _, err := track.Read(buf)
// 						if err != nil {
// 							break
// 						}
// 						if _, err = localTrack.Write(buf[:n]); err != nil {
// 							break
// 						}
// 					}
// 				}()

// 				// 4) 既存 Viewer へ割り当て。ReplaceTrack 成功なら再交渉しない。
// 				for _, vpc := range viewers {
// 					if vpc == pc { // publisher 自身には送らない
// 						continue
// 					}

// 					// 4-1) ReplaceTrack を試す
// 					replaced := false
// 					for _, t := range vpc.GetTransceivers() {
// 						if t.Kind() == track.Kind() && t.Sender() != nil {
// 							if err := t.Sender().ReplaceTrack(localTrack); err == nil {
// 								replaced = true
// 							} else {
// 								log.Println("ReplaceTrack:", err)
// 							}
// 							break
// 						}
// 					}

// 					if replaced {
// 						// ReplaceTrack だけなら再交渉不要
// 						continue
// 					}

// 					// 4-2) Sender が無い場合は AddTrack → Stable の時だけ 1 回だけ再交渉
// 					if _, err := vpc.AddTrack(localTrack); err != nil {
// 						log.Println("AddTrack to viewer:", err)
// 						continue
// 					}

// 					if vpc.SignalingState() != webrtc.SignalingStateStable {
// 						log.Println("skip renegotiate (not stable)")
// 						continue
// 					}

// 					offer, err := vpc.CreateOffer(nil)
// 					if err != nil {
// 						log.Println("renegotiate CreateOffer:", err)
// 						continue
// 					}
// 					g := webrtc.GatheringCompletePromise(vpc)
// 					if err := vpc.SetLocalDescription(offer); err != nil {
// 						log.Println("renegotiate SetLocal:", err)
// 						continue
// 					}
// 					<-g

// 					if vconn := getConnByPC(vpc); vconn != nil {
// 						if err := vconn.WriteJSON(map[string]string{
// 							"type": "offer",
// 							"sdp":  offer.SDP,
// 						}); err != nil {
// 							log.Println("send renegotiate offer:", err)
// 						}
// 					}
// 				}
// 			})

// 			pc.OnICECandidate(func(c *webrtc.ICECandidate) {
// 				if c != nil {
// 					_ = conn.WriteJSON(c.ToJSON())
// 				}
// 			})

// 			offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: msg.SDP}
// 			if err := pc.SetRemoteDescription(offer); err != nil {
// 				log.Println("setRemote:", err)
// 				return
// 			}
// 			answer, err := pc.CreateAnswer(nil)
// 			if err != nil {
// 				log.Println("createAnswer:", err)
// 				return
// 			}
// 			g := webrtc.GatheringCompletePromise(pc)
// 			_ = pc.SetLocalDescription(answer)
// 			<-g

// 			// room に publisher / viewer 共通の peers として登録
// 			room.mu.Lock()
// 			room.peers = append(room.peers, pc)
// 			room.mu.Unlock()

// 			_ = conn.WriteJSON(map[string]string{"type": "answer", "sdp": answer.SDP})

// 		// ─────────────────────────────────────────────
// 		// Publisher/Viewer からの Answer（再交渉含む）
// 		// ─────────────────────────────────────────────
// 		case "answer":
// 			val, ok := clients.Load(conn)
// 			if !ok {
// 				log.Println("no pc for this conn")
// 				return
// 			}
// 			client := val.(*Client)
// 			if err := client.pc.SetRemoteDescription(webrtc.SessionDescription{
// 				Type: webrtc.SDPTypeAnswer,
// 				SDP:  msg.SDP,
// 			}); err != nil {
// 				log.Println("setRemote answer:", err)
// 			}

// 		// ─────────────────────────────────────────────
// 		// ICE candidate 中継
// 		// ─────────────────────────────────────────────
// 		case "candidate":
// 			val, ok := clients.Load(conn)
// 			if !ok {
// 				log.Println("no pc for this conn")
// 				break
// 			}
// 			client := val.(*Client)
// 			cand := webrtc.ICECandidateInit{
// 				Candidate:     msg.Candidate,
// 				SDPMid:        msg.SDPMid,
// 				SDPMLineIndex: msg.SDPMLineIndex,
// 			}
// 			if err := client.pc.AddICECandidate(cand); err != nil {
// 				log.Println("AddICECandidate:", err)
// 			}

// 		// ─────────────────────────────────────────────
// 		// Viewer の入室（先に待機可能）
// 		// ─────────────────────────────────────────────
// 		case "viewer":
// 			pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
// 				ICEServers: []webrtc.ICEServer{
// 					{URLs: []string{"stun:stun.l.google.com:19302"}},
// 				},
// 			})
// 			if err != nil {
// 				log.Println("pc:", err)
// 				return
// 			}
// 			clients.Store(conn, &Client{conn: conn, pc: pc, isPublisher: false, roomId: msg.RoomID})
// 			room := getRoomByID(msg.RoomID)

// 			// 受信専用の transceiver を先に用意（あとから track が来ても受けられる）
// 			_, _ = pc.AddTransceiverFromKind(
// 				webrtc.RTPCodecTypeVideo,
// 				webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly},
// 			)
// 			_, _ = pc.AddTransceiverFromKind(
// 				webrtc.RTPCodecTypeAudio,
// 				webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly},
// 			)

// 			pc.OnICECandidate(func(c *webrtc.ICECandidate) {
// 				if c != nil {
// 					_ = conn.WriteJSON(c.ToJSON())
// 				}
// 			})

// 			// 既に room にトラックがあれば初回 offer に含める
// 			room.mu.Lock()
// 			if room.video != nil {
// 				if _, err := pc.AddTrack(room.video); err != nil {
// 					log.Println("AddTrack(video) to viewer:", err)
// 				}
// 			}
// 			if room.audio != nil {
// 				if _, err := pc.AddTrack(room.audio); err != nil {
// 					log.Println("AddTrack(audio) to viewer:", err)
// 				}
// 			}
// 			room.peers = append(room.peers, pc) // viewer も peers に登録
// 			room.mu.Unlock()

// 			offer, err := pc.CreateOffer(nil)
// 			if err != nil {
// 				log.Println("createOffer viewer:", err)
// 				return
// 			}
// 			g := webrtc.GatheringCompletePromise(pc)
// 			if err := pc.SetLocalDescription(offer); err != nil {
// 				log.Println("setLocal viewer:", err)
// 				return
// 			}
// 			<-g

// 			if err := conn.WriteJSON(map[string]string{"type": "offer", "sdp": offer.SDP}); err != nil {
// 				log.Println("send offer to viewer:", err)
// 			}
// 		}
// 	}
// }

func main() {
	roomRepository := rooms_hub.New()
	joinRoomUsecase := join_room_usecase.NewJoinRoom(roomRepository)
	getOfferUsecase := get_offer_usecase.NewGetOffer(roomRepository)
	createViewerPeerConnectionUsecase := create_viewer_peer_connection_usecase.NewCreateViewerPeerConnection(roomRepository)
	setAnswerUsecase := set_answer_usecase.NewSetAnswer(roomRepository)
	setCandidateUsecase := set_candidate_usecase.NewSetCandidate(roomRepository)
	liveVideoController := live_video_controller.NewLiveVideoController(
		getOfferUsecase,
		joinRoomUsecase,
		createViewerPeerConnectionUsecase,
		setAnswerUsecase,
		setCandidateUsecase,
	)
	websocketController := websocket_controller.NewController(liveVideoController)
	controllers := controllers.NewControllers(
		liveVideoController,
		websocketController,
	)
	muxOrHandler := router.CreateHandler(controllers) // ← ここだけ変更（*ServeMux でなく http.Handler）
	srv := server.NewHTTPServer(muxOrHandler)

	logger.Log.Info("✅ Server listening on :8080")
	logger.Log.Error("%v", srv.ListenAndServe())
}
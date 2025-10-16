// infrastructure/ws/broadcaster.go
package broadcaster

// import (
// 	"context"

// 	"github.com/gorilla/websocket"
// 	notifier "streaming-server.com/application/ports/notifiers"
// 	"streaming-server.com/infrastructure/ws/message"
// )

// type Broadcaster struct {
//     // 例：接続管理
//     Clients map[int]map[int]*rtc_client_entity.RtcClient
//     VideoTracks   map[int]*webrtc.TrackLocalStaticRTP
//     AudioTracks   map[int]*webrtc.TrackLocalStaticRTP
//     MU      sync.Mutex
// }

// func NewBroadcaster() *Broadcaster {
//     return &Broadcaster{
//         users: map[int]*websocket.Conn{},
//         rooms: map[int]map[int]*websocket.Conn{},
//     }
// }

// var _ notifier.Interface = (*Broadcaster)(nil)

// func (b *Broadcaster) SendToUser(ctx context.Context, userID int, typ string, data any) error {
//     conn := b.users[userID]
//     if conn == nil { return nil }
//     return b.send(conn, typ, data)
// }

// func (b *Broadcaster) BroadcastRoom(ctx context.Context, roomID int, typ string, data any) error {
//     conns := b.rooms[roomID]
//     if conns == nil { return nil }
//     for _, c := range conns {
//         _ = b.send(c, typ, data) // エラーは握りつぶしでもOK（要件次第）
//     }
//     return nil
// }

// // 👇 これが欲しかった send(type, data)
// func (b *Broadcaster) send(conn *websocket.Conn, typ string, data any) error {
//     env := message.Envelope{Type: typ, Data: data}
//     // WriteJSON でそのまま送ってOK
//     return conn.WriteJSON(env)

//     // もしくは Raw に組み立てたいなら：
//     // payload, _ := json.Marshal(env)
//     // return conn.WriteMessage(websocket.TextMessage, payload)
// }
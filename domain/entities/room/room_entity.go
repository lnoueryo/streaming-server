package room_entity

// import (
// 	"sync"

// 	"github.com/pion/webrtc/v4"
// 	"streaming-server.com/domain/entities/rtc_client"
// 	user_entity "streaming-server.com/domain/entities/user"
// )

// type Room struct {
//     ID      int
//     Users user_entity.User
//     Clients map[int]*rtc_client_entity.RtcClient
//     VideoTracks   map[int]*webrtc.TrackLocalStaticRTP
//     AudioTracks   map[int]*webrtc.TrackLocalStaticRTP
//     MU      sync.Mutex
// }

// func (r *Room) RemoveClient(userId int) {
//     r.MU.Lock()
//     defer r.MU.Unlock()
//     delete(r.Clients, userId)
// }

// func (r *Room) GetClient(userId int) *rtc_client_entity.RtcClient {
//     r.MU.Lock()
//     defer r.MU.Unlock()
//     client, ok := r.Clients[userId];if !ok {
//         return nil
//     }
//     return client
// }

// func (r *Room) CreateRtcClient(userId int) *rtc_client_entity.RtcClient {
//     r.MU.Lock()
//     defer r.MU.Unlock()
//     return &rtc_client_entity.RtcClient{
//         UserID: userId,
//     }
// }

// func (r *Room) AddRtcClient(client *rtc_client_entity.RtcClient) {
//     r.MU.Lock()
//     defer r.MU.Unlock()

//     if _, exists := r.Clients[client.UserID]; exists {
//         return
//     }
//     r.Clients[client.UserID] = client
// }

// func (r *Room) RemoveVideoTrack(userId int) {
//     r.MU.Lock()
//     defer r.MU.Unlock()
//     delete(r.VideoTracks, userId)
// }

// func (r *Room) RemoveAudioTrack(userId int) {
//     r.MU.Lock()
//     defer r.MU.Unlock()
//     delete(r.AudioTracks, userId)
// }

// func (r *Room) GetVideoTrack(userId int) *webrtc.TrackLocalStaticRTP {
//     r.MU.Lock()
//     defer r.MU.Unlock()

//     track, ok := r.VideoTracks[userId];if !ok {
//         return nil
//     }
//     return track
// }

// func (r *Room) GetAudioTrack(userId int) *webrtc.TrackLocalStaticRTP {
//     r.MU.Lock()
//     defer r.MU.Unlock()

//     track, ok := r.AudioTracks[userId];if !ok {
//         return nil
//     }
//     return track
// }

// func (r *Room) RemoveTracksFromViewers(track *webrtc.TrackLocalStaticRTP) {
//     r.MU.Lock()
//     defer r.MU.Unlock()
//     for _, viewer := range r.Clients {
//         for _, sender := range viewer.PeerConn.GetSenders() {
//             senderTrack := sender.Track()
//             if track == senderTrack {
//                 sender.ReplaceTrack(nil)
//             }
//         }
//     }
// }
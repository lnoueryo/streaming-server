package user_entity

import "streaming-server.com/infrastructure/webrtc/broadcast"

type RuntimeUser struct {
    ID int
    *broadcast.PeerClient
}

func NewRuntimeUser(id int) *RuntimeUser {
    return &RuntimeUser{
        id,
        &broadcast.PeerClient{
            nil,
            nil,
        },
    }
}
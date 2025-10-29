package room_entity

import (
	user_entity "streaming-server.com/domain/entities/user"
)

type Room struct {
    ID      int
    Users	 map[int]*user_entity.RuntimeUser
}

func NewRoom(id int) *Room {
	return &Room{
		id,
		make(map[int]*user_entity.RuntimeUser),
	}
}

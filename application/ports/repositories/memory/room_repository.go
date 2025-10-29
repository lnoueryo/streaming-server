package room_memory_repository

import (
    "streaming-server.com/domain/entities/room"
)

type IRoomRepository interface {
    GetOrCreate(int) *room_entity.RuntimeRoom
    GetRoom(int) (*room_entity.RuntimeRoom, error)
    DeleteRoom(int)
}
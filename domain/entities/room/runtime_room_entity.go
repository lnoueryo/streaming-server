package room_entity

import (
	"context"
	"errors"
	"sync"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	user_entity "streaming-server.com/domain/entities/user"
)

type RuntimeRoom struct {
	ID int
	listLock sync.RWMutex
	Users map[int]*user_entity.RuntimeUser
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
	cancelFunc context.CancelFunc
}


func NewRuntimeRoom(id int) *RuntimeRoom {
	return &RuntimeRoom{
		id,
		sync.RWMutex{},
		make(map[int]*user_entity.RuntimeUser),
		make(map[string]*webrtc.TrackLocalStaticRTP),
		nil,
	}
}

func (r *RuntimeRoom) getClient(userId int) (*user_entity.RuntimeUser, error) {
	client, ok := r.Users[userId];if !ok {
		return nil, errors.New("client not found")
	}
	return client, nil
}

func (r *RuntimeRoom) AddCancelFunc(cancelFunc context.CancelFunc) {
	r.cancelFunc = cancelFunc
}

// dispatchKeyFrame sends a keyframe to all PeerConnections, used everytime a new user joins the call.
func (r *RuntimeRoom) DispatchKeyFrame() {
	r.listLock.Lock()
	defer r.listLock.Unlock()

	for i := range r.Users {
		for _, receiver := range r.Users[i].Peer.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = r.Users[i].Peer.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
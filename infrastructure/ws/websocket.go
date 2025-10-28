package ws

import (
	"sync"
	"github.com/gorilla/websocket"
)

type ThreadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  interface{} `json:"data"`
}

func (t *ThreadSafeWriter) Send(event string, data interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(WebsocketMessage{event, data})
}
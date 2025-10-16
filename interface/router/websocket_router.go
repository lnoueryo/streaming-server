package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type WSMsgHandler func(conn *websocket.Conn, raw interface{})

type WSRoute struct {
	handlers map[string]WSMsgHandler
}

type compiledPath struct {
    any     http.Handler
    methods map[string]http.Handler
}

// イベント登録
func (w *WSRoute) On(event string, h WSMsgHandler) {
	if w.handlers == nil {
		w.handlers = make(map[string]WSMsgHandler)
	}
	w.handlers[event] = h
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (r *Router) WS(path string, setup func(ws *WSRoute)) {
	wsr := &WSRoute{handlers: make(map[string]WSMsgHandler)}
	setup(wsr)

	// Upgrade + メッセージディスパッチ
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Error("websocket upgrade error: %v", err)
			http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
			return
		}
		defer conn.Close()
        wsr.attachCloseHandler(conn)

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				// close/err
				break
			}
			var env struct {
				Type string			`json:"type"`
				Data interface{}	`json:"data"`
			}
			if err := json.Unmarshal(data, &env); err != nil || env.Type == "" {
				log.Error("ws: invalid message (no type)")
				continue
			}
			log.Info("📥 Request: ws%v %v", path, env.Type)
			h, ok := wsr.handlers[env.Type]
			if !ok {
				log.Error("ws: no handler for type=%s", env.Type)
				continue
			}
			// ハンドラには message 全体を渡す（Data を使いたければ各自でパース）
			h(conn, env.Data)
		}
	})

	r.addRoute(http.MethodGet, path, handler) // WS ハンドシェイクは GET
}

func (r *Router) addRoute(method, path string, h http.Handler) {
	full := joinPath(r.prefix, path)
	r.routes = append(r.routes, Route{Path: full, Method: method, Handler: h})
}

// helpers
func joinPath(a, b string) string {
	if a == "" || a == "/" {
		if b == "" {
			return "/"
		}
		if strings.HasPrefix(b, "/") {
			return b
		}
		return "/" + b
	}
	if b == "" || b == "/" {
		return a
	}
	return strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/")
}

func (wsr *WSRoute) attachCloseHandler(conn *websocket.Conn) {
    conn.SetCloseHandler(func(code int, text string) error {
        payload, _ := json.Marshal(struct {
            Code int    `json:"code"`
            Text string `json:"text"`
        }{
            Code: code,
            Text: text,
        })
		h, ok := wsr.handlers["close"]
		if ok {
			h(conn, payload)
		}
		log.Info("close websocket connection code: %v", code)
        return nil
    })
}
package middleware

import (
	"net/http"
	"strings"
	"time"

	"streaming-server.com/infrastructure/logger"
)

var log = logger.Log

// ヘッダだけで WebSocket アップグレードか判定（Gorilla依存なし）
func isWebSocketUpgrade(r *http.Request) bool {
	// 参考: RFC6455 Handshake
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// ステータス/サイズを計測するラッパ
type respCapture struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *respCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *respCapture) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1) リクエストログ（共通）
		start := time.Now()
		log.Info("📥 Request: " + r.Method + " " + r.URL.Path)

		// 2) WebSocket アップグレードは後段ログをスキップ
		if isWebSocketUpgrade(r) {
			// ここで返さず next を必ず呼ぶ（ハンドシェイク/アップグレードは必要）
			next.ServeHTTP(w, r)
			// 応答ログは出さずに終了
			return
		}

		// 3) ふつうの HTTP はステータス/サイズ/所要時間を出す
		rw := &respCapture{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		d := time.Since(start)
		log.Info(
			// 例: 📤 200 GET /items (512B, 23.4ms)
			"📤 " + http.StatusText(rw.status) +
				" " + r.Method + " " + r.URL.Path +
				" (" + d.String() + ")",
		)
	})
}
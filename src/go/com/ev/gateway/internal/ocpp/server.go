package ocpp

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	Upgrader websocket.Upgrader
	ConnMgr  *ConnManager
	Router   *Router
	Auth     Authenticator
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Expect /ocpp16/{chargePointId}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	cp := parts[len(parts)-1]

	if s.Auth != nil {
		if err := s.Auth.Validate(r.Context(), cp, r); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	s.ConnMgr.Set(cp, conn)
	log.Printf("[OCPP] connected cp=%s remote=%s", cp, r.RemoteAddr)

	readLoop(r.Context(), cp, conn, s.Router, s.ConnMgr)

	s.ConnMgr.Delete(cp)
	log.Printf("[OCPP] disconnected cp=%s", cp)
}

func readLoop(ctx context.Context, cp string, conn *websocket.Conn, router *Router, mgr *ConnManager) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		msgType, uid, action, payload, err := ParseFrame(msg)
		if err != nil {
			continue
		}

		switch msgType {
		case MsgCall:
			respPayload, ok, err := router.Dispatch(ctx, action, cp, uid, payload)
			if err != nil || !ok {
				// MVP: ignore unknown actions or handler errors
				continue
			}
			frame, err := BuildCallResult(uid, respPayload)
			if err != nil {
				continue
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_ = conn.WriteMessage(websocket.TextMessage, frame)

		case MsgCallResult:
			_ = mgr.ResolvePending(cp, uid, payload)

		case MsgCallError:
			// MVP: ignore. Later: resolve pending with error object
			var _ json.RawMessage
		}
	}
}

func WithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

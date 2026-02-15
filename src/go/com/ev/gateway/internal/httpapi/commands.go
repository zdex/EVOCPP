package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"ocpp-gateway/internal/events"
	"ocpp-gateway/internal/ocpp"
)

type CommandServer struct {
	ConnMgr   *ocpp.ConnManager
	Publisher events.Publisher
}

func NewCommandServer(mgr *ocpp.ConnManager, pub events.Publisher) *CommandServer {
	return &CommandServer{ConnMgr: mgr, Publisher: pub}
}

type Command struct {
	Type           string         `json:"type"`
	ChargePointId  string         `json:"chargePointId"`
	Payload        map[string]any `json:"payload"`
	IdempotencyKey string         `json:"idempotencyKey"`
}

func (s *CommandServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var cmd Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	var action string
	var payload any

	switch cmd.Type {
	case "RemoteStartTransaction":
		action = "RemoteStartTransaction"
		payload = cmd.Payload
	case "RemoteStopTransaction":
		action = "RemoteStopTransaction"
		payload = cmd.Payload
	case "ChangeConfiguration":
		action = "ChangeConfiguration"
		payload = cmd.Payload
	case "Reset":
		action = "Reset"
		payload = cmd.Payload
	default:
		http.Error(w, "unknown command type", http.StatusBadRequest)
		return
	}

	resp, uniqueId, err := s.ConnMgr.Call(ctx, cmd.ChargePointId, action, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Optional: emit a generic ack event (helps operations)
	_ = s.Publisher.Publish(r.Context(), map[string]any{
		"type":          "CommandAck",
		"chargePointId": cmd.ChargePointId,
		"commandType":   cmd.Type,
		"uniqueId":      uniqueId,
		"ts":            time.Now().UTC().Format(time.RFC3339),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"commandId": uniqueId,
		"response":  json.RawMessage(resp),
	})
}

package handlers

import (
	"context"
	"encoding/json"
	"time"

	"ocpp-gateway/internal/events"
)

type HeartbeatResp struct {
	CurrentTime string `json:"currentTime"`
}

type HeartbeatDeps struct {
	Publish func(ctx context.Context, evt any) error
}

func Heartbeat(d HeartbeatDeps) func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
	return func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
		now := time.Now().UTC().Format(time.RFC3339)
		if d.Publish != nil {
			_ = d.Publish(ctx, events.ChargerHeartbeat{
				Type:          "ChargerHeartbeat",
				ChargePointId: cp,
				Ts:            now,
			})
		}
		return HeartbeatResp{CurrentTime: now}, nil
	}
}

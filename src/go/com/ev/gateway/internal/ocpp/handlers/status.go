package handlers

import (
	"context"
	"encoding/json"
	"time"

	"ocpp-gateway/internal/events"
)

type StatusNotificationReq struct {
	ConnectorId int    `json:"connectorId"`
	Status      string `json:"status"`
	ErrorCode   string `json:"errorCode"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type StatusNotificationResp struct{}

type StatusDeps struct {
	Publish func(ctx context.Context, evt any) error
}

func StatusNotification(d StatusDeps) func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
	return func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
		var req StatusNotificationReq
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, err
		}

		ts := req.Timestamp
		if ts == "" {
			ts = time.Now().UTC().Format(time.RFC3339)
		}

		if d.Publish != nil {
			_ = d.Publish(ctx, events.ConnectorStatusChanged{
				Type:          "ConnectorStatusChanged",
				ChargePointId: cp,
				ConnectorId:   req.ConnectorId,
				Status:        req.Status,
				ErrorCode:     req.ErrorCode,
				Ts:            ts,
			})
		}

		return StatusNotificationResp{}, nil
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"time"

	"ocpp-gateway/internal/events"
)

type BootNotificationReq struct {
	ChargePointVendor string `json:"chargePointVendor"`
	ChargePointModel  string `json:"chargePointModel"`
	FirmwareVersion   string `json:"firmwareVersion,omitempty"`
}

type BootNotificationResp struct {
	Status      string `json:"status"`
	CurrentTime string `json:"currentTime"`
	Interval    int    `json:"interval"`
}

type BootDeps struct {
	Publish            func(ctx context.Context, evt any) error
	DefaultHBIntervalS int
}

func BootNotification(d BootDeps) func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
	return func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
		var req BootNotificationReq
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, err
		}

		now := time.Now().UTC().Format(time.RFC3339)

		if d.Publish != nil {
			_ = d.Publish(ctx, events.ChargerBooted{
				Type:            "ChargerBooted",
				ChargePointId:   cp,
				Vendor:          req.ChargePointVendor,
				Model:           req.ChargePointModel,
				FirmwareVersion: req.FirmwareVersion,
				OcppVersion:     "1.6J",
				Ts:              now,
			})
		}

		interval := d.DefaultHBIntervalS
		if interval <= 0 {
			interval = 300
		}

		return BootNotificationResp{
			Status:      "Accepted",
			CurrentTime: now,
			Interval:    interval,
		}, nil
	}
}

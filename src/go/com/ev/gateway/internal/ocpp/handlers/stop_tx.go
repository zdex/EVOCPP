package handlers

import (
	"context"
	"encoding/json"
	"time"

	"ocpp-gateway/internal/events"
)

type StopTransactionReq struct {
	TransactionId int    `json:"transactionId"`
	MeterStop     int64  `json:"meterStop,omitempty"`
	Timestamp     string `json:"timestamp,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type StopTransactionResp struct {
	IdTagInfo struct {
		Status string `json:"status"`
	} `json:"idTagInfo"`
}

type StopTxDeps struct {
	Publish func(ctx context.Context, evt any) error
}

func StopTransaction(d StopTxDeps) func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
	return func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
		var req StopTransactionReq
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, err
		}

		ts := req.Timestamp
		if ts == "" {
			ts = time.Now().UTC().Format(time.RFC3339)
		}

		if d.Publish != nil {
			_ = d.Publish(ctx, events.TransactionEnded{
				Type:          "TransactionEnded",
				ChargePointId: cp,
				TransactionId: req.TransactionId,
				MeterStopWh:   req.MeterStop,
				Reason:        req.Reason,
				Ts:            ts,
			})
		}

		var resp StopTransactionResp
		resp.IdTagInfo.Status = "Accepted"
		return resp, nil
	}
}

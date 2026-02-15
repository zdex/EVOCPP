package handlers

import (
	"context"
	"encoding/json"
	"time"

	"ocpp-gateway/internal/events"
)

type StartTransactionReq struct {
	ConnectorId int    `json:"connectorId"`
	IdTag       string `json:"idTag,omitempty"`
	MeterStart  int64  `json:"meterStart,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type StartTransactionResp struct {
	TransactionId int `json:"transactionId"`
	IdTagInfo     struct {
		Status string `json:"status"` // Accepted/Blocked/Expired/Invalid/ConcurrentTx
	} `json:"idTagInfo"`
}

type StartTxDeps struct {
	Publish func(ctx context.Context, evt any) error
}

func StartTransaction(d StartTxDeps) func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
	return func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
		var req StartTransactionReq
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, err
		}

		// MVP: generate a pseudo transaction id based on time
		txID := int(time.Now().UTC().Unix() % 100000000)

		ts := req.Timestamp
		if ts == "" {
			ts = time.Now().UTC().Format(time.RFC3339)
		}

		if d.Publish != nil {
			_ = d.Publish(ctx, events.TransactionStarted{
				Type:          "TransactionStarted",
				ChargePointId: cp,
				ConnectorId:   req.ConnectorId,
				TransactionId: txID,
				IdTag:         req.IdTag,
				MeterStartWh:  req.MeterStart,
				Ts:            ts,
			})
		}

		resp := StartTransactionResp{TransactionId: txID}
		resp.IdTagInfo.Status = "Accepted"
		return resp, nil
	}
}

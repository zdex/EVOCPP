package handlers

import (
	"context"
	"encoding/json"
	"time"

	"ocpp-gateway/internal/events"
)

type MeterValuesReq struct {
	ConnectorId   int `json:"connectorId"`
	TransactionId int `json:"transactionId,omitempty"`
	MeterValue    []struct {
		Timestamp    string `json:"timestamp"`
		SampledValue []struct {
			Value     string `json:"value"`
			Measurand string `json:"measurand,omitempty"`
			Unit      string `json:"unit,omitempty"`
		} `json:"sampledValue"`
	} `json:"meterValue"`
}

type MeterValuesResp struct{}

type MeterDeps struct {
	Publish func(ctx context.Context, evt any) error
}

func MeterValues(d MeterDeps) func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
	return func(ctx context.Context, cp, uid string, payload json.RawMessage) (any, error) {
		var req MeterValuesReq
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, err
		}

		if d.Publish != nil {
			for _, mv := range req.MeterValue {
				ts := mv.Timestamp
				if ts == "" {
					ts = time.Now().UTC().Format(time.RFC3339)
				}
				samples := make([]events.MeterValue, 0, len(mv.SampledValue))
				for _, sv := range mv.SampledValue {
					meas := sv.Measurand
					if meas == "" {
						meas = "Energy.Active.Import.Register"
					}
					unit := sv.Unit
					if unit == "" {
						unit = "Wh"
					}
					// Keep value as string to avoid parsing errors; CPMS can normalize.
					samples = append(samples, events.MeterValue{
						Measurand: meas,
						Unit:      unit,
						Value:     sv.Value,
					})
				}

				_ = d.Publish(ctx, events.MeterSample{
					Type:          "MeterSample",
					ChargePointId: cp,
					TransactionId: req.TransactionId,
					Ts:            ts,
					Samples:       samples,
				})
			}
		}

		return MeterValuesResp{}, nil
	}
}

package ocpp

import (
	"context"
	"encoding/json"
)

type HandlerFunc func(ctx context.Context, chargePointId, uniqueId string, payload json.RawMessage) (any, error)

type Router struct {
	handlers map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{handlers: map[string]HandlerFunc{}}
}

func (r *Router) Handle(action string, h HandlerFunc) {
	r.handlers[action] = h
}

func (r *Router) Dispatch(ctx context.Context, action, cp, uniqueId string, payload json.RawMessage) (any, bool, error) {
	h, ok := r.handlers[action]
	if !ok {
		return nil, false, nil
	}
	resp, err := h(ctx, cp, uniqueId, payload)
	return resp, true, err
}

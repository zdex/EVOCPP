package ocpp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type PendingCall struct {
	Ch    chan json.RawMessage
	Added time.Time
}

type ConnManager struct {
	mu      sync.RWMutex
	conns   map[string]*websocket.Conn         // chargePointId -> conn
	pending map[string]map[string]*PendingCall // chargePointId -> uniqueId -> pending
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		conns:   map[string]*websocket.Conn{},
		pending: map[string]map[string]*PendingCall{},
	}
}

func (m *ConnManager) Set(cp string, c *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[cp] = c
	if _, ok := m.pending[cp]; !ok {
		m.pending[cp] = map[string]*PendingCall{}
	}
}

func (m *ConnManager) Delete(cp string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, cp)
	delete(m.pending, cp)
}

func (m *ConnManager) Get(cp string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.conns[cp]
	return c, ok
}

func (m *ConnManager) RegisterPending(cp, uniqueId string) *PendingCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.pending[cp]; !ok {
		m.pending[cp] = map[string]*PendingCall{}
	}
	p := &PendingCall{Ch: make(chan json.RawMessage, 1), Added: time.Now().UTC()}
	m.pending[cp][uniqueId] = p
	return p
}

func (m *ConnManager) ResolvePending(cp, uniqueId string, payload json.RawMessage) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	pm, ok := m.pending[cp]
	if !ok {
		return false
	}
	p, ok := pm[uniqueId]
	if !ok {
		return false
	}
	delete(pm, uniqueId)
	select {
	case p.Ch <- payload:
	default:
	}
	close(p.Ch)
	return true
}

func (m *ConnManager) Call(ctx context.Context, cp string, action string, payload any) (json.RawMessage, string, error) {
	conn, ok := m.Get(cp)
	if !ok {
		return nil, "", errors.New("charger not connected")
	}

	uniqueId := uuid.NewString()
	pending := m.RegisterPending(cp, uniqueId)

	frame, err := BuildCall(uniqueId, action, payload)
	if err != nil {
		return nil, "", err
	}

	// Write should be serialized per connection in a real system.
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		return nil, uniqueId, err
	}

	select {
	case <-ctx.Done():
		return nil, uniqueId, ctx.Err()
	case resp, ok := <-pending.Ch:
		if !ok {
			return nil, uniqueId, errors.New("no response")
		}
		return resp, uniqueId, nil
	}
}

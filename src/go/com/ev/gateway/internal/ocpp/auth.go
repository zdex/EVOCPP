package ocpp

import (
	"context"
	"errors"
	"net/http"

	"ocpp-gateway/internal/cpms"
)

type Authenticator interface {
	Validate(ctx context.Context, chargePointId string, r *http.Request) error
}

type AllowAllAuthenticator struct{}

func (AllowAllAuthenticator) Validate(_ context.Context, _ string, _ *http.Request) error { return nil }

type CPMSAuthenticator struct {
	Client *cpms.Client
}

func NewCPMSAuthenticator(c *cpms.Client) *CPMSAuthenticator {
	return &CPMSAuthenticator{Client: c}
}

func (a *CPMSAuthenticator) Validate(ctx context.Context, chargePointId string, r *http.Request) error {
	secret := r.URL.Query().Get("secret")
	if secret == "" {
		// allow Basic Auth as alternative
		if u, p, ok := r.BasicAuth(); ok {
			_ = u
			secret = p
		}
	}
	if secret == "" {
		return errors.New("missing secret")
	}

	allowed, err := a.Client.ValidateCharger(ctx, chargePointId, secret, r.RemoteAddr)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("unauthorized")
	}
	return nil
}

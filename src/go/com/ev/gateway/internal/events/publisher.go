package events

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, evt any) error
}

type HTTPPublisher struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func NewHTTPPublisher(baseURL, apiKey string) *HTTPPublisher {
	return &HTTPPublisher{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (p *HTTPPublisher) Publish(ctx context.Context, evt any) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/v1/gateway/events", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// For MVP: accept any 2xx/202
	if resp.StatusCode/100 != 2 {
		return &httpStatusError{Code: resp.StatusCode}
	}
	return nil
}

type httpStatusError struct{ Code int }

func (e *httpStatusError) Error() string { return "cpms returned non-2xx" }

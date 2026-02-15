package cpms

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 8 * time.Second},
	}
}

type ValidateReq struct {
	PresentedSecret string `json:"presentedSecret"`
	RemoteAddr      string `json:"remoteAddr,omitempty"`
	CertFingerprint string `json:"certFingerprint,omitempty"`
}
type ValidateResp struct {
	Allowed bool `json:"allowed"`
}

func (c *Client) ValidateCharger(ctx context.Context, chargePointId, presentedSecret, remoteAddr string) (bool, error) {
	reqBody, _ := json.Marshal(ValidateReq{PresentedSecret: presentedSecret, RemoteAddr: remoteAddr})
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/gateway/chargers/"+chargePointId+"/auth", bytes.NewReader(reqBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}

	var vr ValidateResp
	_ = json.NewDecoder(resp.Body).Decode(&vr)
	return vr.Allowed, nil
}

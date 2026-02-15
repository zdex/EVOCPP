package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AuthReq struct {
	PresentedSecret string `json:"presentedSecret"`
	RemoteAddr      string `json:"remoteAddr,omitempty"`
	CertFingerprint string `json:"certFingerprint,omitempty"`
}
type AuthResp struct {
	Allowed     bool   `json:"allowed"`
	OcppVersion string `json:"ocppVersion,omitempty"`
}

func main() {
	addr := getenv("CPMS_MOCK_ADDR", ":8081")
	allowSecret := getenv("CPMS_MOCK_ALLOW_SECRET", "devsecret")

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/gateway/chargers/", func(w http.ResponseWriter, r *http.Request) {
		// POST /v1/gateway/chargers/{id}/auth
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/auth") {
			http.NotFound(w, r)
			return
		}

		var req AuthReq
		_ = json.NewDecoder(r.Body).Decode(&req)

		resp := AuthResp{
			Allowed:     req.PresentedSecret == allowSecret,
			OcppVersion: "1.6J",
		}
		if !resp.Allowed {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v1/gateway/events", func(w http.ResponseWriter, r *http.Request) {
		// Accept events
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		log.Printf("[CPMS-MOCK] event received: %v", body["type"])
		w.WriteHeader(http.StatusAccepted)
	})

	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("CPMS mock listening on", addr)
	log.Fatal(s.ListenAndServe())
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

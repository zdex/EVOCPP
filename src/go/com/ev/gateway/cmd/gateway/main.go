package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"ocpp-gateway/internal/cpms"
	"ocpp-gateway/internal/events"
	"ocpp-gateway/internal/httpapi"
	"ocpp-gateway/internal/ocpp"
	"ocpp-gateway/internal/ocpp/handlers"

	"github.com/gorilla/websocket"
)

func main() {
	listen := getenv("LISTEN_ADDR", ":8080")
	cpmsURL := getenv("CPMS_BASE_URL", "http://localhost:8081")
	apiKey := getenv("CPMS_API_KEY", "dev")

	requireAuth := getenv("REQUIRE_CPMS_AUTH", "true") == "true"
	interval := mustInt(getenv("DEFAULT_HEARTBEAT_INTERVAL", "300"), 300)

	cpmsClient := cpms.NewClient(cpmsURL, apiKey)
	publisher := events.NewHTTPPublisher(cpmsURL, apiKey)

	connMgr := ocpp.NewConnManager()
	router := ocpp.NewRouter()

	// Register OCPP inbound handlers
	router.Handle("BootNotification", handlers.BootNotification(handlers.BootDeps{
		Publish:            publisher.Publish,
		DefaultHBIntervalS: interval,
	}))
	router.Handle("Heartbeat", handlers.Heartbeat(handlers.HeartbeatDeps{
		Publish: publisher.Publish,
	}))
	router.Handle("StatusNotification", handlers.StatusNotification(handlers.StatusDeps{
		Publish: publisher.Publish,
	}))
	router.Handle("StartTransaction", handlers.StartTransaction(handlers.StartTxDeps{
		Publish: publisher.Publish,
	}))
	router.Handle("MeterValues", handlers.MeterValues(handlers.MeterDeps{
		Publish: publisher.Publish,
	}))
	router.Handle("StopTransaction", handlers.StopTransaction(handlers.StopTxDeps{
		Publish: publisher.Publish,
	}))

	var auth ocpp.Authenticator
	if requireAuth {
		auth = ocpp.NewCPMSAuthenticator(cpmsClient)
	} else {
		auth = ocpp.AllowAllAuthenticator{}
	}

	ocppSrv := &ocpp.Server{
		Upgrader: websocket.Upgrader{
			CheckOrigin:  func(r *http.Request) bool { return true }, // tighten in prod
			Subprotocols: []string{"ocpp1.6"},
		},
		ConnMgr: connMgr,
		Router:  router,
		Auth:    auth,
	}

	// HTTP API for CPMS -> Gateway commands
	cmdAPI := httpapi.NewCommandServer(connMgr, publisher)

	mux := http.NewServeMux()
	mux.Handle("/ocpp16/", ocppSrv)
	mux.Handle("/v1/gateway/commands", cmdAPI)

	s := &http.Server{
		Addr:              listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("OCPP Gateway listening on", listen)
	log.Println("OCPP endpoint: ws(s)://<host>" + listen + "/ocpp16/{chargePointId}")
	log.Fatal(s.ListenAndServe())
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func mustInt(v string, d int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}

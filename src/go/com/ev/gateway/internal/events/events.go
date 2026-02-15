package events

type ChargerBooted struct {
	Type            string `json:"type"`
	ChargePointId   string `json:"chargePointId"`
	Vendor          string `json:"vendor"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmwareVersion,omitempty"`
	OcppVersion     string `json:"ocppVersion"`
	Ts              string `json:"ts"`
}

type ChargerHeartbeat struct {
	Type          string `json:"type"`
	ChargePointId string `json:"chargePointId"`
	Ts            string `json:"ts"`
}

type ConnectorStatusChanged struct {
	Type          string `json:"type"`
	ChargePointId string `json:"chargePointId"`
	ConnectorId   int    `json:"connectorId"`
	Status        string `json:"status"`
	ErrorCode     string `json:"errorCode"`
	Ts            string `json:"ts"`
}

type TransactionStarted struct {
	Type          string `json:"type"`
	ChargePointId string `json:"chargePointId"`
	ConnectorId   int    `json:"connectorId"`
	TransactionId int    `json:"transactionId"`
	IdTag         string `json:"idTag,omitempty"`
	MeterStartWh  int64  `json:"meterStartWh,omitempty"`
	Ts            string `json:"ts"`
}

type MeterSample struct {
	Type          string       `json:"type"`
	ChargePointId string       `json:"chargePointId"`
	TransactionId int          `json:"transactionId"`
	Ts            string       `json:"ts"`
	Samples       []MeterValue `json:"samples"`
}

type MeterValue struct {
	Measurand string `json:"measurand"`
	Unit      string `json:"unit"`
	Value     any    `json:"value"`
}

type TransactionEnded struct {
	Type          string `json:"type"`
	ChargePointId string `json:"chargePointId"`
	TransactionId int    `json:"transactionId"`
	MeterStopWh   int64  `json:"meterStopWh,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Ts            string `json:"ts"`
}

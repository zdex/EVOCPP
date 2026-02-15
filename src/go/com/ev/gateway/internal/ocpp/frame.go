package ocpp

import (
	"encoding/json"
	"errors"
)

const (
	MsgCall       = 2
	MsgCallResult = 3
	MsgCallError  = 4
)

func ParseFrame(raw []byte) (msgType int, uniqueId string, action string, payload json.RawMessage, err error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return 0, "", "", nil, err
	}
	if len(arr) < 3 {
		return 0, "", "", nil, errors.New("invalid frame length")
	}

	if err := json.Unmarshal(arr[0], &msgType); err != nil {
		return 0, "", "", nil, err
	}
	if err := json.Unmarshal(arr[1], &uniqueId); err != nil {
		return 0, "", "", nil, err
	}

	switch msgType {
	case MsgCall:
		if len(arr) < 4 {
			return 0, "", "", nil, errors.New("CALL frame requires 4 elements")
		}
		if err := json.Unmarshal(arr[2], &action); err != nil {
			return 0, "", "", nil, err
		}
		payload = arr[3]
		return msgType, uniqueId, action, payload, nil

	case MsgCallResult:
		payload = arr[2]
		return msgType, uniqueId, "", payload, nil

	case MsgCallError:
		// [4, uniqueId, errorCode, errorDescription, errorDetails]
		return msgType, uniqueId, "", nil, nil

	default:
		return 0, "", "", nil, errors.New("unknown message type")
	}
}

func BuildCall(uniqueId, action string, payload any) ([]byte, error) {
	pb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	frame := []any{MsgCall, uniqueId, action, json.RawMessage(pb)}
	return json.Marshal(frame)
}

func BuildCallResult(uniqueId string, payload any) ([]byte, error) {
	pb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	frame := []any{MsgCallResult, uniqueId, json.RawMessage(pb)}
	return json.Marshal(frame)
}

package messages

import (
	"encoding/json"
	"fmt"
	"time"
)

type Head struct {
	Destination   string    `json:"destination"`
	Time          time.Time `json:"time"`
	Correlationid string    `json:"correlation_id"`
	Eventtype     string    `json:"event_type"`
	Source        string    `json:"source"`
}

type Message[T any] struct {
	Head Head `json:"head"`
	Body T    `json:"body"`
}

type UnstricMessage struct {
	Message[any]
}

type WMessageBody struct {
	ClientID   uint   `json:"client_id"`
	CompanyID  uint   `json:"company_id"`
	InstanceID uint   `json:"instance_id"`
	Message    string `json:"message"`
}

type IncomingMessage struct {
	Message[WMessageBody]
}

type SendingMessage struct {
	Message[WMessageBody]
}

type HeadOnly struct {
	Head struct {
		Eventtype string `json:"event_type"`
	} `json:"head"`
}

func GetEventType(data []byte) (string, error) {
	var ho HeadOnly
	if err := json.Unmarshal(data, &ho); err != nil {
		return "", fmt.Errorf("failed to unmarshal head-only message: %w", err)
	}
	return ho.Head.Eventtype, nil
}

func Convert(raw []byte, out interface{}) error {
	return json.Unmarshal(raw, out)
}

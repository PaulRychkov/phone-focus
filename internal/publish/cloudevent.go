package publish

import (
	"encoding/json"
	"time"
)

const (
	SourcePhone = "phone"
	TypeUsage   = "phone.usage.snapshot"
)

type Envelope struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Subject         string          `json:"subject,omitempty"`
	Time            time.Time       `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

func NewEnvelope(id, source, eventType, subject string, occurredAt time.Time, data json.RawMessage) Envelope {
	return Envelope{
		SpecVersion:     "1.0",
		ID:              id,
		Source:          source,
		Type:            eventType,
		Subject:         subject,
		Time:            occurredAt,
		DataContentType: "application/json",
		Data:            data,
	}
}

func Marshal(id, subject string, occurredAt time.Time, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(NewEnvelope(id, SourcePhone, TypeUsage, subject, occurredAt, data))
}

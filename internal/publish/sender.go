package publish

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ingestPath = "/ingest/v1/phone-usage"

type Sender struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewSender(baseURL, token string) *Sender {
	return &Sender{
		baseURL: normalizeURL(baseURL),
		token:   strings.TrimSpace(token),
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func normalizeURL(raw string) string {
	base := strings.TrimSpace(raw)
	if i := strings.Index(base, "://"); i > 0 {
		base = strings.ToLower(base[:i]) + base[i:]
	}
	return strings.TrimRight(base, "/")
}

func (s *Sender) Configured() bool {
	return s.baseURL != ""
}

func (s *Sender) Send(payload []byte) error {
	if !s.Configured() {
		return fmt.Errorf("адрес сервера не задан")
	}
	req, err := http.NewRequest(http.MethodPost, s.baseURL+ingestPath, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("формирование запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("отправка: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("сервер ответил %s", resp.Status)
	}
	return nil
}

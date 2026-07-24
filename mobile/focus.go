package focus

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/PaulRychkov/phone-focus/internal/publish"
	"github.com/PaulRychkov/phone-focus/internal/state"
	"github.com/PaulRychkov/phone-focus/internal/usage"
)

var (
	mu        sync.Mutex
	baseURL   string
	token     string
	deviceID  string
	chatID    int64
	statePath string
	outbox           *publish.Outbox
	current          state.State
	summary          string
	lastSnapshotJSON string
)

func parseAppInfo(raw string) map[string]usage.AppMeta {
	if raw == "" {
		return nil
	}
	var decoded map[string]struct {
		Label    string `json:"label"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil
	}
	out := make(map[string]usage.AppMeta, len(decoded))
	for pkg, m := range decoded {
		out[pkg] = usage.AppMeta{Label: m.Label, Category: m.Category}
	}
	return out
}

func Configure(url, accessToken, device string, chat int64, dataDir string) error {
	mu.Lock()
	defer mu.Unlock()

	baseURL = url
	token = accessToken
	chatID = chat
	deviceID = device
	if deviceID == "" {
		deviceID = "phone"
	}

	box, err := publish.NewOutbox(filepath.Join(dataDir, "outbox"))
	if err != nil {
		return err
	}
	outbox = box
	statePath = filepath.Join(dataDir, "state.json")

	loaded, err := state.Load(statePath)
	if err != nil {
		return err
	}
	current = loaded
	return nil
}

func Push(rawEvents string, appInfoJSON string, windowSeconds int64, nowMillis int64) error {
	mu.Lock()
	defer mu.Unlock()

	if outbox == nil {
		return fmt.Errorf("библиотека не настроена")
	}

	now := time.Now()
	if nowMillis > 0 {
		now = time.UnixMilli(nowMillis)
	}
	window := time.Duration(windowSeconds) * time.Second
	if window <= 0 {
		window = 10 * time.Minute
	}

	result := usage.Build(usage.ParseEvents(rawEvents), usage.Params{
		DeviceID:      deviceID,
		ChatID:        chatID,
		WindowStart:   now.Add(-window),
		WindowEnd:     now,
		Seq:           current.Seq + 1,
		LastSignature: current.LastSignature,
		Location:      time.Local,
		AppInfo:       parseAppInfo(appInfoJSON),
	})

	snapshot := result.Snapshot
	current.Seq = snapshot.Seq
	current.LastSignature = result.Signature
	if err := state.Save(statePath, current); err != nil {
		return err
	}

	if data, err := json.Marshal(snapshot); err == nil {
		lastSnapshotJSON = string(data)
	}

	if snapshot.ChangedSinceLast || len(snapshot.Apps) > 0 {
		id := uuid.NewString()
		payload, err := publish.Marshal(id, deviceID, now, snapshot)
		if err != nil {
			return err
		}
		if err := outbox.Enqueue(snapshot.Seq, id, payload); err != nil {
			return err
		}
	}

	sender := publish.NewSender(baseURL, token)
	var sendErr error
	if sender.Configured() {
		_, sendErr = outbox.Flush(sender.Send)
	}

	summary = describe(snapshot, outbox.Size(), sender.Configured(), sendErr)
	return sendErr
}

func Summary() string {
	mu.Lock()
	defer mu.Unlock()
	if summary == "" {
		return "Замеров ещё не было"
	}
	return summary
}

func LastSnapshot() string {
	mu.Lock()
	defer mu.Unlock()
	return lastSnapshotJSON
}

func Pending() int {
	mu.Lock()
	defer mu.Unlock()
	if outbox == nil {
		return 0
	}
	return outbox.Size()
}

func describe(s usage.Snapshot, queued int, configured bool, sendErr error) string {
	today := s.Today.TotalForegroundSeconds
	lines := []string{
		fmt.Sprintf("Сейчас: %s", orDash(s.ForegroundApp)),
		fmt.Sprintf("За окно: приложений %d, отвлекающих %d мин", len(s.Apps), s.DistractingSeconds/60),
		fmt.Sprintf("Сегодня: %d ч %d мин, разблокировок %d", today/3600, (today%3600)/60, s.Today.Unlocks),
		fmt.Sprintf("В очереди: %d", queued),
	}
	switch {
	case !configured:
		lines = append(lines, "Адрес сервера не задан — данные копятся на телефоне")
	case sendErr != nil:
		lines = append(lines, "Ошибка отправки: "+sendErr.Error())
	default:
		lines = append(lines, "Отправлено")
	}
	return joinLines(lines)
}

func orDash(s string) string {
	if s == "" {
		return "нет активного приложения"
	}
	return s
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}

package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const maxOutboxFiles = 2000

type Outbox struct {
	dir string
	mu  sync.Mutex
}

func NewOutbox(dir string) (*Outbox, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("создание каталога очереди: %w", err)
	}
	return &Outbox{dir: dir}, nil
}

func (o *Outbox) Enqueue(seq int64, id string, payload []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	name := fmt.Sprintf("%020d-%s.json", seq, id)
	if err := os.WriteFile(filepath.Join(o.dir, name), payload, 0o644); err != nil {
		return fmt.Errorf("запись в очередь: %w", err)
	}
	o.trim()
	return nil
}

func (o *Outbox) Size() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.files())
}

func (o *Outbox) Flush(send func([]byte) error) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	sent := 0
	for _, name := range o.files() {
		path := filepath.Join(o.dir, name)
		payload, err := os.ReadFile(path)
		if err != nil {
			_ = os.Remove(path)
			continue
		}
		if err := send(payload); err != nil {
			return sent, err
		}
		_ = os.Remove(path)
		sent++
	}
	return sent, nil
}

func (o *Outbox) files() []string {
	entries, err := os.ReadDir(o.dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

func (o *Outbox) trim() {
	names := o.files()
	if len(names) <= maxOutboxFiles {
		return
	}
	for _, name := range names[:len(names)-maxOutboxFiles] {
		_ = os.Remove(filepath.Join(o.dir, name))
	}
}

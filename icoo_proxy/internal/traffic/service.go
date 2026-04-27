package traffic

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"icoo_proxy/internal/api"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	defaultLimit = 100
	keyPrefix    = "request:"
)

type Service struct {
	db *leveldb.DB
}

func NewService(root string) (*Service, error) {
	storeDir := filepath.Join(root, ".data", "traffic.leveldb")
	if err := os.MkdirAll(filepath.Dir(storeDir), 0o755); err != nil {
		return nil, err
	}
	db, err := leveldb.OpenFile(storeDir, &opt.Options{})
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

func (s *Service) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Service) RecordRequest(item api.RequestView) error {
	if s == nil || s.db == nil {
		return nil
	}
	if strings.TrimSpace(item.CreatedAt) == "" {
		item.CreatedAt = time.Now().Format(time.RFC3339)
	}
	if strings.TrimSpace(item.RequestID) == "" {
		item.RequestID = fmt.Sprintf("request-%d", time.Now().UnixNano())
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return s.db.Put([]byte(requestKey(item)), data, nil)
}

func (s *Service) ListRecent(limit int) []api.RequestView {
	if s == nil || s.db == nil {
		return nil
	}
	if limit <= 0 {
		limit = defaultLimit
	}

	iter := s.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)
	defer iter.Release()

	items := make([]api.RequestView, 0, limit)
	for iter.Next() {
		var item api.RequestView
		if err := json.Unmarshal(iter.Value(), &item); err != nil {
			continue
		}
		items = append(items, item)
		if len(items) >= limit {
			break
		}
	}
	return items
}

func requestKey(item api.RequestView) string {
	createdAt := parseCreatedAt(item.CreatedAt)
	reversed := math.MaxInt64 - createdAt.UnixNano()
	return fmt.Sprintf("%s%020d:%s", keyPrefix, reversed, item.RequestID)
}

func parseCreatedAt(value string) time.Time {
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	return time.Now()
}

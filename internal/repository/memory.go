package repository

import (
	"context"
	"sync"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type MemoryRepository struct {
	mu                  sync.RWMutex
	lastProcessedBlock  int64
	events              []model.Event // 記憶體中的事件清單, v1先不用D
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{
		events: make([]model.Event, 0),
	}
}

func (m *MemoryRepository) GetLastProcessedBlock(ctx context.Context) int64 {
	// RLock()：讀鎖（可多人同時讀）
	// Lock()：寫鎖（只能一個）
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastProcessedBlock
}

func (m *MemoryRepository) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > len(m.events) {
		limit = len(m.events)
	}

	return m.events[len(m.events)-limit:] // 回傳最後limit筆
}

package repository

import (
	"context"
	"sync"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type MemoryRepository struct {
	mu                  sync.RWMutex
	lastProcessedBlock  uint64
	lastBlockHash       string
	events              []model.Event // 記憶體中的事件清單, v1先不用D
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{
		events: make([]model.Event, 0),
	}
}

func (m *MemoryRepository) GetLastBlockHash(ctx context.Context) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastBlockHash, nil
}

func (m *MemoryRepository) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	// RLock()：讀鎖（可多人同時讀）
	// Lock()：寫鎖（只能一個）
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastProcessedBlock, nil
}

func (m *MemoryRepository) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > len(m.events) {
		limit = len(m.events)
	}

	return m.events[len(m.events)-limit:] // 回傳最後limit筆
}


// func (m *MemoryRepository) SetLastProcessedBlock(ctx context.Context, block uint64) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()

// 	m.lastProcessedBlock = block
// }

func (m *MemoryRepository) SetProcessed(ctx context.Context, block uint64, hash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastProcessedBlock = block
	m.lastBlockHash = hash
}

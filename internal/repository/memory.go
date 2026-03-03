package repository

import (
	"context"
	"sync"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type MemoryRepository struct {
	mu                 sync.RWMutex
	lastProcessedBlock uint64
	blockHashes        map[uint64]string
	events             []model.Event // 記憶體中的事件清單, v1先不用D
}

var _ Repository = (*MemoryRepository)(nil)

func NewMemoryRepository() Repository {
	return &MemoryRepository{
		blockHashes: make(map[uint64]string),
		events:      make([]model.Event, 0),
	}
}

func (m *MemoryRepository) GetBlockHash(ctx context.Context, number uint64) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash, ok := m.blockHashes[number]
	if !ok {
		return "", ErrBlockNotFound
	}

	return hash, nil
}

func (m *MemoryRepository) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	// RLock()：讀鎖（可多人同時讀）
	// Lock()：寫鎖（只能一個）
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastProcessedBlock, nil
}

func (m *MemoryRepository) GetMaxBlockNumberLE(ctx context.Context, target uint64) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	found := false
	var max uint64
	for n := range m.blockHashes {
		if n <= target && (!found || n > max) {
			max = n
			found = true
		}
	}
	if !found {
		return 0, ErrBlockNotFound
	}

	return max, nil
}

func (m *MemoryRepository) UpsertBlock(ctx context.Context, block uint64, hash string, parentHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastProcessedBlock = block
	m.blockHashes[block] = hash
	return nil
}

func (m *MemoryRepository) RollbackTo(ctx context.Context, block uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for n := range m.blockHashes {
		if n > block {
			delete(m.blockHashes, n)
		}
	}

	m.lastProcessedBlock = block
	return nil
}

func (m *MemoryRepository) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > len(m.events) {
		limit = len(m.events)
	}

	return m.events[len(m.events)-limit:] // 回傳最後limit筆
}

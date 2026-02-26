package service

import (
	"context"
	"log"
	"time"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
	"github.com/willie0x14/ethereum-block-scanner/internal/repository"
)

// struct tag, JSON encode 時欄位用這個名字     e.g. `json:"listener_running"`
type Status struct {
	ListenerRunning    bool  `json:"listener_running"`
	LastProcessedBlock uint64 `json:"last_processed_block"`
	UpdatedAt          int64 `json:"updated_at_unix"`
}

type ListenerService struct {
	repo repository.Repository
}

// constructor
func NewListenerService(repo repository.Repository) *ListenerService {
	return &ListenerService{
		repo: repo,
	}
}

func (s *ListenerService) GetStatus(ctx context.Context) Status {
	last, err := s.repo.GetLastProcessedBlock(ctx)
    if err != nil {
        log.Println("failed to get last processed:", err)
    }
	return Status{
		ListenerRunning:    true, // TODO
		LastProcessedBlock: last,
		UpdatedAt:          time.Now().Unix(),
	}
}

func (s *ListenerService) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	return s.repo.ListRecentEvents(ctx, limit)
}

func (s *ListenerService) ProcessBlock(ctx context.Context, block uint64, hash string) {
	log.Println("processing block:", block)

	s.repo.SetProcessed(ctx, block, hash)
}


func (s *ListenerService) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.repo.GetLastProcessedBlock(ctx)
}

func (s *ListenerService) MarkProcessed(ctx context.Context, height uint64, hash string) error {
	s.repo.SetProcessed(ctx, height, hash)
	return nil
}

func (s *ListenerService) GetLastBlockHash(ctx context.Context) (string, error) {
	return s.repo.GetLastBlockHash(ctx)
}

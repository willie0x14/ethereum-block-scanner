package service

import (
	"context"
	"time"
	"log"

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
	return Status{
		ListenerRunning:    true, // TODO
		LastProcessedBlock: s.repo.GetLastProcessedBlock(ctx),
		UpdatedAt:          time.Now().Unix(),
	}
}

func (s *ListenerService) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	return s.repo.ListRecentEvents(ctx, limit)
}

func (s *ListenerService) ProcessBlock(ctx context.Context, block uint64) {
	log.Println("processing block:", block)

	s.repo.SetLastProcessedBlock(ctx, block)
}


func (s *ListenerService) GetLastProcessedBlock(ctx context.Context) uint64 {
	return s.repo.GetLastProcessedBlock(ctx)
}


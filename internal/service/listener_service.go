package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
	"github.com/willie0x14/ethereum-block-scanner/internal/repository"
)

// struct tag, JSON encode 時欄位用這個名字     e.g. `json:"listener_running"`
type Status struct {
	ListenerRunning    bool   `json:"listener_running"`
	LastProcessedBlock uint64 `json:"last_processed_block"`
	UpdatedAt          int64  `json:"updated_at_unix"`
}

type ListenerService struct {
	repo repository.Repository
}

const bootstrapLookback uint64 = 5

func bootstrapStart(target uint64) uint64 {
	if target > bootstrapLookback {
		return target - bootstrapLookback
	}
	return 0
}

func (s *ListenerService) GetBlockHash(ctx context.Context, number uint64) (string, error) {
	return s.repo.GetBlockHash(ctx, number)
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
		ListenerRunning:    true,
		LastProcessedBlock: last,
		UpdatedAt:          time.Now().Unix(),
	}
}

func (s *ListenerService) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	return s.repo.ListRecentEvents(ctx, limit)
}

// Finalized write (block + hash)
func (s *ListenerService) MarkFinalized(ctx context.Context, height uint64, hash string, parentHash string) error {
	return s.repo.UpsertBlock(ctx, height, hash, parentHash)
}

func (s *ListenerService) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.repo.GetLastProcessedBlock(ctx)
}

func (s *ListenerService) RollbackTo(ctx context.Context, height uint64) error {
	return s.repo.RollbackTo(ctx, height)
}

func (s *ListenerService) EnsureCanonical(
	ctx context.Context,
	target uint64,
	rpcGetHash func(context.Context, uint64) (string, error),
) (uint64, error) {
	lastProcessed, err := s.repo.GetLastProcessedBlock(ctx)
	if err != nil {
		return 0, fmt.Errorf("get last processed: %w", err)
	}

	if lastProcessed == 0 {
		return bootstrapStart(target), nil
	}

	if lastProcessed > target {
		lastProcessed, err = s.repo.GetMaxBlockNumberLE(ctx, target)
		if err != nil {
			if errors.Is(err, repository.ErrBlockNotFound) {
				return bootstrapStart(target), nil
			}
			return 0, fmt.Errorf("get max block <= %d: %w", target, err)
		}
	}

	dbHash, err := s.repo.GetBlockHash(ctx, lastProcessed)
	if err != nil {
		return 0, fmt.Errorf("get db hash at %d: %w", lastProcessed, err)
	}

	rpcHash, err := rpcGetHash(ctx, lastProcessed)
	if err != nil {
		return 0, fmt.Errorf("get rpc hash at %d: %w", lastProcessed, err)
	}

	if dbHash == rpcHash {
		return lastProcessed, nil
	}

	lca, err := s.FindLCA(ctx, lastProcessed, rpcGetHash)
	if err != nil {
		return 0, err
	}

	if err := s.repo.RollbackTo(ctx, lca); err != nil {
		return 0, fmt.Errorf("rollback to %d: %w", lca, err)
	}

	return lca, nil
}

func (s *ListenerService) FindLCA(
	ctx context.Context,
	start uint64,
	rpcGetHash func(context.Context, uint64) (string, error),
) (uint64, error) {
	for h := start; ; h-- {
		dbHash, err := s.repo.GetBlockHash(ctx, h)
		if err != nil {
			if errors.Is(err, repository.ErrBlockNotFound) {
				return 0, repository.ErrLCANotFound
			}
			return 0, fmt.Errorf("get db hash at %d: %w", h, err)
		}

		rpcHash, err := rpcGetHash(ctx, h)
		if err != nil {
			return 0, fmt.Errorf("get rpc hash at %d: %w", h, err)
		}

		if dbHash == rpcHash {
			return h, nil
		}

		if h == 0 {
			break
		}
	}

	return 0, repository.ErrLCANotFound
}

func (s *ListenerService) SyncFinalized(
	ctx context.Context,
	getFinalizedHead func(context.Context) (uint64, error),
	rpcGetHash func(context.Context, uint64) (string, error),
) error {
	finalizedHead, err := getFinalizedHead(ctx)
	if err != nil {
		return fmt.Errorf("get finalized head: %w", err)
	}

	canonicalHeight, err := s.EnsureCanonical(ctx, finalizedHead, rpcGetHash)
	if err != nil {
		return err
	}

	for h := canonicalHeight + 1; h <= finalizedHead; h++ {
		hash, err := rpcGetHash(ctx, h)
		if err != nil {
			return fmt.Errorf("get hash at %d: %w", h, err)
		}

		parentHash := ""
		if h > 0 {
			parentHash, err = rpcGetHash(ctx, h-1)
			if err != nil {
				return fmt.Errorf("get parent hash at %d: %w", h-1, err)
			}
		}

		if err := s.repo.UpsertBlock(ctx, h, hash, parentHash); err != nil {
			return fmt.Errorf("upsert block %d: %w", h, err)
		}
	}

	log.Printf("finalizedHead=%d canonicalHeight=%d",
    finalizedHead, canonicalHeight)

	return nil
}

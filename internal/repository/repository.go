package repository

import (
	"context"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type Repository interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	GetMaxBlockNumberLE(ctx context.Context, target uint64) (uint64, error)
	GetBlockHash(ctx context.Context, number uint64) (string, error)
	UpsertBlock(ctx context.Context, block uint64, hash string, parentHash string) error
	RollbackTo(ctx context.Context, block uint64) error
	ListRecentEvents(ctx context.Context, limit int) []model.Event
}

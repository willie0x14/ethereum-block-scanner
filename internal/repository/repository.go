package repository

import (
	"context"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type Repository interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	GetLastBlockHash(ctx context.Context) (string, error)
	SetProcessed(ctx context.Context, block uint64, hash string)
	ListRecentEvents(ctx context.Context, limit int) []model.Event
}


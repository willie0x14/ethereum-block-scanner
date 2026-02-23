package repository

import (
	"context"

	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type Repository interface {
	GetLastProcessedBlock(ctx context.Context) uint64
	SetLastProcessedBlock(ctx context.Context, block uint64)

	ListRecentEvents(ctx context.Context, limit int) []model.Event
}


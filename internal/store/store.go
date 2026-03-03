package store

import "context"

// TODO: wire FinalityStore into the active repository path after migration.
type FinalityStore interface {
	// state
	GetState(ctx context.Context, chainID string) (last uint64, lastHash string, err error)
	UpsertState(ctx context.Context, chainID string, last uint64, lastHash string) error

	// blocks
	UpsertBlock(ctx context.Context, chainID string, number uint64, hash, parentHash string) error
	GetBlockHash(ctx context.Context, chainID string, number uint64) (string, error)
	DeleteBlocksAfter(ctx context.Context, chainID string, number uint64) error
}

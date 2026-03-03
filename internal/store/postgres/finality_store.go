package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FinalityPGStore struct {
	pool *pgxpool.Pool
}

func NewFinalityPGStore(pool *pgxpool.Pool) *FinalityPGStore {
	return &FinalityPGStore{pool: pool}
}

func (s *FinalityPGStore) GetState(ctx context.Context, chainID string) (uint64, string, error) {
	// 確保 row 存在（避免第一次查不到）
	_, err := s.pool.Exec(ctx, `
		INSERT INTO public.listener_state(chain_id)
		VALUES ($1)
		ON CONFLICT (chain_id) DO NOTHING
	`, chainID)
	if err != nil {
		return 0, "", err
	}

	var last int64
	var lastHash string
	err = s.pool.QueryRow(ctx, `
		SELECT last_finalized_block, last_finalized_hash
		FROM public.listener_state
		WHERE chain_id = $1
	`, chainID).Scan(&last, &lastHash)
	if err != nil {
		return 0, "", err
	}

	if last < 0 {
		last = 0
	}
	return uint64(last), lastHash, nil
}

func (s *FinalityPGStore) UpsertState(ctx context.Context, chainID string, last uint64, lastHash string) error {

	_, err := s.pool.Exec(ctx, `
		INSERT INTO public.listener_state(chain_id, last_finalized_block, last_finalized_hash, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (chain_id)
		DO UPDATE SET
		  last_finalized_block = EXCLUDED.last_finalized_block,
		  last_finalized_hash  = EXCLUDED.last_finalized_hash,
		  updated_at           = NOW()
	`, chainID, int64(last), lastHash)
	return err
}

func (s *FinalityPGStore) UpsertBlock(
	ctx context.Context,
	chainID string,
	number uint64,
	hash, parentHash string,
) error {
	sql := `
		INSERT INTO public.blocks(chain_id, number, hash, parent_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chain_id, number)
		DO UPDATE SET
			hash = EXCLUDED.hash,
			parent_hash = EXCLUDED.parent_hash
	`

	_, err := s.pool.Exec(ctx, sql, chainID, int64(number), hash, parentHash)
	return err
}

func (s *FinalityPGStore) GetBlockHash(ctx context.Context, chainID string, number uint64) (string, error) {
	var hash string
	err := s.pool.QueryRow(ctx, `
		SELECT hash
		FROM public.blocks
		WHERE chain_id = $1 AND number = $2
	`, chainID, int64(number)).Scan(&hash)
	return hash, err
}

func (s *FinalityPGStore) DeleteBlocksAfter(ctx context.Context, chainID string, number uint64) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM public.blocks
		WHERE chain_id = $1 AND number > $2
	`, chainID, int64(number))
	return err
}

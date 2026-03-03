package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/willie0x14/ethereum-block-scanner/internal/model"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

const defaultChainID = "sepolia"

var _ Repository = (*PostgresRepository)(nil)

func (p *PostgresRepository) GetBlockHash(ctx context.Context, number uint64) (string, error) {
	var hash string
	err := p.pool.QueryRow(ctx,
		`SELECT hash FROM public.blocks
	         WHERE chain_id = $1 AND number = $2`,
		defaultChainID,
		int64(number),
	).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBlockNotFound
		}
		return "", err
	}

	return hash, nil
}

func NewPostgresRepository(ctx context.Context, dbURL string) (Repository, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{
		pool: pool,
	}, nil
}

func (p *PostgresRepository) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	var block int64
	err := p.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(number), 0) FROM public.blocks WHERE chain_id = $1`,
		defaultChainID,
	).Scan(&block)
	if err != nil {
		return 0, err
	}

	return uint64(block), nil
}

func (p *PostgresRepository) GetMaxBlockNumberLE(ctx context.Context, target uint64) (uint64, error) {
	var maxNumber *int64
	err := p.pool.QueryRow(ctx,
		`SELECT MAX(number) FROM public.blocks
         WHERE chain_id = $1 AND number <= $2`,
		defaultChainID,
		int64(target),
	).Scan(&maxNumber)
	if err != nil {
		return 0, err
	}
	if maxNumber == nil {
		return 0, ErrBlockNotFound
	}

	return uint64(*maxNumber), nil
}

func (p *PostgresRepository) UpsertBlock(
	ctx context.Context,
	block uint64,
	hash string,
	parentHash string,
) error {

	sql := `
		INSERT INTO public.blocks(chain_id, number, hash, parent_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chain_id, number)
		DO UPDATE SET
			hash = EXCLUDED.hash,
			parent_hash = EXCLUDED.parent_hash
	`

	_, err := p.pool.Exec(
		ctx,
		sql,
		defaultChainID,
		int64(block),
		hash,
		parentHash,
	)

	return err
}

func (p *PostgresRepository) RollbackTo(ctx context.Context, block uint64) error {
	_, err := p.pool.Exec(ctx,
		`DELETE FROM public.blocks
			WHERE chain_id = $1 AND number > $2`,
		defaultChainID,
		block,
	)

	return err
}

func (p *PostgresRepository) ListRecentEvents(ctx context.Context, limit int) []model.Event {
	return []model.Event{}
}

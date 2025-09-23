package domain

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	InsertDomain(ctx context.Context, d string) (bool, error)
	GetAllDomains(ctx context.Context) ([]string, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) GetAllDomains(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT domain FROM domain`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		domains = append(domains, h)
	}
	return domains, nil
}

func (r *PGRepository) InsertDomain(ctx context.Context, d string) (bool, error) {
	if d == "" {
		return false, errors.New("empty domain")
	}

	const q = `
		INSERT INTO domain (domain)
		VALUES ($1)
		ON CONFLICT (domain) DO NOTHING
	`

	ct, err := r.pool.Exec(ctx, q, d)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

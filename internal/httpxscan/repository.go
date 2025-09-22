package httpxscan

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HttpxRepository interface {
	SaveResult(ctx context.Context, rec HttpxRecord) error
	GetAll(ctx context.Context) ([]HttpxRecord, error)
}

type PGHttpxRepository struct {
	pool *pgxpool.Pool
}

func NewPGHttpxRepository(pool *pgxpool.Pool) *PGHttpxRepository {
	return &PGHttpxRepository{pool: pool}
}

func (r *PGHttpxRepository) SaveResult(ctx context.Context, rec HttpxRecord) error {
	_, err := r.pool.Exec(
		ctx,
		`INSERT INTO httpx (host, status, title, location, url, technologies)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		rec.Host, rec.Status, rec.Title, rec.Location, rec.URL, rec.Technologies,
	)
	return err
}

func (r *PGHttpxRepository) GetAll(ctx context.Context) ([]HttpxRecord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT host, status, title, location, url, technologies FROM httpx`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []HttpxRecord
	for rows.Next() {
		var rec HttpxRecord
		if err := rows.Scan(
			&rec.Host,
			&rec.Status,
			&rec.Title,
			&rec.Location,
			&rec.URL,
			&rec.Technologies,
		); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, nil
}

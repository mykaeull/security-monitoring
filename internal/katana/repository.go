package katana

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type KatanaRepository interface {
	SaveRecords(ctx context.Context, recs []STRecord) error
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) SaveRecords(ctx context.Context, recs []STRecord) error {
	if len(recs) == 0 {
		return nil
	}

	const q = `
		INSERT INTO katana (occurred_at, method, endpoint, status_code, content_length, content_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`

	for _, rec := range recs {
		_, err := r.pool.Exec(
			ctx, q,
			rec.OccurredAt,
			rec.Request.Method,
			rec.Request.Endpoint,
			rec.Response.StatusCode,
			rec.Response.ContentLength,
			rec.Response.ContentType,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

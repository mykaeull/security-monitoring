package subdomainenum

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
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

	// Inserimos um a um (simples). Em produção, dá pra usar Batch para maior performance.
	for _, rec := range recs {
		// host_provider (TEXT[]), hostname (TEXT), ips (TEXT[])
		_, err := r.pool.Exec(
			ctx,
			`INSERT INTO subdomain (host_provider, hostname, ips)
			 VALUES ($1, $2, $3)
			 ON CONFLICT DO NOTHING`, // opcional se você depois criar UNIQUE(hostname)
			rec.HostProvider,
			rec.Hostname,
			rec.IPs,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

package subdomainenum

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	SaveRecords(ctx context.Context, recs []STRecord) error
	GetAllRecords(ctx context.Context) ([]STRecord, error)
	GetAllHostnames(ctx context.Context) ([]string, error)
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

func (r *PGRepository) GetAllRecords(ctx context.Context) ([]STRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT host_provider, hostname, ips FROM subdomain`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []STRecord
	for rows.Next() {
		var rec STRecord
		err := rows.Scan(&rec.HostProvider, &rec.Hostname, &rec.IPs)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func (r *PGRepository) GetAllHostnames(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT hostname FROM subdomain`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

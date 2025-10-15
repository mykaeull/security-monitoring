package nuclei

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NucleiRepository interface {
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
		INSERT INTO nuclei (
			occurred_at,
			template_id,
			info_name,
			info_description,
			info_reference,
			info_severity,
			info_cve_id,
			info_cwe_ids,
			rec_type,
			host,
			port,
			scheme,
			url,
			matched_at,
			ip,
			curl_command
		)
		VALUES (
			$1,  $2,  $3,  $4,
			$5,  $6,  $7,  $8,
			$9,  $10, $11, $12,
			$13, $14, $15, $16
		)
		ON CONFLICT DO NOTHING
	`

	for _, rec := range recs {
		// Garantir slices não-nil para TEXT[]
		refs := rec.Info.Reference
		if refs == nil {
			refs = []string{}
		}
		cwes := rec.Info.Classification.CWEID
		if cwes == nil {
			cwes = []string{}
		}

		// info_cve_id (TEXT único). Use o primeiro, ou NULL.
		var cve any
		if len(rec.Info.Classification.CVEID) > 0 && rec.Info.Classification.CVEID[0] != "" {
			cve = rec.Info.Classification.CVEID[0]
		} else {
			cve = nil
		}

		// 1 INSERT por registro
		if _, err := r.pool.Exec(
			ctx, q,
			rec.OccurredAt,       // $1  TIMESTAMPTZ
			rec.TemplateID,       // $2  TEXT
			rec.Info.Name,        // $3  TEXT
			rec.Info.Description, // $4  TEXT
			refs,                 // $5  TEXT[]
			rec.Info.Severity,    // $6  TEXT
			cve,                  // $7  TEXT (ou NULL)
			cwes,                 // $8  TEXT[]
			rec.Type,             // $9  TEXT
			rec.Host,             // $10 TEXT
			rec.Port,             // $11 TEXT (atenção ao tipo da coluna)
			rec.Scheme,           // $12 TEXT
			rec.URL,              // $13 TEXT
			rec.MatchedAt,        // $14 TEXT
			rec.IP,               // $15 TEXT
			rec.CurlCommand,      // $16 TEXT
		); err != nil {
			return err
		}
	}

	return nil
}

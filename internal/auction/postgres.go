package auction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Partners(ctx context.Context) ([]Partner, error) {
	const query = `
		SELECT uuid::text, code, name, endpoint, is_enabled, countries,
		       device_types, min_bid_floor, blocked_categories
		FROM partners
		ORDER BY code`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query partners: %w", err)
	}
	defer rows.Close()

	partners := make([]Partner, 0)
	for rows.Next() {
		var partner Partner
		if err := rows.Scan(
			&partner.UUID,
			&partner.Code,
			&partner.Name,
			&partner.Endpoint,
			&partner.IsEnabled,
			&partner.Countries,
			&partner.DeviceTypes,
			&partner.MinBidFloor,
			&partner.BlockedCategories,
		); err != nil {
			return nil, fmt.Errorf("scan partner: %w", err)
		}
		partners = append(partners, partner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate partners: %w", err)
	}

	return partners, nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}

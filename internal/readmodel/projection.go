package readmodel

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/peyrone/go-event-sourcing/internal/domain"
)

type Projector struct{ db *sql.DB }

func NewProjector(db *sql.DB) *Projector { return &Projector{db: db} }

func (p *Projector) Init() error {
	_, err := p.db.Exec(`CREATE TABLE IF NOT EXISTS order_summary(
							order_id TEXT PRIMARY KEY,
							customer TEXT,
							total_items INTEGER,
							checked_out INTEGER
						);`,
	)
	return err
}

func (p *Projector) Apply(ctx context.Context, e domain.Event) error {
	switch ev := e.(type) {
	case domain.OrderCreated:
		_, err := p.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO order_summary(order_id,customer,total_items,checked_out)
							VALUES (?,?,0,0)`,
			ev.ID,
			ev.Customer,
		)
		return err
	case domain.ItemAdded:
		_, err := p.db.ExecContext(ctx,
			`UPDATE order_summary SET total_items = total_items + ?
							WHERE order_id=?`,
			ev.Qty,
			ev.OrderID,
		)
		return err
	case domain.ItemRemoved:
		_, err := p.db.ExecContext(ctx,
			`UPDATE order_summary SET total_items = MAX(total_items - ?, 0)
							WHERE order_id=?`,
			ev.Qty,
			ev.OrderID,
		)
		return err
	case domain.OrderCheckedOut:
		_, err := p.db.ExecContext(ctx,
			`UPDATE order_summary SET checked_out = 1 WHERE order_id=?`,
			ev.ID,
		)
		return err
	default:
		_ = json.NewDecoder(nil)
		return nil
	}
}

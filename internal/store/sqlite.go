package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/peyrone/go-event-sourcing/internal/domain"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schema := `CREATE TABLE IF NOT EXISTS events(
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					aggregate_id TEXT NOT NULL,
					version INTEGER NOT NULL,
					name TEXT NOT NULL,
					occurred_at INTEGER NOT NULL,
					payload TEXT NOT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS ux_stream ON events(aggregate_id, version);
				CREATE INDEX IF NOT EXISTS ix_time ON events(occurred_at);`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func marshal(e domain.Event) (
	name string,
	when int64,
	payload []byte,
	agg string,
	err error,
) {
	switch v := e.(type) {
	case domain.OrderCreated:
		name, when, agg = v.Name(), v.OccurredAt().Unix(), v.AggregateID()
		payload, err = json.Marshal(v)
	case domain.ItemAdded:
		name, when, agg = v.Name(), v.OccurredAt().Unix(), v.AggregateID()
		payload, err = json.Marshal(v)
	case domain.ItemRemoved:
		name, when, agg = v.Name(), v.OccurredAt().Unix(), v.AggregateID()
		payload, err = json.Marshal(v)
	case domain.OrderCheckedOut:
		name, when, agg = v.Name(), v.OccurredAt().Unix(), v.AggregateID()
		payload, err = json.Marshal(v)
	default:
		err = errors.New("unknown event")
	}
	return
}

func (s *SQLiteStore) Append(
	ctx context.Context,
	events []domain.Event,
	expectedVersion int,
) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return expectedVersion, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var current int
	row := tx.QueryRowContext(ctx,
		`SELECT IFNULL(MAX(version),0) FROM events WHERE aggregate_id=?`,
		events[0].AggregateID(),
	)
	if err = row.Scan(&current); err != nil {
		return expectedVersion, err
	}
	if current != expectedVersion {
		return expectedVersion, errors.New("concurrency conflict")
	}

	for i, e := range events {
		name, when, payload, agg, err2 := marshal(e)
		if err2 != nil {
			err = err2
			return expectedVersion, err
		}
		ver := expectedVersion + i + 1
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO events(aggregate_id,version,name,occurred_at,payload)
						VALUES(?,?,?,?,?)`,
			agg,
			ver,
			name,
			when,
			string(payload),
		)
		if err != nil {
			return expectedVersion, err
		}
	}
	if err = tx.Commit(); err != nil {
		return expectedVersion, err
	}
	return expectedVersion + len(events), nil
}

func (s *SQLiteStore) Load(
	ctx context.Context,
	aggregateID string,
) ([]domain.Event, int, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT name, occurred_at, payload, version
						FROM events WHERE aggregate_id=? ORDER BY version ASC`,
		aggregateID,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Event
	var latest int
	for rows.Next() {
		var name, payload string
		var when int64
		var ver int
		if err = rows.Scan(&name, &when, &payload, &ver); err != nil {
			return nil, 0, err
		}

		latest = ver
		switch name {
		case "OrderCreated":
			var e domain.OrderCreated
			_ = json.Unmarshal([]byte(payload), &e)
			out = append(out, e)
		case "ItemAdded":
			var e domain.ItemAdded
			_ = json.Unmarshal([]byte(payload), &e)
			out = append(out, e)
		case "ItemRemoved":
			var e domain.ItemRemoved
			_ = json.Unmarshal([]byte(payload), &e)
			out = append(out, e)
		case "OrderCheckedOut":
			var e domain.OrderCheckedOut
			_ = json.Unmarshal([]byte(payload), &e)
			out = append(out, e)
		}
	}

	return out, latest, nil
}

func (s *SQLiteStore) ListByTime(
	ctx context.Context,
	fromUnix, toUnix int64,
) ([]domain.Event, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT name, occurred_at, payload FROM events
						WHERE occurred_at BETWEEN ? AND ? ORDER BY occurred_at ASC`,
		fromUnix,
		toUnix,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var out []domain.Event

	for rows.Next() {
		var n, p string
		var t int64
		if err := rows.Scan(&n, &t, &p); err != nil {
			return nil, err
		}

		// Deserialize based on event name
		switch n {
		case "OrderCreated":
			var e domain.OrderCreated
			_ = json.Unmarshal([]byte(p), &e)
			out = append(out, e)
		case "ItemAdded":
			var e domain.ItemAdded
			_ = json.Unmarshal([]byte(p), &e)
			out = append(out, e)
		case "ItemRemoved":
			var e domain.ItemRemoved
			_ = json.Unmarshal([]byte(p), &e)
			out = append(out, e)
		case "OrderCheckedOut":
			var e domain.OrderCheckedOut
			_ = json.Unmarshal([]byte(p), &e)
			out = append(out, e)
		}
	}
	return out, nil
}

package store

import (
	"context"

	"github.com/peyrone/go-event-sourcing/internal/domain"
)

type EventStore interface {
	Append(
		ctx context.Context,
		events []domain.Event,
		expectedVersion int,
	) (newVersion int, err error)
	Load(
		ctx context.Context,
		aggregateID string,
	) (events []domain.Event, version int, err error)
	ListByTime(
		ctx context.Context,
		fromUnix, toUnix int64,
	) ([]domain.Event, error)
}

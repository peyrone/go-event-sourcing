package domain

import "time"

type Event interface {
	Name() string
	AggregateID() string
	OccurredAt() time.Time
}

type OrderCreated struct {
	ID       string
	Customer string
	At       time.Time
}

func (e OrderCreated) Name() string {
	return "OrderCreated"
}
func (e OrderCreated) AggregateID() string {
	return e.ID
}
func (e OrderCreated) OccurredAt() time.Time {
	return e.At
}

type ItemAdded struct {
	OrderID string
	SKU     string
	Qty     int
	At      time.Time
}

func (e ItemAdded) Name() string {
	return "ItemAdded"
}
func (e ItemAdded) AggregateID() string {
	return e.OrderID
}
func (e ItemAdded) OccurredAt() time.Time {
	return e.At
}

type ItemRemoved struct {
	OrderID string
	SKU     string
	Qty     int
	At      time.Time
}

func (e ItemRemoved) Name() string {
	return "ItemRemoved"
}
func (e ItemRemoved) AggregateID() string {
	return e.OrderID
}
func (e ItemRemoved) OccurredAt() time.Time {
	return e.At
}

type OrderCheckedOut struct {
	ID string
	At time.Time
}

func (e OrderCheckedOut) Name() string {
	return "OrderCheckedOut"
}
func (e OrderCheckedOut) AggregateID() string {
	return e.ID
}
func (e OrderCheckedOut) OccurredAt() time.Time {
	return e.At
}

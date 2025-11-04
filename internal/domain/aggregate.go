package domain

import "errors"

type Order struct {
	ID        string
	Customer  string
	Items     map[string]int
	CheckedOut bool
}

func NewOrder(id, customer string) *Order {
	return &Order{ID: id, Customer: customer, Items: map[string]int{}}
}

func (o *Order) Apply(e Event) {
	switch ev := e.(type) {
	case OrderCreated:
		o.ID, o.Customer = ev.ID, ev.Customer
	case ItemAdded:
		o.Items[ev.SKU] += ev.Qty
	case ItemRemoved:
		o.Items[ev.SKU] -= ev.Qty
		if o.Items[ev.SKU] <= 0 { delete(o.Items, ev.SKU) }
	case OrderCheckedOut:
		o.CheckedOut = true
	}
}

func (o *Order) EnsureNotCheckedOut() error {
	if o.CheckedOut { return errors.New("order already checked out") }
	return nil
}

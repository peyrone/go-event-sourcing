package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"database/sql"

	"github.com/peyrone/go-event-sourcing/internal/domain"
	"github.com/peyrone/go-event-sourcing/internal/store"
)

type Server struct {
	ES    store.EventStore
	DB    *sql.DB // Read model DB
	Apply func(ctx context.Context, e domain.Event) error
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	type req struct{ ID, Customer string }
	var q req
	_ = json.NewDecoder(r.Body).Decode(&q)

	now := time.Now()
	ev := domain.OrderCreated{ID: q.ID, Customer: q.Customer, At: now}

	_, v, _ := s.ES.Load(r.Context(), q.ID)
	newVer, err := s.ES.Append(r.Context(), []domain.Event{ev}, v)

	if err != nil {
		writeJSON(w, 409, map[string]string{"error": err.Error()})
		return
	}

	_ = s.Apply(r.Context(), ev)
	writeJSON(w, 201, map[string]any{"version": newVer})
}

func (s *Server) AddItem(w http.ResponseWriter, r *http.Request) {
	type req struct {
		OrderID, SKU string
		Qty          int
	}
	var q req
	_ = json.NewDecoder(r.Body).Decode(&q)
	ev := domain.ItemAdded{OrderID: q.OrderID, SKU: q.SKU, Qty: q.Qty, At: time.Now()}
	events, ver, _ := s.ES.Load(r.Context(), q.OrderID)

	// Reconstruct current state to enforce business rules
	order := domain.NewOrder(q.OrderID, "")
	for _, e := range events {
		order.Apply(e)
	}

	if err := order.EnsureNotCheckedOut(); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	newVer, err := s.ES.Append(r.Context(), []domain.Event{ev}, ver)
	if err != nil {
		writeJSON(w, 409, map[string]string{"error": err.Error()})
		return
	}

	_ = s.Apply(r.Context(), ev)
	writeJSON(w, 200, map[string]any{"version": newVer})
}

func (s *Server) RemoveItem(w http.ResponseWriter, r *http.Request) {
	type req struct {
		OrderID, SKU string
		Qty          int
	}

	var q req
	_ = json.NewDecoder(r.Body).Decode(&q)
	ev := domain.ItemRemoved{OrderID: q.OrderID, SKU: q.SKU, Qty: q.Qty, At: time.Now()}
	_, ver, _ := s.ES.Load(r.Context(), q.OrderID)
	newVer, err := s.ES.Append(r.Context(), []domain.Event{ev}, ver)

	if err != nil {
		writeJSON(w, 409, map[string]string{"error": err.Error()})
		return
	}

	_ = s.Apply(r.Context(), ev)
	writeJSON(w, 200, map[string]any{"version": newVer})
}

func (s *Server) Checkout(w http.ResponseWriter, r *http.Request) {
	type req struct{ ID string }
	var q req
	_ = json.NewDecoder(r.Body).Decode(&q)
	ev := domain.OrderCheckedOut{ID: q.ID, At: time.Now()}
	events, ver, _ := s.ES.Load(r.Context(), q.ID)

	order := domain.NewOrder(q.ID, "")
	for _, e := range events {
		order.Apply(e)
	}

	if err := order.EnsureNotCheckedOut(); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	newVer, err := s.ES.Append(r.Context(), []domain.Event{ev}, ver)
	if err != nil {
		writeJSON(w, 409, map[string]string{"error": err.Error()})
		return
	}

	_ = s.Apply(r.Context(), ev)
	writeJSON(w, 200, map[string]any{"version": newVer})
}

func (s *Server) GetSummary(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	row := s.DB.QueryRow(`SELECT order_id, customer, total_items, checked_out FROM order_summary WHERE order_id=?`, orderID)
	var id, customer string
	var total, out int

	if err := row.Scan(&id, &customer, &total, &out); err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, 200, map[string]any{"orderId": id, "customer": customer, "totalItems": total, "checkedOut": out == 1})
}

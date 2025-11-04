package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/peyrone/go-event-sourcing/internal/api"
	"github.com/peyrone/go-event-sourcing/internal/readmodel"
	"github.com/peyrone/go-event-sourcing/internal/store"
)

func main() {
	es, err := store.NewSQLite("events.db")
	if err != nil {
		log.Fatal(err)
	}

	// share same db handle for projections
	db, _ := sql.Open("sqlite", "events.db")
	prj := readmodel.NewProjector(db)
	if err := prj.Init(); err != nil {
		log.Fatal(err)
	}

	srv := &api.Server{
		ES:    es,
		DB:    db,
		Apply: prj.Apply, // inline projection for simplicity
	}

	http.HandleFunc("/orders/create", srv.CreateOrder)
	http.HandleFunc("/orders/add-item", srv.AddItem)
	http.HandleFunc("/orders/remove-item", srv.RemoveItem)
	http.HandleFunc("/orders/checkout", srv.Checkout)
	http.HandleFunc("/orders/summary", srv.GetSummary)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

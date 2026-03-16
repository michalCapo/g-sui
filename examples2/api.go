package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/michalCapo/g-sui/ui"
)

// RegisterAPI registers all REST API endpoints on the app.
func RegisterAPI(app *ui.App, store *InvoiceStore) {
	// GET /api/invoices — list all invoices (without line items)
	app.GET("/api/invoices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.All())
	})

	// POST /api/invoices — create a new invoice
	app.POST("/api/invoices", func(w http.ResponseWriter, r *http.Request) {
		var inv Invoice
		if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if inv.Company == "" {
			http.Error(w, `{"error":"company is required"}`, http.StatusBadRequest)
			return
		}
		if len(inv.Items) == 0 {
			http.Error(w, `{"error":"at least one line item is required"}`, http.StatusBadRequest)
			return
		}

		created := store.Create(inv)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	})

	// GET /api/invoices/{id} — get invoice detail with line items
	// DELETE /api/invoices/{id} — delete an invoice
	// Since custom routes use exact match, we handle both via a single Custom handler
	// that parses the ID from the URL path.
	app.Custom("GET", "/api/invoices/detail", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			http.Error(w, `{"error":"invalid invoice id"}`, http.StatusBadRequest)
			return
		}

		inv := store.ByID(uint(id))
		if inv == nil {
			http.Error(w, `{"error":"invoice not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(inv)
	})

	app.Custom("DELETE", "/api/invoices/delete", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			// try body
			var body struct {
				ID uint `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.ID > 0 {
				idStr = strconv.FormatUint(uint64(body.ID), 10)
			}
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			http.Error(w, `{"error":"invalid invoice id"}`, http.StatusBadRequest)
			return
		}

		if !store.Delete(uint(id)) {
			http.Error(w, `{"error":"invoice not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "message": "Invoice deleted"})
	})

}

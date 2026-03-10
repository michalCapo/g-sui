package pages

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// RegisterNewFeatureAPIs registers all /api/new/* endpoints for the new feature demo pages.
func RegisterNewFeatureAPIs(app *ui.App) {

	// --- Filterable invoices with enum/bool/date/number columns ---
	app.GET("/api/new/filterable-invoices", func(w http.ResponseWriter, r *http.Request) {
		statuses := []string{"paid", "pending", "overdue", "cancelled"}
		invoices := make([]map[string]any, 30)
		now := time.Now()
		for i := range invoices {
			daysAgo := rand.Intn(365)
			date := now.AddDate(0, 0, -daysAgo)
			amount := 500 + rand.Float64()*49500
			invoices[i] = map[string]any{
				"id":       fmt.Sprintf("INV-%04d", 1000+i),
				"company":  demoCompanies[rand.Intn(len(demoCompanies))],
				"amount":   math.Round(amount*100) / 100,
				"date":     date.Format("2006-01-02"),
				"status":   statuses[rand.Intn(len(statuses))],
				"verified": rand.Intn(2) == 1,
				"priority": rand.Intn(5) + 1,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(invoices)
	})

	// --- Invoice detail (for async row detail) ---
	app.GET("/api/new/invoice-detail", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		time.Sleep(300 * time.Millisecond) // simulate latency
		detail := map[string]any{
			"id":          id,
			"contact":     "John Smith",
			"email":       "john@example.com",
			"phone":       "+1-555-0123",
			"address":     "123 Business St, Suite 100",
			"notes":       "Premium customer. Net-30 payment terms.",
			"created":     time.Now().AddDate(0, -6, 0).Format("2006-01-02"),
			"lastPayment": time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(detail)
	})

	// --- Two-series chart data ---
	app.GET("/api/new/revenue-comparison", func(w http.ResponseWriter, r *http.Request) {
		months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		data := make([]map[string]any, len(months))
		for i, m := range months {
			data[i] = map[string]any{
				"label":  m,
				"value":  8000 + rand.Intn(12000),
				"value2": 5000 + rand.Intn(15000),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// --- Badge count ---
	app.GET("/api/new/notification-count", func(w http.ResponseWriter, r *http.Request) {
		count := rand.Intn(15)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"count": count})
	})

	// --- Autocomplete data ---
	app.GET("/api/new/autocomplete-cities", func(w http.ResponseWriter, r *http.Request) {
		cities := []map[string]string{
			{"value": "new-york", "label": "New York"},
			{"value": "los-angeles", "label": "Los Angeles"},
			{"value": "chicago", "label": "Chicago"},
			{"value": "houston", "label": "Houston"},
			{"value": "phoenix", "label": "Phoenix"},
			{"value": "philadelphia", "label": "Philadelphia"},
			{"value": "san-antonio", "label": "San Antonio"},
			{"value": "san-diego", "label": "San Diego"},
			{"value": "dallas", "label": "Dallas"},
			{"value": "san-jose", "label": "San Jose"},
			{"value": "bratislava", "label": "Bratislava"},
			{"value": "kosice", "label": "Kosice"},
			{"value": "prague", "label": "Prague"},
			{"value": "brno", "label": "Brno"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cities)
	})

	// --- AJAX form endpoint ---
	app.POST("/api/new/submit-form", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond) // simulate processing
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "Form submitted successfully at " + time.Now().Format("15:04:05"),
		})
	})

	// --- File upload endpoint ---
	app.POST("/api/new/upload", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20) // 32MB
		files := r.MultipartForm.File["files"]
		results := make([]map[string]any, len(files))
		for i, fh := range files {
			results[i] = map[string]any{
				"name": fh.Filename,
				"size": fh.Size,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// --- Health check endpoint ---
	app.GET("/api/new/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	// --- Async button endpoint ---
	app.POST("/api/new/async-action", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(800 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"message": "Action completed at " + time.Now().Format("15:04:05"),
		})
	})

	// --- Conditional polling data ---
	app.GET("/api/new/job-status", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a job that progresses each call
		progress := rand.Intn(30) + 70 // 70-99
		status := "processing"
		if progress >= 95 {
			progress = 100
			status = "completed"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":   status,
			"progress": progress,
			"message":  fmt.Sprintf("Processing... %d%%", progress),
			"updated":  time.Now().Format("15:04:05"),
		})
	})
}

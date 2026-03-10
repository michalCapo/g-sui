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

// Invoice represents a mock invoice record for client-side demo.
type Invoice struct {
	Number  string  `json:"Number"`
	Company string  `json:"Company"`
	Amount  float64 `json:"Amount"`
	Date    string  `json:"Date"`
	Status  string  `json:"Status"`
}

var demoCompanies = []string{
	"Acme Corp", "Globex Industries", "Stark Enterprises", "Wayne Technologies",
	"Umbrella Corporation", "Cyberdyne Systems", "Oscorp Industries", "Soylent Corp",
	"Initech", "Hooli", "Pied Piper", "Massive Dynamic",
	"Aperture Science", "Black Mesa", "LexCorp", "Tyrell Corporation",
	"Weyland-Yutani", "Buy n Large", "Dunder Mifflin", "Sterling Cooper",
}

var demoStatuses = []string{"paid", "pending", "overdue"}

func generateInvoices() []Invoice {
	invoices := make([]Invoice, 40)
	now := time.Now()

	for i := range invoices {
		daysAgo := rand.Intn(180) // within last 6 months
		date := now.AddDate(0, 0, -daysAgo)

		status := demoStatuses[rand.Intn(len(demoStatuses))]
		amount := 500 + rand.Float64()*49500 // 500 - 50000
		amount = math.Round(amount*100) / 100

		invoices[i] = Invoice{
			Number:  fmt.Sprintf("INV-%04d", 2024000+i+1),
			Company: demoCompanies[rand.Intn(len(demoCompanies))],
			Amount:  amount,
			Date:    date.Format("2006-01-02"),
			Status:  status,
		}
	}

	return invoices
}

// Process represents a live system process for the polling demo.
type Process struct {
	Name    string  `json:"Name"`
	Status  string  `json:"Status"`
	CPU     float64 `json:"CPU"`
	Memory  float64 `json:"Memory"`
	Updated string  `json:"Updated"`
}

var processNames = []string{
	"web-server", "db-primary", "cache-redis", "worker-queue", "scheduler", "api-gateway",
}

var processStatuses = []string{"running", "idle", "busy", "warning"}

func generateLiveData() []Process {
	processes := make([]Process, len(processNames))
	now := time.Now()

	for i, name := range processNames {
		processes[i] = Process{
			Name:    name,
			Status:  processStatuses[rand.Intn(len(processStatuses))],
			CPU:     math.Round(rand.Float64()*100*10) / 10,
			Memory:  math.Round(rand.Float64()*8192*100) / 100,
			Updated: now.Format("2006-01-02T15:04:05"),
		}
	}

	return processes
}

// RegisterClientDemoAPIs registers all /api/client-demo/... endpoints for the
// client-side rendering example pages.
func RegisterClientDemoAPIs(app *ui.App) {
	// Invoice data for table demos
	app.GET("/api/client-demo/invoices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateInvoices())
	})

	// Monthly revenue (bar chart)
	app.GET("/api/client-demo/revenue-monthly", func(w http.ResponseWriter, r *http.Request) {
		months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		data := make([]map[string]any, len(months))
		for i, m := range months {
			data[i] = map[string]any{"label": m, "value": 5000 + rand.Intn(15000)}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// Revenue by category (donut chart)
	app.GET("/api/client-demo/revenue-category", func(w http.ResponseWriter, r *http.Request) {
		categories := []string{"Software", "Consulting", "Hardware", "Support", "Training"}
		data := make([]map[string]any, len(categories))
		for i, c := range categories {
			data[i] = map[string]any{"label": c, "value": 3000 + rand.Intn(20000)}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// Trend data (area chart)
	app.GET("/api/client-demo/trend", func(w http.ResponseWriter, r *http.Request) {
		weeks := 12
		data := make([]map[string]any, weeks)
		base := 8000.0
		for i := 0; i < weeks; i++ {
			base += float64(rand.Intn(2000) - 800) // random walk
			if base < 2000 {
				base = 2000
			}
			data[i] = map[string]any{
				"label": fmt.Sprintf("W%d", i+1),
				"value": int(math.Round(base)),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// Top customers (horizontal bar chart)
	app.GET("/api/client-demo/top-customers", func(w http.ResponseWriter, r *http.Request) {
		customers := []string{
			"Acme Corp", "Globex Industries", "Stark Enterprises",
			"Hooli", "Pied Piper", "Initech", "Wayne Technologies",
		}
		data := make([]map[string]any, len(customers))
		for i, c := range customers {
			data[i] = map[string]any{"label": c, "value": 10000 + rand.Intn(90000)}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// Stats for KPI dashboard (single object, not array)
	app.GET("/api/client-demo/stats", func(w http.ResponseWriter, r *http.Request) {
		total := 600000 + rand.Intn(500000)
		count := 30 + rand.Intn(20)
		avg := float64(total) / float64(count)
		overdue := 2 + rand.Intn(8)

		data := map[string]any{
			"total":   total,
			"count":   count,
			"avg":     math.Round(avg*100) / 100,
			"overdue": overdue,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// Live data for polling demo (changes every call)
	app.GET("/api/client-demo/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateLiveData())
	})

	// Empty endpoint (returns empty array)
	app.GET("/api/client-demo/empty", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
	})

	// Error endpoint (returns 500)
	app.GET("/api/client-demo/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
	})
}

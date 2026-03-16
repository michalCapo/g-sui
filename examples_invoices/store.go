package main

import (
	"fmt"
	"sync"
	"time"
)

// InvoiceItem represents a single line item on an invoice.
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Total       float64 `json:"total"`
}

// Invoice represents an invoice record.
// JSON tags use lowercase keys so the client-side JS table can access `item.id` for expandable rows.
type Invoice struct {
	ID          uint          `json:"id"`
	Number      string        `json:"number"`
	Company     string        `json:"company"`
	Description string        `json:"description"`
	Amount      float64       `json:"amount"`
	Tax         float64       `json:"tax"`
	Total       float64       `json:"total"`
	Status      string        `json:"status"`
	DueDate     string        `json:"dueDate"`
	CreatedAt   string        `json:"createdAt"`
	Items       []InvoiceItem `json:"items,omitempty"`
}

// InvoiceStore is a thread-safe in-memory store for invoices.
type InvoiceStore struct {
	mu      sync.RWMutex
	data    []Invoice
	counter uint
}

// NewStore creates a new InvoiceStore pre-populated with seed data.
func NewStore() *InvoiceStore {
	s := &InvoiceStore{}
	s.seed()
	return s
}

// All returns all invoices without their line items (for list view).
func (s *InvoiceStore) All() []Invoice {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Invoice, len(s.data))
	for i, inv := range s.data {
		result[i] = Invoice{
			ID:        inv.ID,
			Number:    inv.Number,
			Company:   inv.Company,
			Amount:    inv.Amount,
			Tax:       inv.Tax,
			Total:     inv.Total,
			Status:    inv.Status,
			DueDate:   inv.DueDate,
			CreatedAt: inv.CreatedAt,
		}
	}
	return result
}

// ByID returns a single invoice with its line items.
func (s *InvoiceStore) ByID(id uint) *Invoice {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, inv := range s.data {
		if inv.ID == id {
			cp := inv
			return &cp
		}
	}
	return nil
}

// Create adds a new invoice to the store.
func (s *InvoiceStore) Create(inv Invoice) Invoice {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	inv.ID = s.counter
	inv.Number = fmt.Sprintf("INV-%04d", s.counter)
	inv.CreatedAt = time.Now().Format("2006-01-02")

	// calculate totals
	var amount float64
	for i := range inv.Items {
		inv.Items[i].Total = float64(inv.Items[i].Quantity) * inv.Items[i].UnitPrice
		amount += inv.Items[i].Total
	}
	inv.Amount = amount
	inv.Tax = amount * 0.20
	inv.Total = amount + inv.Tax

	s.data = append(s.data, inv)
	return inv
}

// Delete removes an invoice by ID. Returns true if found and removed.
func (s *InvoiceStore) Delete(id uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, inv := range s.data {
		if inv.ID == id {
			s.data = append(s.data[:i], s.data[i+1:]...)
			return true
		}
	}
	return false
}

func (s *InvoiceStore) seed() {
	companies := []string{
		"Acme Corp", "Globex Industries", "Stark Enterprises", "Wayne Technologies",
		"Umbrella Corp", "Cyberdyne Systems", "Oscorp Industries", "Soylent Corp",
		"Initech", "Hooli", "Pied Piper", "Massive Dynamic",
		"Aperture Science", "Black Mesa", "Dunder Mifflin",
	}

	statuses := []string{"draft", "sent", "paid", "paid", "overdue", "sent", "paid"}

	type seedItem struct {
		Desc  string
		Qty   int
		Price float64
	}

	itemPool := [][]seedItem{
		{
			{Desc: "Web Development", Qty: 40, Price: 95.00},
			{Desc: "UI/UX Design", Qty: 16, Price: 110.00},
			{Desc: "Project Management", Qty: 8, Price: 85.00},
		},
		{
			{Desc: "Annual License", Qty: 1, Price: 4999.00},
			{Desc: "Support Package", Qty: 1, Price: 1200.00},
		},
		{
			{Desc: "Consulting Hours", Qty: 24, Price: 150.00},
			{Desc: "Travel Expenses", Qty: 1, Price: 340.00},
			{Desc: "Report Preparation", Qty: 8, Price: 95.00},
		},
		{
			{Desc: "Server Hosting (Annual)", Qty: 1, Price: 2400.00},
			{Desc: "SSL Certificate", Qty: 3, Price: 79.00},
			{Desc: "Domain Registration", Qty: 5, Price: 14.99},
		},
		{
			{Desc: "Training Workshop", Qty: 2, Price: 750.00},
			{Desc: "Course Materials", Qty: 10, Price: 45.00},
		},
		{
			{Desc: "Database Migration", Qty: 1, Price: 3200.00},
			{Desc: "Data Cleanup", Qty: 16, Price: 85.00},
			{Desc: "Testing & QA", Qty: 24, Price: 75.00},
		},
		{
			{Desc: "API Integration", Qty: 32, Price: 120.00},
		},
		{
			{Desc: "Security Audit", Qty: 1, Price: 5500.00},
			{Desc: "Penetration Testing", Qty: 1, Price: 3200.00},
			{Desc: "Compliance Report", Qty: 1, Price: 1800.00},
		},
	}

	now := time.Now()

	for i, company := range companies {
		items := itemPool[i%len(itemPool)]
		invoiceItems := make([]InvoiceItem, len(items))
		var amount float64
		for j, item := range items {
			total := float64(item.Qty) * item.Price
			invoiceItems[j] = InvoiceItem{
				Description: item.Desc,
				Quantity:    item.Qty,
				UnitPrice:   item.Price,
				Total:       total,
			}
			amount += total
		}

		tax := amount * 0.20
		daysAgo := (i * 11) % 120
		dueOffset := 30 - daysAgo
		created := now.AddDate(0, 0, -daysAgo)
		due := created.AddDate(0, 0, 30)
		if dueOffset < -10 {
			// overdue ones get past due dates
		}

		s.counter++
		s.data = append(s.data, Invoice{
			ID:          s.counter,
			Number:      fmt.Sprintf("INV-%04d", s.counter),
			Company:     company,
			Description: fmt.Sprintf("Services for %s - %s %d", company, created.Month().String(), created.Year()),
			Amount:      amount,
			Tax:         tax,
			Total:       amount + tax,
			Status:      statuses[i%len(statuses)],
			DueDate:     due.Format("2006-01-02"),
			CreatedAt:   created.Format("2006-01-02"),
			Items:       invoiceItems,
		})
	}
}

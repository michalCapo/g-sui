package pages

import (
	"fmt"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Invoice Store
// ---------------------------------------------------------------------------

type InvoiceStore struct {
	mu      sync.RWMutex
	data    []Invoice
	counter uint
}

func NewInvoiceStore() *InvoiceStore {
	s := &InvoiceStore{}
	s.seed()
	return s
}

func (s *InvoiceStore) All() []Invoice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Invoice, len(s.data))
	copy(out, s.data)
	return out
}

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

func (s *InvoiceStore) Create(inv Invoice) Invoice {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	inv.ID = s.counter
	inv.Number = fmt.Sprintf("INV-%04d", s.counter)
	inv.CreatedAt = time.Now().Format("2006-01-02")

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

// ---------------------------------------------------------------------------
// Invoice seed
// ---------------------------------------------------------------------------

func (s *InvoiceStore) seed() {
	companies := []string{
		"Acme Corp", "Globex Industries", "Stark Enterprises", "Wayne Technologies",
		"Umbrella Corp", "Cyberdyne Systems", "Oscorp Industries", "Soylent Corp",
		"Initech", "Hooli", "Pied Piper", "Massive Dynamic",
	}
	statuses := []string{"draft", "sent", "paid", "paid", "overdue", "sent", "paid"}

	type seed struct {
		Desc  string
		Qty   int
		Price float64
	}
	pool := [][]seed{
		{{Desc: "Web Development", Qty: 40, Price: 95.00}, {Desc: "UI/UX Design", Qty: 16, Price: 110.00}},
		{{Desc: "Annual License", Qty: 1, Price: 4999.00}, {Desc: "Support Package", Qty: 1, Price: 1200.00}},
		{{Desc: "Consulting Hours", Qty: 24, Price: 150.00}, {Desc: "Travel Expenses", Qty: 1, Price: 340.00}},
		{{Desc: "Server Hosting", Qty: 1, Price: 2400.00}, {Desc: "SSL Certificate", Qty: 3, Price: 79.00}},
		{{Desc: "Training Workshop", Qty: 2, Price: 750.00}, {Desc: "Course Materials", Qty: 10, Price: 45.00}},
		{{Desc: "Security Audit", Qty: 1, Price: 5500.00}, {Desc: "Penetration Testing", Qty: 1, Price: 3200.00}},
	}

	now := time.Now()
	for i, company := range companies {
		items := pool[i%len(pool)]
		invoiceItems := make([]InvoiceItem, len(items))
		var amount float64
		for j, item := range items {
			total := float64(item.Qty) * item.Price
			invoiceItems[j] = InvoiceItem{Description: item.Desc, Quantity: item.Qty, UnitPrice: item.Price, Total: total}
			amount += total
		}
		tax := amount * 0.20
		daysAgo := (i * 11) % 90
		created := now.AddDate(0, 0, -daysAgo)
		due := created.AddDate(0, 0, 30)

		s.counter++
		s.data = append(s.data, Invoice{
			ID: s.counter, Number: fmt.Sprintf("INV-%04d", s.counter),
			Company: company, Description: fmt.Sprintf("Services for %s", company),
			Amount: amount, Tax: tax, Total: amount + tax,
			Status: statuses[i%len(statuses)], DueDate: due.Format("2006-01-02"),
			CreatedAt: created.Format("2006-01-02"), Items: invoiceItems,
		})
	}
}

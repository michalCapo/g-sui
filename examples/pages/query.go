package pages

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/michalCapo/g-sui/ui"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Person represents a demo record to exercise ui.TCollate features
type Person struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Email     string
	Country   string
	Status    string    // SELECT filter
	Active    bool      // BOOL filter
	CreatedAt time.Time // DATES filter
	LastLogin time.Time // ZERO_DATE / NOT_ZERO_DATE filter
	DeletedAt gorm.DeletedAt
}

func (Person) TableName() string { return "people" }

var (
	collateOnce sync.Once
	db          *gorm.DB
)

// initDB initializes an in-memory SQLite DB with 100 seeded records
func initDB() (*gorm.DB, error) {
	var err error
	collateOnce.Do(func() {
		// Use shared in-memory so the DB persists across connections
		dsn := "file:querydemo?mode=memory&cache=shared&_busy_timeout=5000"
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		// Enable custom normalize() SQLite function for better search
		_ = ui.RegisterSQLiteNormalize(db)

		// Auto-migrate schema
		if e := db.AutoMigrate(&Person{}); e != nil {
			err = e
			return
		}

		// Seed only if empty
		var count int64
		_ = db.Model(&Person{}).Count(&count).Error
		if count > 0 {
			return
		}

		rand.Seed(time.Now().UnixNano())
		statuses := []string{"new", "active", "blocked"}
		countries := []string{"Slovakia", "Czechia", "Austria", "Poland", "Hungary", "Germany", "Spain", "France", "Italy", "Portugal"}

		for i := 1; i <= 1000; i++ {
			name := fmt.Sprintf("User %03d", i)
			email := fmt.Sprintf("user%03d@example.com", i)
			country := countries[rand.Intn(len(countries))]
			status := statuses[rand.Intn(len(statuses))]
			active := rand.Intn(2) == 1
			// Spread CreatedAt over ~200 days
			created := time.Now().Add(-time.Duration(rand.Intn(200)) * 24 * time.Hour)
			// About half have LastLogin zero; others within last 100 days
			var lastLogin time.Time
			if rand.Intn(2) == 1 {
				lastLogin = time.Now().Add(-time.Duration(rand.Intn(100)) * 24 * time.Hour)
			} else {
				lastLogin = time.Time{} // zero
			}

			err = db.Create(&Person{
				Name:      name,
				Email:     email,
				Country:   country,
				Status:    status,
				Active:    active,
				CreatedAt: created,
				LastLogin: lastLogin,
			}).Error
		}
	})

	return db, err
}

// Query demonstrates full ui.TCollate feature set over an in-memory DB
func Query(ctx *ui.Context) string {
	database, err := initDB()
	if err != nil {
		return ui.Div("text-red-700 font-semibold bg-red-50 p-3 rounded border border-red-200")(fmt.Sprintf("DB error: %v", err))
	}

	// Define a unique target id for this page's result area and filter panel
	target := ui.Attr{ID: "query-target"}
	filterTarget := ui.Attr{ID: "query-filter"}

	// Options for SELECT filter
	statusOptions := []ui.AOption{
		{ID: "", Value: "All"},
		{ID: "new", Value: "New"},
		{ID: "active", Value: "Active"},
		{ID: "blocked", Value: "Blocked"},
	}

	// Build collate instance
	c := &ui.TCollate[Person]{
		Init: &ui.TQuery{
			Limit:  12,
			Offset: 0,
			Order:  "name asc",
		},
		Limit:        12,
		Target:       target,
		TargetFilter: filterTarget,
		Database:     database,
		// Search across multiple fields (diacritic-insensitive via normalize())
		Search: []ui.TField{
			{DB: "name", Field: "Name", Text: "Name"},
			{DB: "email", Field: "Email", Text: "Email"},
			{DB: "country", Field: "Country", Text: "Country"},
			{DB: "status", Field: "Status", Text: "Status"},
		},
		// Sort by typical fields
		Sort: []ui.TField{
			{DB: "name", Field: "Name", Text: "Name"},
			{DB: "email", Field: "Email", Text: "Email"},
			{DB: "created_at", Field: "CreatedAt", Text: "Created"},
			{DB: "last_login", Field: "LastLogin", Text: "Last login"},
		},
		// Filters exercising all supported kinds
		Filter: []ui.TField{
			// BOOL: is active
			{DB: "active", Field: "Active", Text: "Active", As: ui.BOOL, Bool: false},
			// ZERO_DATE: last_login is zero
			{DB: "last_login", Field: "LastLogin", Text: "Never logged in", As: ui.ZERO_DATE, Bool: false},
			// NOT_ZERO_DATE: has logged in at least once
			{DB: "last_login", Field: "LastLogin", Text: "Has logged in", As: ui.NOT_ZERO_DATE, Bool: false},
			// DATES: created between
			{DB: "created_at", Field: "CreatedAt", Text: "Created between", As: ui.DATES},
			// SELECT: account status
			{DB: "status", Field: "Status", Text: "Status", As: ui.SELECT, Options: statusOptions},
		},
		// Excel export columns
		Excel: []ui.TField{
			{Field: "Name", Text: "Name"},
			{Field: "Email", Text: "Email"},
			{Field: "Country", Text: "Country"},
			{Field: "Status", Text: "Status"},
			{Field: "Active", Text: "Active"},
			{Field: "CreatedAt", Text: "Created"},
			{Field: "LastLogin", Text: "Last login"},
		},
		OnRow: func(p *Person, _ int) string {
			// Build a simple row card for each record
			badges := []string{}
			if p.Active {
				badges = append(badges, ui.Span("px-2 py-0.5 rounded text-xs bg-green-100 text-green-700 border border-green-200")("active"))
			} else {
				badges = append(badges, ui.Span("px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-700 border border-gray-200")("inactive"))
			}
			badges = append(badges, ui.Span("px-2 py-0.5 rounded text-xs bg-blue-100 text-blue-700 border border-blue-200")(p.Status))

			last := "—"
			if !p.LastLogin.IsZero() {
				last = p.LastLogin.Format("2006-01-02")
			}

			header := ui.Div("flex items-center justify-between")(
				ui.Div("font-semibold")(
					p.Name,
					ui.Span("ml-2 text-gray-500 text-sm")(fmt.Sprintf("<%s>", p.Email)),
				),
				ui.Div("flex gap-1")(strings.Join(badges, "")),
			)

			meta := ui.Div("text-sm text-gray-600 mt-1")(
				fmt.Sprintf("Country: %s | Created: %s | Last login: %s", p.Country, p.CreatedAt.Format("2006-01-02"), last),
			)

			return ui.Div("bg-white rounded-lg border border-gray-200 shadow-sm p-3")(header + meta)
		},
	}

	// Render the collate UI with search, sort, filters, paging and XLS export
	body := ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-3")( // wrapper
		ui.Div("text-3xl font-bold")("Query Demo"),
		ui.Div("text-gray-600")("In-memory SQLite with 100 seeded records. Supports search, sort, filters, paging, and XLS export."),
		c.Render(ctx, &ui.TQuery{}),
	)

	return body
}

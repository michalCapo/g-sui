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
	Surname   string
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

var firstNames = []string{
	"Alexander", "Benjamin", "Christopher", "Daniel", "Edward", "Frederick", "Gabriel", "Harrison", "Isabella", "Jonathan",
	"Katherine", "Leonardo", "Margaret", "Nathaniel", "Olivia", "Patricia", "Quentin", "Rebecca", "Sebastian", "Theodore",
	"Victoria", "William", "Xavier", "Yvonne", "Zachary", "Amelia", "Brandon", "Charlotte", "Dominic", "Eleanor",
}

var lastNames = []string{
	"Anderson", "Brown", "Carter", "Davis", "Evans", "Fisher", "Garcia", "Harris", "Johnson", "King",
	"Lewis", "Miller", "Nelson", "O'Connor", "Parker", "Quinn", "Roberts", "Smith", "Taylor", "Underwood",
	"Valdez", "Wilson", "Xavier", "Young", "Zhang", "Adams", "Bell", "Cooper", "Dixon", "Edwards",
}

var (
	countries = []string{"Slovakia", "Czechia", "Austria", "Poland", "Hungary", "Germany", "Spain", "France", "Italy", "Portugal"}
	statuses  = []string{"new", "active", "blocked"}
)

func randomEmail(name string) string {
	// Sanitize name for email: lowercase, replace spaces with dots, remove special chars
	emailName := strings.ToLower(strings.ReplaceAll(name, " ", "."))
	emailName = strings.ReplaceAll(emailName, "'", "")
	return fmt.Sprintf("%s@example.com", emailName)
}

func randomCountry() string {
	return countries[rand.Intn(len(countries))]
}

func randomStatus() string {
	return statuses[rand.Intn(len(statuses))]
}

func randomPerson() *Person {
	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]

	lastLogin := time.Time{}
	if rand.Intn(2) == 1 {
		lastLogin = time.Now().Add(-time.Duration(rand.Intn(100)) * 24 * time.Hour)
	}

	return &Person{
		Name:      firstName,
		Surname:   lastName,
		Email:     randomEmail(fmt.Sprintf("%s %s", firstName, lastName)),
		Country:   randomCountry(),
		Status:    randomStatus(),
		Active:    rand.Intn(2) == 1,
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(200)) * 24 * time.Hour),
		LastLogin: lastLogin,
	}
}

// initDB initializes an in-memory SQLite DB with 100 seeded records
func initDB() (*gorm.DB, error) {
	var err error

	collateOnce.Do(func() {
		// Use shared in-memory so the DB persists across connections
		dsn := "file:querydemo?mode=memory&cache=shared&_busy_timeout=5000"
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		// Enable custom normalize() SQLite function for better search
		err = ui.RegisterSQLiteNormalize(db)
		if err != nil {
			panic(err)
		}

		// Auto-migrate schema
		if e := db.AutoMigrate(&Person{}); e != nil {
			err = e
			panic(err)
		}

		// Seed only if empty
		var count int64
		err = db.Model(&Person{}).Count(&count).Error
		if err != nil {
			panic(err)
		}

		if count > 0 {
			return
		}

		for i := 1; i <= 100; i++ {
			err = db.Create(randomPerson()).Error
		}
	})

	return db, err
}

func Collate(ctx *ui.Context) string {
	database, err := initDB()
	if err != nil {
		return ui.Div("text-red-700 font-semibold bg-red-50 p-3 rounded border border-red-200")(fmt.Sprintf("DB error: %v", err))
	}

	// Fields
	name := ui.TField{DB: "name", Field: "Name", Text: "Name"}
	email := ui.TField{DB: "email", Field: "Email", Text: "Email"}
	surname := ui.TField{DB: "surname", Field: "Surname", Text: "Surname", As: ui.SELECT, Options: ui.MakeOptions(lastNames)}
	active := ui.TField{DB: "active", Field: "Active", Text: "Active", As: ui.BOOL, Bool: false}
	lastLogin := ui.TField{DB: "last_login", Field: "LastLogin", Text: "Has logged in", As: ui.NOT_ZERO_DATE, Bool: false}
	neverLoggedIn := ui.TField{DB: "last_login", Field: "LastLogin", Text: "Never logged in", As: ui.ZERO_DATE, Bool: false}
	createdAt := ui.TField{DB: "created_at", Field: "CreatedAt", Text: "Created between", As: ui.DATES}
	status := ui.TField{DB: "status", Field: "Status", Text: "Status", As: ui.SELECT, Options: ui.MakeOptions(statuses)}
	country := ui.TField{DB: "country", Field: "Country", Text: "Country", As: ui.SELECT, Options: ui.MakeOptions(countries)}

	// Init
	init := &ui.TQuery{
		Limit: 8,
		Order: "surname asc",
	}

	// Collate
	collate := ui.Collate[Person](init)
	collate.Search(
		surname,
		name,
		email,
		country,
		status,
	)
	collate.Sort(
		surname,
		email,
		lastLogin,
	)
	collate.Filter(
		active,
		lastLogin,
		neverLoggedIn,
		createdAt,
	)
	collate.Excel(
		surname,
		name,
		email,
		country,
		status,
		active,
		createdAt,
		lastLogin,
	)
	collate.Row(func(p *Person, _ int) string {
		// Build a simple row card for each record
		badges := []string{}
		if p.Active {
			badges = append(badges, ui.Span("w-20 text-center px-2 py-0.5 rounded text-xs bg-green-100 text-green-700 border border-green-200")("active"))
		} else {
			badges = append(badges, ui.Span("w-20 text-center px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-700 border border-gray-200")("inactive"))
		}
		badges = append(badges, ui.Span("w-20 text-center px-2 py-0.5 rounded text-xs bg-blue-100 text-blue-700 border border-blue-200")(p.Status))

		last := "—"
		if !p.LastLogin.IsZero() {
			last = p.LastLogin.Format("2006-01-02")
		}

		header := ui.Div("flex items-center justify-between")(
			ui.Div("font-semibold")(
				fmt.Sprintf("%s %s", p.Surname, p.Name),
			),
			ui.Div("flex-1 ml-2 text-gray-500 text-sm")(p.Email),
			ui.Div("flex gap-1")(badges...),
		)

		meta := ui.Div("text-sm text-gray-600 mt-1")(
			fmt.Sprintf("Country: %s | Created: %s | Last login: %s", p.Country, p.CreatedAt.Format("2006-01-02"), last),
		)

		return ui.Div("bg-white rounded-lg border border-gray-200 shadow-sm p-3")(header + meta)
	})

	// Render the collate UI with search, sort, filters, paging and XLS export
	body := ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-3")( // wrapper
		ui.Div("text-3xl font-bold")("Collate Demo"),
		ui.Div("text-gray-600")("In-memory SQLite with 100 seeded records. Supports search, sort, filters, paging, and XLS export."),
		collate.Render(ctx, database),
	)

	return body
}

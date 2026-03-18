package pages

import (
	"fmt"
	"sort"
	"strings"

	r "github.com/michalCapo/g-sui/ui"
)

// Employee is the sample type for the Collate demo.
type Employee struct {
	ID         int
	Name       string
	Department string
	Salary     float64
	HireDate   string
	Active     bool
	Role       string
}

var allEmployees = []*Employee{
	{ID: 1, Name: "Anna Horváthová", Department: "Engineering", Salary: 4200, HireDate: "2026-01-15", Active: true, Role: "Senior Developer"},
	{ID: 2, Name: "Marek Kováč", Department: "Engineering", Salary: 3600, HireDate: "2026-02-10", Active: true, Role: "Developer"},
	{ID: 3, Name: "Jana Novotná", Department: "Marketing", Salary: 3100, HireDate: "2025-06-20", Active: true, Role: "Marketing Manager"},
	{ID: 4, Name: "Peter Baláž", Department: "Engineering", Salary: 4800, HireDate: "2025-09-01", Active: true, Role: "Tech Lead"},
	{ID: 5, Name: "Eva Šimková", Department: "HR", Salary: 2800, HireDate: "2026-03-05", Active: true, Role: "HR Specialist"},
	{ID: 6, Name: "Tomáš Krajčí", Department: "Sales", Salary: 3400, HireDate: "2025-11-05", Active: false, Role: "Sales Rep"},
	{ID: 7, Name: "Lucia Molnárová", Department: "Engineering", Salary: 3900, HireDate: "2025-08-22", Active: true, Role: "Developer"},
	{ID: 8, Name: "Martin Černý", Department: "Marketing", Salary: 2900, HireDate: "2026-01-08", Active: true, Role: "Content Writer"},
	{ID: 9, Name: "Katarína Vargová", Department: "Engineering", Salary: 5200, HireDate: "2025-05-12", Active: true, Role: "Architect"},
	{ID: 10, Name: "Jakub Pokorný", Department: "Sales", Salary: 3200, HireDate: "2025-07-01", Active: true, Role: "Sales Rep"},
	{ID: 11, Name: "Zuzana Rybárová", Department: "HR", Salary: 3500, HireDate: "2025-12-10", Active: true, Role: "HR Manager"},
	{ID: 12, Name: "Daniel Vlček", Department: "Engineering", Salary: 4100, HireDate: "2025-02-28", Active: false, Role: "Senior Developer"},
	{ID: 13, Name: "Michaela Tóthová", Department: "Marketing", Salary: 3300, HireDate: "2026-03-15", Active: true, Role: "Designer"},
	{ID: 14, Name: "Ondrej Hudák", Department: "Sales", Salary: 3700, HireDate: "2025-04-18", Active: true, Role: "Sales Manager"},
	{ID: 15, Name: "Barbora Kučerová", Department: "Engineering", Salary: 3800, HireDate: "2026-02-14", Active: true, Role: "Developer"},
	{ID: 16, Name: "Štefan Mészáros", Department: "HR", Salary: 2600, HireDate: "2026-03-01", Active: true, Role: "Recruiter"},
	{ID: 17, Name: "Natália Szabóová", Department: "Engineering", Salary: 4500, HireDate: "2025-07-20", Active: true, Role: "DevOps Engineer"},
	{ID: 18, Name: "Richard Páleník", Department: "Sales", Salary: 3000, HireDate: "2025-11-12", Active: false, Role: "Sales Rep"},
	{ID: 19, Name: "Lenka Fiľová", Department: "Marketing", Salary: 3200, HireDate: "2025-10-25", Active: true, Role: "SEO Specialist"},
	{ID: 20, Name: "Adam Gregor", Department: "Engineering", Salary: 4000, HireDate: "2026-01-03", Active: true, Role: "Developer"},
}

// collateFilters stores active filters (global for demo, per-session in real app)
var collateFilters = map[string]*r.CollateFilterValue{}

func CollatePage(ctx *r.Context) *r.Node {
	limit := 8
	data := allEmployees
	if len(data) > limit {
		data = data[:limit]
	}

	collate := newCollate().
		Page(1).TotalItems(len(allEmployees)).HasMore(len(allEmployees) > limit).
		Render(data)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Collate"),
		r.Div("text-gray-600 dark:text-gray-400").Text(
			"Data component with filter/sort panel, search, load-more pagination, and export. "+
				"Click the Filter button to open the sort & filter panel.",
		),
		r.Div("bg-white dark:bg-gray-900 rounded-xl shadow p-6 border border-gray-200 dark:border-gray-800 relative").Render(
			collate,
		),
	)
}

func newCollate() *r.Collate[Employee] {
	return r.NewCollate[Employee]("employees-collate").
		Action("collate.data").
		Limit(8).
		Sort(
			r.CollateSortField{Field: "name", Label: "Meno"},
			r.CollateSortField{Field: "department", Label: "Oddelenie"},
			r.CollateSortField{Field: "salary", Label: "Plat"},
			r.CollateSortField{Field: "hire_date", Label: "Nástup"},
		).
		Filter(
			r.CollateFilterField{Field: "active", Label: "Iba aktívni", Type: r.CollateBool},
			r.CollateFilterField{Field: "hire_date", Label: "Dátum nástupu", Type: r.CollateDateRange},
			r.CollateFilterField{
				Field: "department",
				Label: "Oddelenie",
				Type:  r.CollateSelect,
				Options: []r.CollateOption{
					{Value: "Engineering", Label: "Engineering"},
					{Value: "Marketing", Label: "Marketing"},
					{Value: "Sales", Label: "Sales"},
					{Value: "HR", Label: "HR"},
				},
			},
		).
		Row(renderEmployeeRow).
		Empty("Žiadni zamestnanci").
		EmptyIcon("group_off")
}

func renderEmployeeRow(emp *Employee, idx int) *r.Node {
	stripeCls := "p-4 border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors"
	if idx%2 == 1 {
		stripeCls = "p-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/20 hover:bg-gray-100/60 dark:hover:bg-gray-800/40 transition-colors"
	}

	statusColor := "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400"
	statusText := "Aktívny"
	if !emp.Active {
		statusColor = "bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
		statusText = "Neaktívny"
	}

	return r.Div(stripeCls).Render(
		r.Div("flex items-center justify-between").Render(
			// Left: avatar + name + role
			r.Div("flex items-center gap-3").Render(
				// Avatar circle with initials
				r.Div("w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center "+
					"text-sm font-bold text-gray-600 dark:text-gray-300").
					Text(initials(emp.Name)),
				r.Div("flex flex-col").Render(
					r.Span("font-medium text-gray-900 dark:text-gray-100").Text(emp.Name),
					r.Span("text-xs text-gray-500 dark:text-gray-400").Text(emp.Role),
				),
			),
			// Right: department + salary + status + date (fixed widths)
			r.Div("flex items-center gap-4").Render(
				r.Span("text-sm text-gray-600 dark:text-gray-400 w-24 text-right").Text(emp.Department),
				r.Span("text-sm font-medium text-gray-800 dark:text-gray-200 w-16 text-right").
					Text(fmt.Sprintf("€%.0f", emp.Salary)),
				r.Span("text-xs px-2 py-0.5 rounded-full font-medium w-20 text-center "+statusColor).
					Text(statusText),
				r.Span("text-xs text-gray-400 dark:text-gray-500 w-20 text-right").Text(emp.HireDate),
			),
		),
	)
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	result := string([]rune(parts[0])[:1])
	if len(parts) > 1 {
		result += string([]rune(parts[len(parts)-1])[:1])
	}
	return strings.ToUpper(result)
}

// ---------------------------------------------------------------------------
// Collate data request
// ---------------------------------------------------------------------------

type CollateDataRequest struct {
	Operation string                    `json:"operation"`
	Search    string                    `json:"search"`
	Page      int                       `json:"page"`
	Limit     int                       `json:"limit"`
	Order     string                    `json:"order"`
	Filters   []r.CollateFilterValue    `json:"filters"`
}

func handleCollateData(ctx *r.Context) string {
	var req CollateDataRequest
	ctx.Body(&req)

	limit := 8
	if req.Limit > 0 {
		limit = req.Limit
	}

	// Handle filter/reset: update active filters
	switch req.Operation {
	case "filter":
		collateFilters = map[string]*r.CollateFilterValue{}
		for i := range req.Filters {
			f := &req.Filters[i]
			collateFilters[f.Field] = f
		}
	case "reset":
		collateFilters = map[string]*r.CollateFilterValue{}
		req.Search = ""
		req.Order = ""
	}

	// Apply filters + search
	filtered := filterEmployees(req.Search, collateFilters)

	// Apply sort
	sortEmployees(filtered, req.Order)

	totalItems := len(filtered)

	// Handle export
	if req.Operation == "export" {
		return fmt.Sprintf("console.log('Exporting %d employees');", totalItems)
	}

	// Handle load more
	if req.Operation == "loadmore" {
		if req.Page < 1 {
			req.Page = 1
		}
		start := (req.Page - 1) * limit
		end := start + limit
		if end > totalItems {
			end = totalItems
		}
		if start >= totalItems {
			return ""
		}
		pageData := filtered[start:end]
		hasMore := end < totalItems

		dt := newCollateWithState(req.Search, req.Order).
			Page(req.Page).TotalItems(totalItems).HasMore(hasMore).
			RowOffset(start)

		resp := r.NewResponse()
		rows := dt.RenderRows(pageData)
		for _, row := range rows {
			resp.Append(dt.BodyID(), row)
		}
		resp.Replace(dt.FooterID(), dt.RenderFooter())
		return resp.Build()
	}

	// Default: full re-render (search, sort, filter)
	if req.Page < 1 {
		req.Page = 1
	}
	end := req.Page * limit
	if end > totalItems {
		end = totalItems
	}
	pageData := filtered[:end]
	hasMore := end < totalItems

	collateNode := newCollateWithState(req.Search, req.Order).
		Page(req.Page).TotalItems(totalItems).HasMore(hasMore).
		Render(pageData)

	return collateNode.ToJSReplace("employees-collate")
}

func newCollateWithState(search, order string) *r.Collate[Employee] {
	c := newCollate().Search(search).Order(order)

	for field, val := range collateFilters {
		c.SetFilter(field, val)
	}

	return c
}

func filterEmployees(search string, filters map[string]*r.CollateFilterValue) []*Employee {
	result := make([]*Employee, 0, len(allEmployees))
	for _, emp := range allEmployees {
		// Text search
		if search != "" {
			s := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(emp.Name), s) &&
				!strings.Contains(strings.ToLower(emp.Department), s) &&
				!strings.Contains(strings.ToLower(emp.Role), s) {
				continue
			}
		}

		// Apply filters
		if !applyEmployeeFilters(emp, filters) {
			continue
		}

		result = append(result, emp)
	}
	return result
}

func applyEmployeeFilters(emp *Employee, filters map[string]*r.CollateFilterValue) bool {
	for _, f := range filters {
		switch f.Type {
		case "bool":
			if f.Field == "active" && f.Bool && !emp.Active {
				return false
			}
		case "date":
			if f.Field == "hire_date" {
				if f.From != "" && emp.HireDate < f.From {
					return false
				}
				if f.To != "" && emp.HireDate > f.To {
					return false
				}
			}
		case "select":
			if f.Field == "department" && f.Value != "" {
				if emp.Department != f.Value {
					return false
				}
			}
		}
	}
	return true
}

func sortEmployees(data []*Employee, order string) {
	if order == "" {
		return
	}
	parts := strings.Fields(strings.TrimSpace(order))
	if len(parts) < 2 {
		return
	}
	field := strings.ToLower(parts[0])
	dir := strings.ToLower(parts[1])

	sort.Slice(data, func(i, j int) bool {
		var cmp int
		switch field {
		case "name":
			cmp = strings.Compare(data[i].Name, data[j].Name)
		case "department":
			cmp = strings.Compare(data[i].Department, data[j].Department)
		case "salary":
			if data[i].Salary < data[j].Salary {
				cmp = -1
			} else if data[i].Salary > data[j].Salary {
				cmp = 1
			}
		case "hire_date":
			cmp = strings.Compare(data[i].HireDate, data[j].HireDate)
		default:
			return false
		}
		if dir == "desc" {
			return cmp > 0
		}
		return cmp < 0
	})
}

func RegisterCollate(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/collate", func(ctx *r.Context) *r.Node { return layout(CollatePage(ctx)) })
	app.Action("nav.collate", NavTo("/collate", func() *r.Node { return CollatePage(nil) }))
	app.Action("collate.data", handleCollateData)
}

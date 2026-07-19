package pages

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	r "github.com/michalCapo/g-sui/ui"
)

// ContentID is the shared target ID for the main content area.
var ContentID = r.Target()

func RoutesExample(ctx *r.Context) *r.Node {
	codeSnippet := func(text string) *r.Node {
		return r.Code("bg-gray-100 px-1 rounded text-sm").Text(text)
	}

	return r.Div("flex flex-col gap-8").Render(
		r.Div("text-3xl font-bold").Text("Route Parameters"),
		r.Div("text-gray-600").Text("Demonstrates real parameterized app.Page routes."),

		// Overview
		r.Div("bg-white rounded-lg shadow p-6").Render(
			r.Div("font-bold text-lg mb-2").Text("Overview"),
			r.Div("flex flex-col gap-2 text-sm text-gray-600").Render(
				r.Div("flex items-center gap-2").Render(
					r.Span().Text("Routes use curly braces: "),
					codeSnippet("/user/{id}"),
				),
				r.Div("flex items-center gap-2").Render(
					r.Span().Text("Multiple params: "),
					codeSnippet("/user/{userId}/post/{postId}"),
				),
				r.Div("flex items-center gap-2").Render(
					r.Span().Text("Access path params via: "),
					codeSnippet("ctx.Request.PathValue(\"id\")"),
				),
				r.Div("flex items-center gap-2").Render(
					r.Span().Text("Query params via: "),
					codeSnippet("ctx.Query[\"name\"]"),
				),
			),
		),

		// Single parameter
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-2xl font-bold").Text("Single Parameter Routes"),
			r.Div("bg-white rounded-lg shadow p-6").Render(
				r.Div("text-sm text-gray-600 mb-3").Text("Click to view user details:"),
				r.Div("flex flex-wrap gap-2").Render(
					routeLink("View User 123", "/routes/user/123", "bg-blue-600 text-white hover:bg-blue-700"),
					routeLink("View User 456", "/routes/user/456", "bg-blue-600 text-white hover:bg-blue-700"),
					routeLink("View User alice", "/routes/user/alice", "bg-blue-600 text-white hover:bg-blue-700"),
				),
				r.Div("text-xs text-gray-500 mt-2").Render(
					r.Span().Text("Route pattern: "),
					codeSnippet("/routes/user/{id}"),
				),
			),
		),

		// Multiple parameters
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-2xl font-bold").Text("Multiple Parameter Routes"),
			r.Div("bg-white rounded-lg shadow p-6").Render(
				r.Div("text-sm text-gray-600 mb-3").Text("Navigate to routes with multiple parameters:"),
				r.Div("flex flex-wrap gap-2").Render(
					routeLink("User 123, Post 1", "/routes/user/123/post/1", "bg-green-600 text-white hover:bg-green-700"),
					routeLink("User 456, Post 42", "/routes/user/456/post/42", "bg-green-600 text-white hover:bg-green-700"),
					routeLink("User alice, Post my-first-post", "/routes/user/alice/post/my-first-post", "bg-green-600 text-white hover:bg-green-700"),
				),
				r.Div("text-xs text-gray-500 mt-2").Render(
					r.Span().Text("Route pattern: "),
					codeSnippet("/routes/user/{userId}/post/{postId}"),
				),
			),
		),

		// Nested routes
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-2xl font-bold").Text("Nested Routes"),
			r.Div("bg-white rounded-lg shadow p-6").Render(
				r.Div("text-sm text-gray-600 mb-3").Text("Routes can have parameters at any level:"),
				r.Div("flex flex-wrap gap-2").Render(
					routeLink("Electronics > Laptop", "/routes/category/electronics/product/laptop", "bg-purple-600 text-white hover:bg-purple-700"),
					routeLink("Books > Novel", "/routes/category/books/product/novel", "bg-purple-600 text-white hover:bg-purple-700"),
				),
				r.Div("text-xs text-gray-500 mt-2").Render(
					r.Span().Text("Route pattern: "),
					codeSnippet("/routes/category/{category}/product/{product}"),
				),
			),
		),

		// Query parameters
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-2xl font-bold").Text("Query Parameters"),
			r.Div("bg-white rounded-lg shadow p-6").Render(
				r.Div("text-sm text-gray-600 mb-3").Text("Query parameters are passed after a ? in the URL:"),
				r.Div("flex flex-wrap gap-2").Render(
					routeLink("name=Smith, age=30", "/routes/search?name=Smith&age=30", "bg-yellow-600 text-white hover:bg-yellow-700"),
					routeLink("name=Johnson, city=NYC", "/routes/search?name=Johnson&city=NYC", "bg-yellow-600 text-white hover:bg-yellow-700"),
					routeLink("q=g-sui, type=tutorial", "/routes/search?q=g-sui&type=tutorial", "bg-yellow-600 text-white hover:bg-yellow-700"),
				),
				r.Div("text-xs text-gray-500 mt-2").Render(
					r.Span().Text("Accessed via: "),
					codeSnippet("ctx.Query[\"name\"]"),
				),
			),
		),

		// Combined path + query parameters
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-2xl font-bold").Text("Combined Path + Query Parameters"),
			r.Div("bg-white rounded-lg shadow p-6").Render(
				r.Div("text-sm text-gray-600 mb-3").Text("Combine path parameters with query parameters:"),
				r.Div("flex flex-wrap gap-2").Render(
					routeLink("User 123: tab=profile", "/routes/user/123?tab=profile&view=detailed", "bg-indigo-600 text-white hover:bg-indigo-700"),
					routeLink("User 456: tab=settings", "/routes/user/456?tab=settings", "bg-indigo-600 text-white hover:bg-indigo-700"),
					routeLink("User alice: sort=name", "/routes/user/alice?sort=name&order=asc", "bg-indigo-600 text-white hover:bg-indigo-700"),
				),
				r.Div("text-xs text-gray-500 mt-2").Render(
					r.Span().Text("Path values and ctx.Query are populated from the URL"),
				),
			),
		),
	)
}

func routeLink(label, href, cls string) *r.Node {
	return r.A("inline-block px-4 py-2 rounded text-sm "+cls).
		Attr("href", href).
		Text(label)
}

func routesUserDetail(id string, query map[string]string) *r.Node {
	users := map[string]map[string]string{
		"123":   {"name": "John Doe", "email": "john@example.com", "role": "Admin"},
		"456":   {"name": "Jane Smith", "email": "jane@example.com", "role": "User"},
		"alice": {"name": "Alice Johnson", "email": "alice@example.com", "role": "Moderator"},
	}

	user, exists := users[id]
	name := "Unknown User"
	email := "N/A"
	role := "Guest"
	if exists {
		name = user["name"]
		email = user["email"]
		role = user["role"]
	}

	tab := query["tab"]
	view := query["view"]
	sort := query["sort"]
	order := query["order"]

	var querySection *r.Node
	if tab != "" || view != "" || sort != "" || order != "" {
		params := make([]*r.Node, 0)
		if tab != "" {
			params = append(params, paramBadge("tab", tab, "bg-yellow-100 text-yellow-700"))
		}
		if view != "" {
			params = append(params, paramBadge("view", view, "bg-yellow-100 text-yellow-700"))
		}
		if sort != "" {
			params = append(params, paramBadge("sort", sort, "bg-yellow-100 text-yellow-700"))
		}
		if order != "" {
			params = append(params, paramBadge("order", order, "bg-yellow-100 text-yellow-700"))
		}
		querySection = r.Div("flex flex-col gap-2 mt-4").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase").Text("Query Parameters"),
			r.Div("flex flex-wrap gap-2").Render(params...),
		)
	}

	detail := r.Div("flex flex-col gap-6").Render(
		r.Div("flex items-center gap-4").Render(
			r.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm").
				Attr("href", "/routes").Text("Back"),
			r.Div("text-2xl font-bold").Text("User: "+name),
		),
		r.Div("bg-white rounded-lg shadow p-6").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase mb-2").Text("Route Parameter"),
			r.Div("text-lg font-mono bg-gray-100 px-3 py-2 rounded mb-4").Text("ID: "+id),
			r.Div("grid grid-cols-2 gap-4").Render(
				infoBox("Name", name),
				infoBox("Email", email),
				infoBox("Role", role),
			),
			querySection,
			r.Div("text-xs text-gray-500 mt-4 p-3 bg-blue-50 rounded").Render(
				r.Strong().Text("Code: "),
				r.Code("bg-white px-1 rounded").Text(`ctx.Request.PathValue("id")`),
				r.Span().Text(" for path params, "),
				r.Code("bg-white px-1 rounded").Text(`ctx.Query["tab"]`),
				r.Span().Text(" for query params"),
			),
		),
	)

	return detail
}

func routesUserPostDetail(userID, postID string) *r.Node {
	posts := map[string]map[string]string{
		"1":             {"title": "Getting Started with g-sui", "content": "This is a comprehensive guide to building server-rendered UIs with g-sui."},
		"42":            {"title": "Advanced Routing Patterns", "content": "Learn how to use parameterized routes effectively in your applications."},
		"my-first-post": {"title": "My First Post", "content": "This is a blog post with a slug-based ID instead of numeric."},
	}

	post, exists := posts[postID]
	if !exists {
		post = map[string]string{
			"title":   "Post Not Found",
			"content": "This post doesn't exist in our mock database.",
		}
	}

	detail := r.Div("flex flex-col gap-6").Render(
		r.Div("flex items-center gap-4").Render(
			r.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm").
				Attr("href", "/routes").Text("Back"),
			r.Div("text-2xl font-bold").Text("Post Details"),
		),
		r.Div("bg-white rounded-lg shadow p-6 flex flex-col gap-4").Render(
			r.Div("flex flex-col gap-2").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase").Text("Route Parameters"),
				r.Div("grid grid-cols-2 gap-2").Render(
					r.Div("text-sm font-mono bg-gray-100 px-3 py-2 rounded").Text("userId: "+userID),
					r.Div("text-sm font-mono bg-gray-100 px-3 py-2 rounded").Text("postId: "+postID),
				),
			),
			infoBox("Title", post["title"]),
			infoBox("Content", post["content"]),
			r.Div("flex items-center gap-2").Render(
				r.Div("text-sm font-bold text-gray-500").Text("Author:"),
				r.A("px-3 py-1 border-2 border-blue-600 text-blue-600 rounded text-sm").
					Attr("href", "/routes/user/"+url.PathEscape(userID)).
					Text("User "+userID),
			),
		),
	)

	return detail
}

func routesProductDetail(category, product string) *r.Node {
	products := map[string]map[string]map[string]string{
		"electronics": {
			"laptop": {"name": "Gaming Laptop", "price": "$1,299", "description": "High-performance gaming laptop with RTX graphics."},
		},
		"books": {
			"novel": {"name": "The Great Novel", "price": "$19.99", "description": "A captivating story that will keep you reading all night."},
		},
	}

	var prod map[string]string
	if cat, ok := products[category]; ok {
		if p, ok := cat[product]; ok {
			prod = p
		}
	}
	if prod == nil {
		prod = map[string]string{"name": "Not Found", "price": "N/A", "description": "This product doesn't exist in our catalog."}
	}

	detail := r.Div("flex flex-col gap-6").Render(
		r.Div("flex items-center gap-4").Render(
			r.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm").
				Attr("href", "/routes").Text("Back"),
			r.Div("text-2xl font-bold").Text("Product Details"),
		),
		r.Div("bg-white rounded-lg shadow p-6 flex flex-col gap-4").Render(
			r.Div("flex flex-col gap-2").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase").Text("Route Parameters"),
				r.Div("grid grid-cols-2 gap-2").Render(
					r.Div("text-sm font-mono bg-gray-100 px-3 py-2 rounded").Text("category: "+category),
					r.Div("text-sm font-mono bg-gray-100 px-3 py-2 rounded").Text("product: "+product),
				),
			),
			infoBox("Product Name", prod["name"]),
			r.Div("flex flex-col gap-1").Render(
				r.Div("text-sm font-bold text-gray-500").Text("Price"),
				r.Div("text-2xl font-bold text-green-600").Text(prod["price"]),
			),
			infoBox("Description", prod["description"]),
		),
	)

	return detail
}

func routesSearchDetail(query map[string]string) *r.Node {
	// Build extracted params display
	paramNames := []string{"name", "age", "city", "q", "type"}
	params := make([]*r.Node, 0)
	for _, key := range paramNames {
		if val := query[key]; val != "" {
			params = append(params, paramBadge(key, val, "bg-gray-100 text-gray-700"))
		}
	}

	// Build all params display
	var allParts []string
	for key, val := range query {
		if val != "" {
			allParts = append(allParts, fmt.Sprintf("%s=%s", key, val))
		}
	}
	sort.Strings(allParts)
	allParamsText := "No query parameters"
	if len(allParts) > 0 {
		allParamsText = strings.Join(allParts, ", ")
	}

	detail := r.Div("flex flex-col gap-6").Render(
		r.Div("flex items-center gap-4").Render(
			r.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 text-sm").
				Attr("href", "/routes").Text("Back"),
			r.Div("text-2xl font-bold").Text("Search Results"),
		),
		r.Div("bg-white rounded-lg shadow p-6 flex flex-col gap-4").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase").Text("Extracted Query Parameters"),
			r.Div("flex flex-wrap gap-2").Render(params...),
			r.Div("text-sm font-bold text-gray-500 mt-2").Text("All Parameters"),
			r.Div("text-xs font-mono bg-gray-100 px-3 py-2 rounded").Text(allParamsText),
			r.Div("text-xs text-gray-500 mt-4 p-3 bg-yellow-50 rounded").Render(
				r.Strong().Text("Code: "),
				r.Code("bg-white px-1 rounded").Text(`ctx.Query["name"]`),
			),
		),
		r.Div("flex flex-col gap-2").Render(
			r.Div("text-sm font-bold").Text("Try different queries:"),
			r.Div("flex flex-wrap gap-2").Render(
				routeLink("name=Smith, age=30", "/routes/search?name=Smith&age=30", "border-2 border-yellow-600 text-yellow-600 hover:bg-yellow-50"),
				routeLink("name=Johnson, city=NYC", "/routes/search?name=Johnson&city=NYC", "border-2 border-yellow-600 text-yellow-600 hover:bg-yellow-50"),
			),
		),
	)

	return detail
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func infoBox(label, value string) *r.Node {
	return r.Div("flex flex-col gap-1").Render(
		r.Div("text-sm font-bold text-gray-500").Text(label),
		r.Div("text-sm").Text(value),
	)
}

func paramBadge(key, value, cls string) *r.Node {
	return r.Div("text-sm font-mono px-3 py-2 rounded " + cls).
		Text(fmt.Sprintf("%s: %s", key, value))
}

// NavTo creates a navigation action handler that replaces the content area
// and updates the browser URL.
func NavTo(url string, content func() *r.Node) r.ActionHandler {
	return func(ctx *r.Context) string {
		return r.NewResponse().
			Inner(ContentID, content()).
			Navigate(url).
			Build()
	}
}

func RegisterRoutes(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/routes", func(ctx *r.Context) *r.Node { return layout(ctx, RoutesExample(ctx)) })
	app.Page("/routes/user/{id}", func(ctx *r.Context) *r.Node {
		return layout(ctx, routesUserDetail(ctx.Request.PathValue("id"), ctx.Query))
	})
	app.Page("/routes/user/{userId}/post/{postId}", func(ctx *r.Context) *r.Node {
		return layout(ctx, routesUserPostDetail(
			ctx.Request.PathValue("userId"),
			ctx.Request.PathValue("postId"),
		))
	})
	app.Page("/routes/category/{category}/product/{product}", func(ctx *r.Context) *r.Node {
		return layout(ctx, routesProductDetail(
			ctx.Request.PathValue("category"),
			ctx.Request.PathValue("product"),
		))
	})
	app.Page("/routes/search", func(ctx *r.Context) *r.Node {
		return layout(ctx, routesSearchDetail(ctx.Query))
	})
	app.Action("nav.routes", NavTo("/routes", func() *r.Node { return RoutesExample(nil) }))
}

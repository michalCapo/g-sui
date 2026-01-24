package pages

import (
	"github.com/michalCapo/g-sui/ui"
)

// RoutesExample showcases parameterized routes and route parameters
func RoutesExample(ctx *ui.Context) string {
	return ui.Div("flex flex-col gap-8")(
		ui.Div("text-3xl font-bold")("Route Parameters"),
		ui.Div("text-gray-600")("Demonstrates how to use parameterized routes with g-sui."),

		// Overview section
		ui.Card().Header("<h3 class='font-bold text-lg'>Overview</h3>").
			Body(ui.Div("flex flex-col gap-2 text-sm text-gray-600 dark:text-gray-400")(
				ui.P("")(`Routes can include parameters using curly braces: <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">/user/{id}</code>`),
				ui.P("")(`Access path parameters in handlers using <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">ctx.PathParam("id")</code>`),
				ui.P("")(`Multiple parameters are supported: <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">/user/{userId}/post/{postId}</code>`),
				ui.P("")(`Query parameters are accessed via <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">ctx.QueryParam("name")</code>`),
				ui.P("")(`Example: <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">/search?name=Smith&age=30</code>`),
			)).Render(),

		// Single parameter examples
		ui.Div("flex flex-col gap-4")(
			ui.Div("text-2xl font-bold")("Single Parameter Routes"),
			ui.Card().Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("text-sm text-gray-600 dark:text-gray-400")(
						"Click the links below to navigate to routes with single parameters:",
					),
					ui.Div("flex flex-wrap gap-2")(
						ui.A("px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 cursor-pointer inline-block", ui.Href("/routes/user/123"), ctx.Load("/routes/user/123"))("View User 123"),
						ui.A("px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 cursor-pointer inline-block", ui.Href("/routes/user/456"), ctx.Load("/routes/user/456"))("View User 456"),
						ui.A("px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 cursor-pointer inline-block", ui.Href("/routes/user/alice"), ctx.Load("/routes/user/alice"))("View User 'alice'"),
					),
					ui.Div("text-xs text-gray-500 mt-2")(
						"Route pattern: <code class='bg-gray-100 dark:bg-gray-800 px-1 rounded'>/routes/user/{id}</code>",
					),
				),
			).Render(),
		),

		// Multiple parameter examples
		ui.Div("flex flex-col gap-4")(
			ui.Div("text-2xl font-bold")("Multiple Parameter Routes"),
			ui.Card().Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("text-sm text-gray-600 dark:text-gray-400")(
						"Navigate to routes with multiple parameters:",
					),
					ui.Div("flex flex-wrap gap-2")(
						ui.A("px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 cursor-pointer inline-block", ui.Href("/routes/user/123/post/1"), ctx.Load("/routes/user/123/post/1"))("User 123, Post 1"),
						ui.A("px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 cursor-pointer inline-block", ui.Href("/routes/user/456/post/42"), ctx.Load("/routes/user/456/post/42"))("User 456, Post 42"),
						ui.A("px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 cursor-pointer inline-block", ui.Href("/routes/user/alice/post/my-first-post"), ctx.Load("/routes/user/alice/post/my-first-post"))("User alice, Post 'my-first-post'"),
					),
					ui.Div("text-xs text-gray-500 mt-2")(
						"Route pattern: <code class='bg-gray-100 dark:bg-gray-800 px-1 rounded'>/routes/user/{userId}/post/{postId}</code>",
					),
				),
			).Render(),
		),

		// Nested routes example
		ui.Div("flex flex-col gap-4")(
			ui.Div("text-2xl font-bold")("Nested Routes"),
			ui.Card().Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("text-sm text-gray-600 dark:text-gray-400")(
						"Routes can have parameters at any level:",
					),
					ui.Div("flex flex-wrap gap-2")(
						ui.A("px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 cursor-pointer inline-block", ui.Href("/routes/category/electronics/product/laptop"), ctx.Load("/routes/category/electronics/product/laptop"))("Electronics → Laptop"),
						ui.A("px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 cursor-pointer inline-block", ui.Href("/routes/category/books/product/novel"), ctx.Load("/routes/category/books/product/novel"))("Books → Novel"),
					),
					ui.Div("text-xs text-gray-500 mt-2")(
						"Route pattern: <code class='bg-gray-100 dark:bg-gray-800 px-1 rounded'>/routes/category/{category}/product/{product}</code>",
					),
				),
			).Render(),
		),

		// Query parameter examples
		ui.Div("flex flex-col gap-4")(
			ui.Div("text-2xl font-bold")("Query Parameters"),
			ui.Card().Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("text-sm text-gray-600 dark:text-gray-400")(
						"Query parameters are passed in the URL after a question mark. They work with any route:",
					),
					ui.Div("flex flex-wrap gap-2")(
						ui.A("px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700 cursor-pointer inline-block", ui.Href("/routes/search?name=Smith&age=30"), ctx.Load("/routes/search?name=Smith&age=30"))("Search: name=Smith, age=30"),
						ui.A("px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700 cursor-pointer inline-block", ui.Href("/routes/search?name=Johnson&city=NYC"), ctx.Load("/routes/search?name=Johnson&city=NYC"))("Search: name=Johnson, city=NYC"),
						ui.A("px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700 cursor-pointer inline-block", ui.Href("/routes/search?q=g-sui&type=tutorial"), ctx.Load("/routes/search?q=g-sui&type=tutorial"))("Search: q=g-sui, type=tutorial"),
					),
					ui.Div("text-xs text-gray-500 mt-2")(
						"Route: <code class='bg-gray-100 dark:bg-gray-800 px-1 rounded'>/routes/search</code> (no path params needed)",
					),
				),
			).Render(),
		),

		// Combined path and query parameters
		ui.Div("flex flex-col gap-4")(
			ui.Div("text-2xl font-bold")("Combined Path + Query Parameters"),
			ui.Card().Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("text-sm text-gray-600 dark:text-gray-400")(
						"You can combine path parameters with query parameters:",
					),
					ui.Div("flex flex-wrap gap-2")(
						ui.A("px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 cursor-pointer inline-block", ui.Href("/routes/user/123?tab=profile&view=detailed"), ctx.Load("/routes/user/123?tab=profile&view=detailed"))("User 123: tab=profile, view=detailed"),
						ui.A("px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 cursor-pointer inline-block", ui.Href("/routes/user/456?tab=settings"), ctx.Load("/routes/user/456?tab=settings"))("User 456: tab=settings"),
						ui.A("px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 cursor-pointer inline-block", ui.Href("/routes/user/alice?sort=name&order=asc"), ctx.Load("/routes/user/alice?sort=name&order=asc"))("User alice: sort=name, order=asc"),
					),
					ui.Div("text-xs text-gray-500 mt-2")(
						"Route pattern: <code class='bg-gray-100 dark:bg-gray-800 px-1 rounded'>/routes/user/{id}</code> + query params",
					),
				),
			).Render(),
		),
	)
}

// SearchExample demonstrates query parameters
func SearchExample(ctx *ui.Context) string {
	// Get query parameters using ctx.QueryParam() - works with SPA navigation
	name := ctx.QueryParam("name")
	age := ctx.QueryParam("age")
	city := ctx.QueryParam("city")
	q := ctx.QueryParam("q")
	searchType := ctx.QueryParam("type")

	// Get all query parameters for display
	allParams := ctx.AllQueryParams()

	return ui.Div("flex flex-col gap-6")(
		ui.Div("flex items-center gap-4")(
			ui.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 cursor-pointer inline-block", ui.Href("/routes"), ctx.Load("/routes"))("← Back"),
			ui.Div("text-2xl font-bold")("Search Results"),
		),

		ui.Card().Header("<h3 class='font-bold text-lg'>Query Parameters</h3>").
			Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500 uppercase")("Extracted Query Parameters"),
						ui.Div("grid grid-cols-1 md:grid-cols-2 gap-2")(
							func() string {
								if name != "" {
									return ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
										"name: " + name,
									)
								}
								return ""
							}(),
							func() string {
								if age != "" {
									return ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
										"age: " + age,
									)
								}
								return ""
							}(),
							func() string {
								if city != "" {
									return ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
										"city: " + city,
									)
								}
								return ""
							}(),
							func() string {
								if q != "" {
									return ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
										"q: " + q,
									)
								}
								return ""
							}(),
							func() string {
								if searchType != "" {
									return ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
										"type: " + searchType,
									)
								}
								return ""
							}(),
						),
					),
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("All Query Parameters"),
						ui.Div("text-xs font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded break-all")(
							func() string {
								if len(allParams) == 0 {
									return "No query parameters provided"
								}
								result := ""
								for key, values := range allParams {
									for _, val := range values {
										if result != "" {
											result += ", "
										}
										result += key + "=" + val
									}
								}
								return result
							}(),
						),
					),
					ui.Div("text-xs text-gray-500 mt-4 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded")(
						`<strong>Code:</strong> <code class="bg-white dark:bg-gray-800 px-1 rounded">name := ctx.QueryParam("name")</code>`,
					),
				),
			).Render(),

		// Example links
		ui.Div("flex flex-col gap-2")(
			ui.Div("text-sm font-bold")("Try different queries:"),
			ui.Div("flex flex-wrap gap-2")(
				ui.A("px-3 py-1 border-2 border-yellow-600 text-yellow-600 rounded hover:bg-yellow-50 cursor-pointer inline-block text-sm", ui.Href("/routes/search?name=Smith&age=30"), ctx.Load("/routes/search?name=Smith&age=30"))("name=Smith, age=30"),
				ui.A("px-3 py-1 border-2 border-yellow-600 text-yellow-600 rounded hover:bg-yellow-50 cursor-pointer inline-block text-sm", ui.Href("/routes/search?name=Johnson&city=NYC"), ctx.Load("/routes/search?name=Johnson&city=NYC"))("name=Johnson, city=NYC"),
				ui.A("px-3 py-1 border-2 border-yellow-600 text-yellow-600 rounded hover:bg-yellow-50 cursor-pointer inline-block text-sm", ui.Href("/routes/search?q=g-sui&type=tutorial"), ctx.Load("/routes/search?q=g-sui&type=tutorial"))("q=g-sui, type=tutorial"),
			),
		),
	)
}

// UserDetail shows a single user by ID
func UserDetail(ctx *ui.Context) string {
	userID := ctx.PathParam("id")
	if userID == "" {
		userID = "unknown"
	}

	// Get query parameters (if any) using ctx.QueryParam() - works with SPA navigation
	tab := ctx.QueryParam("tab")
	view := ctx.QueryParam("view")
	sort := ctx.QueryParam("sort")
	order := ctx.QueryParam("order")

	// Mock user data
	users := map[string]map[string]string{
		"123": {
			"name":  "John Doe",
			"email": "john@example.com",
			"role":  "Admin",
		},
		"456": {
			"name":  "Jane Smith",
			"email": "jane@example.com",
			"role":  "User",
		},
		"alice": {
			"name":  "Alice Johnson",
			"email": "alice@example.com",
			"role":  "Moderator",
		},
	}

	user, exists := users[userID]
	if !exists {
		user = map[string]string{
			"name":  "Unknown User",
			"email": "N/A",
			"role":  "Guest",
		}
	}

	return ui.Div("flex flex-col gap-6")(
		ui.Div("flex items-center gap-4")(
			ui.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 cursor-pointer inline-block", ui.Href("/routes"), ctx.Load("/routes"))("← Back"),
			ui.Div("text-2xl font-bold")("User Details"),
		),

		ui.Card().Header("<h3 class='font-bold text-lg'>User Information</h3>").
			Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500 uppercase")("Route Parameter"),
						ui.Div("text-lg font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
							"ID: "+userID,
						),
					),
					ui.Div("grid grid-cols-1 md:grid-cols-2 gap-4")(
						ui.Div("flex flex-col gap-1")(
							ui.Div("text-sm font-bold text-gray-500")("Name"),
							ui.Div("text-lg")(user["name"]),
						),
						ui.Div("flex flex-col gap-1")(
							ui.Div("text-sm font-bold text-gray-500")("Email"),
							ui.Div("text-lg")(user["email"]),
						),
						ui.Div("flex flex-col gap-1")(
							ui.Div("text-sm font-bold text-gray-500")("Role"),
							ui.Badge().Color("blue").Text(user["role"]).Render(),
						),
					),
					func() string {
						if tab != "" || view != "" || sort != "" || order != "" {
							return ui.Div("flex flex-col gap-2 mt-4")(
								ui.Div("text-sm font-bold text-gray-500 uppercase")("Query Parameters"),
								ui.Div("grid grid-cols-1 md:grid-cols-2 gap-2")(
									func() string {
										if tab != "" {
											return ui.Div("text-sm font-mono bg-yellow-100 dark:bg-yellow-900/20 px-3 py-2 rounded")(
												"tab: " + tab,
											)
										}
										return ""
									}(),
									func() string {
										if view != "" {
											return ui.Div("text-sm font-mono bg-yellow-100 dark:bg-yellow-900/20 px-3 py-2 rounded")(
												"view: " + view,
											)
										}
										return ""
									}(),
									func() string {
										if sort != "" {
											return ui.Div("text-sm font-mono bg-yellow-100 dark:bg-yellow-900/20 px-3 py-2 rounded")(
												"sort: " + sort,
											)
										}
										return ""
									}(),
									func() string {
										if order != "" {
											return ui.Div("text-sm font-mono bg-yellow-100 dark:bg-yellow-900/20 px-3 py-2 rounded")(
												"order: " + order,
											)
										}
										return ""
									}(),
								),
							)
						}
						return ""
					}(),
					ui.Div("text-xs text-gray-500 mt-4 p-3 bg-blue-50 dark:bg-blue-900/20 rounded")(
						`<strong>Code:</strong> <code class="bg-white dark:bg-gray-800 px-1 rounded">userID := ctx.PathParam("id")</code> for path params, <code class="bg-white dark:bg-gray-800 px-1 rounded">tab := ctx.QueryParam("tab")</code> for query params`,
					),
				),
			).Render(),

		// Related links
		ui.Div("flex flex-col gap-2")(
			ui.Div("text-sm font-bold")("Try other users:"),
			ui.Div("flex flex-wrap gap-2")(
				ui.A("px-3 py-1 border-2 border-blue-600 text-blue-600 rounded hover:bg-blue-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/123"), ctx.Load("/routes/user/123"))("User 123"),
				ui.A("px-3 py-1 border-2 border-blue-600 text-blue-600 rounded hover:bg-blue-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/456"), ctx.Load("/routes/user/456"))("User 456"),
				ui.A("px-3 py-1 border-2 border-blue-600 text-blue-600 rounded hover:bg-blue-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/alice"), ctx.Load("/routes/user/alice"))("User alice"),
			),
			ui.Div("text-sm font-bold mt-2")("Try with query parameters:"),
			ui.Div("flex flex-wrap gap-2")(
				ui.A("px-3 py-1 border-2 border-indigo-600 text-indigo-600 rounded hover:bg-indigo-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/123?tab=profile&view=detailed"), ctx.Load("/routes/user/123?tab=profile&view=detailed"))("User 123: tab=profile"),
				ui.A("px-3 py-1 border-2 border-indigo-600 text-indigo-600 rounded hover:bg-indigo-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/456?tab=settings"), ctx.Load("/routes/user/456?tab=settings"))("User 456: tab=settings"),
				ui.A("px-3 py-1 border-2 border-indigo-600 text-indigo-600 rounded hover:bg-indigo-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/alice?sort=name&order=asc"), ctx.Load("/routes/user/alice?sort=name&order=asc"))("User alice: sort=name"),
			),
		),
	)
}

// UserPostDetail shows a post by user ID and post ID
func UserPostDetail(ctx *ui.Context) string {
	userID := ctx.PathParam("userId")
	postID := ctx.PathParam("postId")

	if userID == "" {
		userID = "unknown"
	}
	if postID == "" {
		postID = "unknown"
	}

	// Mock post data
	posts := map[string]map[string]string{
		"1": {
			"title":   "Getting Started with g-sui",
			"content": "This is a comprehensive guide to building server-rendered UIs with g-sui.",
			"author":  "123",
		},
		"42": {
			"title":   "Advanced Routing Patterns",
			"content": "Learn how to use parameterized routes effectively in your applications.",
			"author":  "456",
		},
		"my-first-post": {
			"title":   "My First Post",
			"content": "This is a blog post with a slug-based ID instead of numeric.",
			"author":  "alice",
		},
	}

	post, exists := posts[postID]
	if !exists {
		post = map[string]string{
			"title":   "Post Not Found",
			"content": "This post doesn't exist in our mock database.",
			"author":  "unknown",
		}
	}

	return ui.Div("flex flex-col gap-6")(
		ui.Div("flex items-center gap-4")(
			ui.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 cursor-pointer inline-block", ui.Href("/routes"), ctx.Load("/routes"))("← Back"),
			ui.Div("text-2xl font-bold")("Post Details"),
		),

		ui.Card().Header("<h3 class='font-bold text-lg'>Post Information</h3>").
			Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500 uppercase")("Route Parameters"),
						ui.Div("grid grid-cols-1 md:grid-cols-2 gap-2")(
							ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
								"userId: "+userID,
							),
							ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
								"postId: "+postID,
							),
						),
					),
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("Title"),
						ui.Div("text-xl font-bold")(post["title"]),
					),
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("Content"),
						ui.Div("text-gray-700 dark:text-gray-300")(post["content"]),
					),
					ui.Div("flex items-center gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("Author:"),
						ui.A("px-3 py-1 border-2 border-blue-600 text-blue-600 rounded hover:bg-blue-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/"+userID), ctx.Load("/routes/user/"+userID))("User "+userID),
					),
					ui.Div("text-xs text-gray-500 mt-4 p-3 bg-blue-50 dark:bg-blue-900/20 rounded")(
						`<strong>Code:</strong> <code class="bg-white dark:bg-gray-800 px-1 rounded">userId := ctx.PathParam("userId")</code> and <code class="bg-white dark:bg-gray-800 px-1 rounded">postId := ctx.PathParam("postId")</code>`,
					),
				),
			).Render(),

		// Related links
		ui.Div("flex flex-col gap-2")(
			ui.Div("text-sm font-bold")("Try other posts:"),
			ui.Div("flex flex-wrap gap-2")(
				ui.A("px-3 py-1 border-2 border-green-600 text-green-600 rounded hover:bg-green-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/123/post/1"), ctx.Load("/routes/user/123/post/1"))("User 123, Post 1"),
				ui.A("px-3 py-1 border-2 border-green-600 text-green-600 rounded hover:bg-green-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/456/post/42"), ctx.Load("/routes/user/456/post/42"))("User 456, Post 42"),
				ui.A("px-3 py-1 border-2 border-green-600 text-green-600 rounded hover:bg-green-50 cursor-pointer inline-block text-sm", ui.Href("/routes/user/alice/post/my-first-post"), ctx.Load("/routes/user/alice/post/my-first-post"))("User alice, Post 'my-first-post'"),
			),
		),
	)
}

// CategoryProductDetail shows a product by category and product name
func CategoryProductDetail(ctx *ui.Context) string {
	category := ctx.PathParam("category")
	product := ctx.PathParam("product")

	if category == "" {
		category = "unknown"
	}
	if product == "" {
		product = "unknown"
	}

	// Mock product data
	products := map[string]map[string]map[string]string{
		"electronics": {
			"laptop": {
				"name":        "Gaming Laptop",
				"price":       "$1,299",
				"description": "High-performance gaming laptop with RTX graphics.",
			},
		},
		"books": {
			"novel": {
				"name":        "The Great Novel",
				"price":       "$19.99",
				"description": "A captivating story that will keep you reading all night.",
			},
		},
	}

	var prod map[string]string
	if cat, ok := products[category]; ok {
		if p, ok := cat[product]; ok {
			prod = p
		}
	}

	if prod == nil {
		prod = map[string]string{
			"name":        "Product Not Found",
			"price":       "N/A",
			"description": "This product doesn't exist in our catalog.",
		}
	}

	return ui.Div("flex flex-col gap-6")(
		ui.Div("flex items-center gap-4")(
			ui.A("px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 cursor-pointer inline-block", ui.Href("/routes"), ctx.Load("/routes"))("← Back"),
			ui.Div("text-2xl font-bold")("Product Details"),
		),

		ui.Card().Header("<h3 class='font-bold text-lg'>Product Information</h3>").
			Body(
				ui.Div("flex flex-col gap-4")(
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500 uppercase")("Route Parameters"),
						ui.Div("grid grid-cols-1 md:grid-cols-2 gap-2")(
							ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
								"category: "+category,
							),
							ui.Div("text-sm font-mono bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded")(
								"product: "+product,
							),
						),
					),
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("Product Name"),
						ui.Div("text-xl font-bold")(prod["name"]),
					),
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("Price"),
						ui.Div("text-2xl font-bold text-green-600")(prod["price"]),
					),
					ui.Div("flex flex-col gap-2")(
						ui.Div("text-sm font-bold text-gray-500")("Description"),
						ui.Div("text-gray-700 dark:text-gray-300")(prod["description"]),
					),
					ui.Div("text-xs text-gray-500 mt-4 p-3 bg-purple-50 dark:bg-purple-900/20 rounded")(
						`<strong>Code:</strong> <code class="bg-white dark:bg-gray-800 px-1 rounded">category := ctx.PathParam("category")</code> and <code class="bg-white dark:bg-gray-800 px-1 rounded">product := ctx.PathParam("product")</code>`,
					),
				),
			).Render(),

		// Related links
		ui.Div("flex flex-col gap-2")(
			ui.Div("text-sm font-bold")("Try other products:"),
			ui.Div("flex flex-wrap gap-2")(
				ui.A("px-3 py-1 border-2 border-purple-600 text-purple-600 rounded hover:bg-purple-50 cursor-pointer inline-block text-sm", ui.Href("/routes/category/electronics/product/laptop"), ctx.Load("/routes/category/electronics/product/laptop"))("Electronics → Laptop"),
				ui.A("px-3 py-1 border-2 border-purple-600 text-purple-600 rounded hover:bg-purple-50 cursor-pointer inline-block text-sm", ui.Href("/routes/category/books/product/novel"), ctx.Load("/routes/category/books/product/novel"))("Books → Novel"),
			),
		),
	)
}

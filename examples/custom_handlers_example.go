package main

// This file demonstrates how to use custom HTTP handlers with g-sui
// Run this example separately: go run examples/custom_handlers_example.go

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// Example data structures
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type HealthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

func main_custom_handlers_example() {
	app := ui.MakeApp("en")

	// Register g-sui pages
	app.Page("/", "Home", func(ctx *ui.Context) string {
		return app.HTML("Home", "bg-gray-100",
			ui.Div("p-8 max-w-4xl mx-auto")(
				ui.Div("text-3xl font-bold mb-6")("Custom Handlers Example"),
				ui.Div("prose mb-8")(
					ui.P("")("This example demonstrates mixing g-sui pages with REST API endpoints."),
				),
				ui.Div("grid grid-cols-1 md:grid-cols-2 gap-4")(
					apiEndpointCard("GET", "/api/health", "Health check endpoint"),
					apiEndpointCard("GET", "/api/users", "Get all users"),
					apiEndpointCard("POST", "/api/users", "Create a new user"),
					apiEndpointCard("PUT", "/api/users/:id", "Update a user"),
					apiEndpointCard("DELETE", "/api/users/:id", "Delete a user"),
				),
			),
		)
	})

	// Register custom HTTP handlers (REST API endpoints)
	// These are checked BEFORE g-sui routes

	// Health check endpoint
	app.GET("/api/health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status: "ok",
			Time:   time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Get all users
	app.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
		users := []User{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
			{ID: 3, Name: "Charlie"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	// Create a new user
	app.POST("/api/users", func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Simulate saving (in real app, save to database)
		user.ID = 4 // Assign new ID

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	})

	// Update a user
	app.PUT("/api/users/:id", func(w http.ResponseWriter, r *http.Request) {
		// Note: Path parameters like :id need to be extracted manually in custom handlers
		// (g-sui's ctx.PathParam() only works in Page handlers)

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Simulate updating
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	// Delete a user
	app.DELETE("/api/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Alternative: Using Custom with full method specification
	app.Custom("PATCH", "/api/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "User patched"})
	})

	log.Println("Server starting on http://localhost:8080")
	log.Println("Try:")
	log.Println("  curl http://localhost:8080/api/health")
	log.Println("  curl http://localhost:8080/api/users")
	log.Println("  curl -X POST http://localhost:8080/api/users -d '{\"name\":\"Dave\"}'")

	app.Listen(":8080")
}

func apiEndpointCard(method, path, description string) string {
	colorMap := map[string]string{
		"GET":    "bg-blue-100 text-blue-800",
		"POST":   "bg-green-100 text-green-800",
		"PUT":    "bg-yellow-100 text-yellow-800",
		"DELETE": "bg-red-100 text-red-800",
		"PATCH":  "bg-purple-100 text-purple-800",
	}

	color := colorMap[method]
	if color == "" {
		color = "bg-gray-100 text-gray-800"
	}

	return ui.Div("bg-white rounded-lg shadow p-4")(
		ui.Div("flex items-center gap-2 mb-2")(
			ui.Span("px-2 py-1 rounded text-xs font-semibold "+color)(method),
			ui.Span("font-mono text-sm")(path),
		),
		ui.Div("text-gray-600 text-sm")(description),
	)
}

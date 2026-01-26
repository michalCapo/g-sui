package ui

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestParseRoutePattern(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		wantSegs   int
		wantParams []string
		wantHas    bool
	}{
		{"exact path", "/vehicles", 1, nil, false},
		{"single param", "/vehicles/{id}", 2, []string{"id"}, true},
		{"multiple params", "/users/{userId}/posts/{postId}", 4, []string{"userId", "postId"}, true},
		{"root with param", "/{slug}", 1, []string{"slug"}, true},
		{"nested path", "/api/v1/users/{id}", 4, []string{"id"}, true},
		{"no leading slash", "vehicles/{id}", 2, []string{"id"}, true},
		{"trailing slash", "/vehicles/{id}/", 2, []string{"id"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs, params, has := parseRoutePattern(tt.pattern)
			if len(segs) != tt.wantSegs {
				t.Errorf("parseRoutePattern() segments = %v, want %d", segs, tt.wantSegs)
			}
			if len(params) != len(tt.wantParams) {
				t.Errorf("parseRoutePattern() params = %v, want %v", params, tt.wantParams)
			}
			for i, p := range tt.wantParams {
				if i < len(params) && params[i] != p {
					t.Errorf("parseRoutePattern() params[%d] = %q, want %q", i, params[i], p)
				}
			}
			if has != tt.wantHas {
				t.Errorf("parseRoutePattern() hasParams = %v, want %v", has, tt.wantHas)
			}
		})
	}
}

func TestMatchRoutePattern(t *testing.T) {
	route := &Route{
		Path:       "/vehicles/edit/{id}",
		Segments:   []string{"vehicles", "edit", "{id}"},
		ParamNames: []string{"id"},
		HasParams:  true,
	}

	tests := []struct {
		name      string
		path      string
		wantMatch bool
		wantID    string
	}{
		{"exact match", "/vehicles/edit/123", true, "123"},
		{"different id", "/vehicles/edit/456", true, "456"},
		{"wrong segment", "/vehicles/view/123", false, ""},
		{"missing segment", "/vehicles/123", false, ""},
		{"extra segment", "/vehicles/edit/123/details", false, ""},
		{"empty param", "/vehicles/edit/", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := matchRoutePattern(tt.path, route)
			if (params != nil) != tt.wantMatch {
				t.Errorf("matchRoutePattern() match = %v, want %v", params != nil, tt.wantMatch)
				return
			}
			if tt.wantMatch {
				if params["id"] != tt.wantID {
					t.Errorf("matchRoutePattern() id = %q, want %q", params["id"], tt.wantID)
				}
			}
		})
	}
}

func TestMatchRoutePattern_MultipleParams(t *testing.T) {
	route := &Route{
		Path:       "/users/{userId}/posts/{postId}",
		Segments:   []string{"users", "{userId}", "posts", "{postId}"},
		ParamNames: []string{"userId", "postId"},
		HasParams:  true,
	}

	params := matchRoutePattern("/users/123/posts/456", route)
	if params == nil {
		t.Fatal("matchRoutePattern() should match")
	}
	if params["userId"] != "123" {
		t.Errorf("matchRoutePattern() userId = %q, want %q", params["userId"], "123")
	}
	if params["postId"] != "456" {
		t.Errorf("matchRoutePattern() postId = %q, want %q", params["postId"], "456")
	}
}

func TestMatchRoute_ExactMatch(t *testing.T) {
	app := MakeApp("en")
	app.Page("/vehicles", "Vehicles", func(ctx *Context) string {
		return "<div>Vehicles</div>"
	})

	route, params := app.matchRoute("/vehicles")
	if route == nil {
		t.Fatal("matchRoute() should find exact match")
	}
	if params != nil {
		t.Error("matchRoute() should not return params for exact match")
	}
	if route.Path != "/vehicles" {
		t.Errorf("matchRoute() path = %q, want %q", route.Path, "/vehicles")
	}
}

func TestMatchRoute_PatternMatch(t *testing.T) {
	app := MakeApp("en")
	app.Page("/vehicles/edit/{id}", "Edit Vehicle", func(ctx *Context) string {
		return "<div>Edit " + ctx.PathParam("id") + "</div>"
	})

	route, params := app.matchRoute("/vehicles/edit/123")
	if route == nil {
		t.Fatal("matchRoute() should find pattern match")
	}
	if params == nil {
		t.Fatal("matchRoute() should return params for pattern match")
	}
	if params["id"] != "123" {
		t.Errorf("matchRoute() id = %q, want %q", params["id"], "123")
	}
}

func TestContext_PathParam(t *testing.T) {
	ctx := &Context{
		pathParams: map[string]string{
			"id":     "123",
			"slug":   "test-page",
			"userId": "456",
		},
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing param", "id", "123"},
		{"another param", "slug", "test-page"},
		{"third param", "userId", "456"},
		{"non-existent", "missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctx.PathParam(tt.key)
			if got != tt.want {
				t.Errorf("PathParam(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestContext_PathParam_Nil(t *testing.T) {
	ctx := &Context{}
	if ctx.PathParam("id") != "" {
		t.Error("PathParam() should return empty string when pathParams is nil")
	}
}

func TestPageRoute_WithPathParams(t *testing.T) {
	app := MakeApp("en")

	app.Page("/vehicles/edit/{id}", "Edit Vehicle", func(ctx *Context) string {
		return "<div>Edit " + ctx.PathParam("id") + "</div>"
	})

	// Verify route was registered with pattern
	app.routesMu.RLock()
	route, exists := app.routes["/vehicles/edit/{id}"]
	app.routesMu.RUnlock()

	if !exists {
		t.Fatal("Route should be registered")
	}
	if !route.HasParams {
		t.Error("Route should have HasParams = true")
	}
	if len(route.ParamNames) != 1 || route.ParamNames[0] != "id" {
		t.Errorf("Route ParamNames = %v, want [id]", route.ParamNames)
	}
}

func TestPageRoute_SPARequest(t *testing.T) {
	// This test verifies that routes with patterns are registered correctly
	// Full HTTP integration testing would require the full Listen() handler
	app := MakeApp("en")

	app.Page("/vehicles/edit/{id}", "Edit Vehicle", func(ctx *Context) string {
		return "<div>Edit " + ctx.PathParam("id") + "</div>"
	})

	// Verify route was registered
	app.routesMu.RLock()
	_, exists := app.routes["/vehicles/edit/{id}"]
	app.routesMu.RUnlock()

	if !exists {
		t.Fatal("Route should be registered")
	}

	// Verify pattern matching works
	matchedRoute, params := app.matchRoute("/vehicles/edit/123")
	if matchedRoute == nil {
		t.Fatal("matchRoute() should find pattern match")
	}
	if params == nil {
		t.Fatal("matchRoute() should return params")
	}
	if params["id"] != "123" {
		t.Errorf("matchRoute() id = %q, want %q", params["id"], "123")
	}
}

func TestRouteManifest_WithPatterns(t *testing.T) {
	app := MakeApp("en")
	app.Page("/", "Home", func(ctx *Context) string { return "" })
	app.Page("/vehicles", "Vehicles", func(ctx *Context) string { return "" })
	app.Page("/vehicles/edit/{id}", "Edit Vehicle", func(ctx *Context) string { return "" })

	app.routesMu.RLock()
	manifest := make(map[string]interface{})
	for path, route := range app.routes {
		if route.HasParams {
			manifest[path] = map[string]interface{}{
				"path":    path,
				"pattern": true,
			}
		} else {
			manifest[path] = path
		}
	}
	app.routesMu.RUnlock()

	// Check exact routes are strings (paths)
	if path, ok := manifest["/"].(string); !ok || path == "" {
		t.Error("Exact route should be path string")
	}
	if path, ok := manifest["/vehicles"].(string); !ok || path == "" {
		t.Error("Exact route should be path string")
	}

	// Check pattern route is object
	patternRoute, ok := manifest["/vehicles/edit/{id}"].(map[string]interface{})
	if !ok {
		t.Fatal("Pattern route should be object")
	}
	if patternRoute["path"] == "" {
		t.Error("Pattern route should have path")
	}
	if patternRoute["pattern"] != true {
		t.Error("Pattern route should have pattern = true")
	}
}

func TestQueryStringPreservation(t *testing.T) {
	// Test that query strings are preserved in URL parsing
	// Full integration testing requires the Listen() handler
	app := MakeApp("en")

	app.Page("/search", "Search", func(ctx *Context) string {
		return "<div>Search</div>"
	})

	// Verify route matching strips query string correctly
	pathWithQuery := "/search?q=test"
	pathForMatching := pathWithQuery
	if queryIdx := strings.Index(pathForMatching, "?"); queryIdx >= 0 {
		pathForMatching = pathForMatching[:queryIdx]
	}

	route, _ := app.matchRoute(pathForMatching)
	if route == nil {
		t.Error("matchRoute() should find route when query string is stripped")
	}
	if pathForMatching != "/search" {
		t.Errorf("Query string stripping failed: got %q, want %q", pathForMatching, "/search")
	}
}

func TestContext_QueryParam(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string][]string
		key         string
		want        string
	}{
		{"single value", map[string][]string{"name": {"Smith"}}, "name", "Smith"},
		{"multiple values", map[string][]string{"tags": {"a", "b", "c"}}, "tags", "a"},
		{"empty value", map[string][]string{"empty": {""}}, "empty", ""},
		{"non-existent", map[string][]string{"name": {"Smith"}}, "missing", ""},
		{"nil map", nil, "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				queryParams: tt.queryParams,
			}
			got := ctx.QueryParam(tt.key)
			if got != tt.want {
				t.Errorf("QueryParam(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestContext_QueryParam_FallbackToRequest(t *testing.T) {
	// Test fallback to Request.URL.Query() when queryParams is nil
	reqURL, _ := url.Parse("http://example.com/search?name=Smith&age=30")
	req := &http.Request{
		URL: reqURL,
	}

	ctx := &Context{
		Request:     req,
		queryParams: nil, // Not set, should fallback
	}

	if ctx.QueryParam("name") != "Smith" {
		t.Errorf("QueryParam(\"name\") = %q, want %q", ctx.QueryParam("name"), "Smith")
	}
	if ctx.QueryParam("age") != "30" {
		t.Errorf("QueryParam(\"age\") = %q, want %q", ctx.QueryParam("age"), "30")
	}
	if ctx.QueryParam("missing") != "" {
		t.Errorf("QueryParam(\"missing\") = %q, want empty string", ctx.QueryParam("missing"))
	}
}

func TestContext_QueryParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string][]string
		key         string
		want        []string
	}{
		{"single value", map[string][]string{"name": {"Smith"}}, "name", []string{"Smith"}},
		{"multiple values", map[string][]string{"tags": {"a", "b", "c"}}, "tags", []string{"a", "b", "c"}},
		{"empty value", map[string][]string{"empty": {""}}, "empty", []string{""}},
		{"non-existent", map[string][]string{"name": {"Smith"}}, "missing", nil},
		{"nil map", nil, "name", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				queryParams: tt.queryParams,
			}
			got := ctx.QueryParams(tt.key)
			if len(got) != len(tt.want) {
				t.Errorf("QueryParams(%q) length = %d, want %d", tt.key, len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("QueryParams(%q)[%d] = %q, want %q", tt.key, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestContext_QueryParams_FallbackToRequest(t *testing.T) {
	// Test fallback to Request.URL.Query() when queryParams is nil
	reqURL, _ := url.Parse("http://example.com/search?tags=a&tags=b&tags=c")
	req := &http.Request{
		URL: reqURL,
	}

	ctx := &Context{
		Request:     req,
		queryParams: nil, // Not set, should fallback
	}

	got := ctx.QueryParams("tags")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Errorf("QueryParams(\"tags\") length = %d, want %d", len(got), len(want))
		return
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("QueryParams(\"tags\")[%d] = %q, want %q", i, v, want[i])
		}
	}
}

func TestContext_AllQueryParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string][]string
		want        map[string][]string
	}{
		{"with params", map[string][]string{"name": {"Smith"}, "age": {"30"}}, map[string][]string{"name": {"Smith"}, "age": {"30"}}},
		{"empty map", map[string][]string{}, map[string][]string{}},
		{"nil map", nil, nil},
		{"multiple values", map[string][]string{"tags": {"a", "b"}}, map[string][]string{"tags": {"a", "b"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				queryParams: tt.queryParams,
			}
			got := ctx.AllQueryParams()
			if tt.want == nil {
				if got != nil {
					t.Errorf("AllQueryParams() = %v, want nil", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("AllQueryParams() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				gotV := got[k]
				if len(gotV) != len(v) {
					t.Errorf("AllQueryParams()[%q] length = %d, want %d", k, len(gotV), len(v))
					continue
				}
				for i, val := range v {
					if gotV[i] != val {
						t.Errorf("AllQueryParams()[%q][%d] = %q, want %q", k, i, gotV[i], val)
					}
				}
			}
		})
	}
}

func TestContext_AllQueryParams_FallbackToRequest(t *testing.T) {
	// Test fallback to Request.URL.Query() when queryParams is nil
	reqURL, _ := url.Parse("http://example.com/search?name=Smith&age=30")
	req := &http.Request{
		URL: reqURL,
	}

	ctx := &Context{
		Request:     req,
		queryParams: nil, // Not set, should fallback
	}

	got := ctx.AllQueryParams()
	if got == nil {
		t.Fatal("AllQueryParams() should not return nil when Request has query params")
	}
	if len(got["name"]) == 0 || got["name"][0] != "Smith" {
		t.Errorf("AllQueryParams()[\"name\"][0] = %q, want %q", got["name"], "Smith")
	}
	if len(got["age"]) == 0 || got["age"][0] != "30" {
		t.Errorf("AllQueryParams()[\"age\"][0] = %q, want %q", got["age"], "30")
	}
}

func TestContext_QueryParam_Priority(t *testing.T) {
	// Test that queryParams takes priority over Request.URL.Query()
	reqURL, _ := url.Parse("http://example.com/search?name=RequestValue")
	req := &http.Request{
		URL: reqURL,
	}

	ctx := &Context{
		Request: req,
		queryParams: map[string][]string{
			"name": {"ContextValue"},
		},
	}

	got := ctx.QueryParam("name")
	if got != "ContextValue" {
		t.Errorf("QueryParam(\"name\") = %q, want %q (should use queryParams, not Request)", got, "ContextValue")
	}
}

func TestContext_PathAndQueryParams(t *testing.T) {
	// Test that path params and query params work together
	ctx := &Context{
		pathParams: map[string]string{
			"id": "123",
		},
		queryParams: map[string][]string{
			"tab":  {"profile"},
			"view": {"detailed"},
		},
	}

	if ctx.PathParam("id") != "123" {
		t.Errorf("PathParam(\"id\") = %q, want %q", ctx.PathParam("id"), "123")
	}
	if ctx.QueryParam("tab") != "profile" {
		t.Errorf("QueryParam(\"tab\") = %q, want %q", ctx.QueryParam("tab"), "profile")
	}
	if ctx.QueryParam("view") != "detailed" {
		t.Errorf("QueryParam(\"view\") = %q, want %q", ctx.QueryParam("view"), "detailed")
	}
}

func TestQueryParamParsing(t *testing.T) {
	// Test parsing query string from path
	tests := []struct {
		name       string
		path       string
		wantParams map[string][]string
		wantPath   string
	}{
		{
			"single param",
			"/search?name=Smith",
			map[string][]string{"name": {"Smith"}},
			"/search",
		},
		{
			"multiple params",
			"/search?name=Smith&age=30",
			map[string][]string{"name": {"Smith"}, "age": {"30"}},
			"/search",
		},
		{
			"no query string",
			"/search",
			nil,
			"/search",
		},
		{
			"empty value",
			"/search?name=",
			map[string][]string{"name": {""}},
			"/search",
		},
		{
			"special characters",
			"/search?q=hello%20world&tag=g-sui",
			map[string][]string{"q": {"hello world"}, "tag": {"g-sui"}},
			"/search",
		},
		{
			"with path params",
			"/user/123?tab=profile&view=detailed",
			map[string][]string{"tab": {"profile"}, "view": {"detailed"}},
			"/user/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathForMatching := tt.path
			var queryParams map[string][]string

			if pathForMatching != "" {
				if queryIdx := strings.Index(pathForMatching, "?"); queryIdx >= 0 {
					queryString := pathForMatching[queryIdx+1:]
					pathForMatching = pathForMatching[:queryIdx]
					// Parse query string into map
					if parsedURL, err := url.Parse("?" + queryString); err == nil {
						queryParams = parsedURL.Query()
					}
				}
			}

			if pathForMatching != tt.wantPath {
				t.Errorf("pathForMatching = %q, want %q", pathForMatching, tt.wantPath)
			}

			if tt.wantParams == nil {
				if queryParams != nil {
					t.Errorf("queryParams = %v, want nil", queryParams)
				}
				return
			}

			if len(queryParams) != len(tt.wantParams) {
				t.Errorf("queryParams length = %d, want %d", len(queryParams), len(tt.wantParams))
				return
			}

			for k, wantV := range tt.wantParams {
				gotV := queryParams[k]
				if len(gotV) != len(wantV) {
					t.Errorf("queryParams[%q] length = %d, want %d", k, len(gotV), len(wantV))
					continue
				}
				for i, val := range wantV {
					if gotV[i] != val {
						t.Errorf("queryParams[%q][%d] = %q, want %q", k, i, gotV[i], val)
					}
				}
			}
		})
	}
}

package ui

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/html"
	"golang.org/x/net/websocket"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Callable = func(*Context) string

var (
	eventPath      = "/"
	reReplaceChars = regexp.MustCompile(`[./:-]`)
	reRemoveChars  = regexp.MustCompile(`[*()\[\]]`)
)

type BodyItem struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// FileUpload represents an uploaded file from a form submission
type FileUpload struct {
	Name        string // Original filename
	Data        []byte // File content (decoded from Base64)
	ContentType string // MIME type (e.g., "image/png")
	Size        int    // File size in bytes
}

// JSON DOM Protocol Types

// JSEvent represents a declarative event handler in JSON format
type JSEvent struct {
	Act  string     `json:"act"`            // "post", "form", or "raw"
	Swap string     `json:"swap,omitempty"` // "inline", "outline", "append", "prepend", "none"
	Tgt  string     `json:"tgt,omitempty"`  // target element ID
	Path string     `json:"path,omitempty"` // server endpoint path
	Vals []BodyItem `json:"vals,omitempty"` // pre-populated values
	JS   string     `json:"js,omitempty"`   // raw JavaScript code (for act="raw")
}

// JSElement represents a DOM element in JSON format
type JSElement struct {
	T string              `json:"t"`           // tag name
	A map[string]string   `json:"a,omitempty"` // attributes (id, class, style, etc.)
	E map[string]*JSEvent `json:"e,omitempty"` // events (click, change, submit)
	C []interface{}       `json:"c,omitempty"` // children (strings or JSElement objects)
}

// JSPatchOp represents a single patch operation
type JSPatchOp struct {
	Op  string     `json:"op"`            // "inline", "outline", "append", "prepend", "none", "notify", "title", "reload", "redirect", "download"
	Tgt string     `json:"tgt,omitempty"` // target element ID
	El  *JSElement `json:"el,omitempty"`  // element to insert/replace
	JS  string     `json:"js,omitempty"`  // raw JavaScript (for backwards compatibility)
	// Notification fields (when op == "notify")
	Msg     string `json:"msg,omitempty"`     // notification message
	Variant string `json:"variant,omitempty"` // "success", "error", "info", "error-reload"
	// Title field (when op == "title")
	Title string `json:"title,omitempty"` // page title
	// Redirect field (when op == "redirect")
	Href string `json:"href,omitempty"` // redirect URL
	// Download fields (when op == "download")
	Data        string `json:"data,omitempty"`         // base64-encoded file content
	ContentType string `json:"content_type,omitempty"` // MIME type
	Filename    string `json:"filename,omitempty"`     // download filename
}

// JSPatchMessage is the WebSocket patch message format
type JSPatchMessage struct {
	Type string       `json:"type"` // "patch"
	Ops  []*JSPatchOp `json:"ops"`  // operations to apply
}

// JSHTTPResponse is the HTTP POST response format
type JSHTTPResponse struct {
	El  *JSElement   `json:"el"`            // element to render
	Ops []*JSPatchOp `json:"ops,omitempty"` // additional operations (notifications, title, etc.)
}

// JSCallMessage is the WebSocket request for callable actions
type JSCallMessage struct {
	Type string     `json:"type"` // "call"
	RID  string     `json:"rid"`  // request ID for correlation
	Act  string     `json:"act"`  // "post" or "form"
	Path string     `json:"path"` // callable endpoint path
	Swap string     `json:"swap"` // "inline", "outline", "append", "prepend", "none"
	Tgt  string     `json:"tgt"`  // target element ID
	Vals []BodyItem `json:"vals"` // values/payload
}

// JSResponseMessage is the WebSocket response for callable actions
type JSResponseMessage struct {
	Type string       `json:"type"` // "response"
	RID  string       `json:"rid"`  // matching request ID
	El   *JSElement   `json:"el"`   // element to render
	Ops  []*JSPatchOp `json:"ops"`  // additional operations
}

type CSS struct {
	Orig   string
	Set    string
	Append []string
}

func (c *CSS) Value() string {
	if len(c.Set) > 0 {
		return c.Set
	}

	c.Append = append(c.Append, c.Orig)
	return Classes(c.Append...)
}

type Swap string

const (
	OUTLINE Swap = "outline"
	INLINE  Swap = "inline"
	APPEND  Swap = "append"
	PREPEND Swap = "prepend"
	NONE    Swap = "none"
)

type ActionType string

const (
	POST ActionType = "POST"
	FORM ActionType = "FORM"
)

type Context struct {
	App         *App
	Request     *http.Request
	Response    http.ResponseWriter
	SessionID   string
	append      []string
	ops         []*JSPatchOp        // additional operations (notifications, title changes, etc.)
	pathParams  map[string]string   // extracted path parameters from route patterns
	queryParams map[string][]string // query parameters from URL (for SPA navigation)
	fileItems   []BodyItem          // Store file uploads from WebSocket
}

type TSession struct {
	DB        *gorm.DB `gorm:"-"`
	SessionID string
	Name      string
	Data      datatypes.JSON
}

func (TSession) TableName() string {
	return "_session"
}

func (session *TSession) Load(data any) {
	temp := &TSession{}

	err := session.DB.Where("session_id = ? and name = ?", session.SessionID, session.Name).Take(temp).Error
	if err != nil {
		return
	}

	err = json.Unmarshal(temp.Data, data)
	if err != nil {
		log.Println(err)
		return
	}
}

func (session *TSession) Save(output any) {
	data, err := json.Marshal(output)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	temp := &TSession{
		SessionID: session.SessionID,
		Name:      session.Name,
		Data:      data,
	}

	err = session.DB.Where("session_id = ? and name = ?", session.SessionID, session.Name).Save(temp).Error
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func (ctx *Context) IP() string {
	return ctx.Request.RemoteAddr
}

// PathParam returns the value of a path parameter extracted from the route pattern.
// Returns empty string if the parameter doesn't exist.
func (ctx *Context) PathParam(name string) string {
	if ctx.pathParams == nil {
		return ""
	}
	return ctx.pathParams[name]
}

// QueryParam returns the value of a query parameter from the URL.
// For SPA navigation, this returns params from the navigated URL.
// For direct requests, this falls back to ctx.Request.URL.Query().
// Returns empty string if the parameter doesn't exist.
func (ctx *Context) QueryParam(name string) string {
	if ctx.queryParams != nil {
		if values, ok := ctx.queryParams[name]; ok && len(values) > 0 {
			return values[0]
		}
		return ""
	}
	// Fallback to request URL query for non-SPA requests
	if ctx.Request != nil && ctx.Request.URL != nil {
		return ctx.Request.URL.Query().Get(name)
	}
	return ""
}

// QueryParams returns all values for a query parameter (for multi-value params).
// Returns nil if the parameter doesn't exist.
func (ctx *Context) QueryParams(name string) []string {
	if ctx.queryParams != nil {
		return ctx.queryParams[name]
	}
	// Fallback to request URL query for non-SPA requests
	if ctx.Request != nil && ctx.Request.URL != nil {
		return ctx.Request.URL.Query()[name]
	}
	return nil
}

// AllQueryParams returns all query parameters as a map.
func (ctx *Context) AllQueryParams() map[string][]string {
	if ctx.queryParams != nil {
		return ctx.queryParams
	}
	if ctx.Request != nil && ctx.Request.URL != nil {
		return ctx.Request.URL.Query()
	}
	return nil
}

// File returns a single uploaded file by name
// Returns nil if no file found with that name
func (ctx *Context) File(name string) (*FileUpload, error) {
	for _, item := range ctx.fileItems {
		if item.Type == "file" && item.Name == name && item.Value != "" {
			data, err := base64.StdEncoding.DecodeString(item.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to decode file %s: %w", name, err)
			}
			return &FileUpload{
				Name:        item.Filename,
				Data:        data,
				ContentType: item.ContentType,
				Size:        len(data),
			}, nil
		}
	}
	return nil, nil // No file found (not an error)
}

// Files returns all uploaded files with the given name (for multiple file inputs)
func (ctx *Context) Files(name string) ([]*FileUpload, error) {
	var files []*FileUpload
	for _, item := range ctx.fileItems {
		if item.Type == "file" && item.Name == name && item.Value != "" {
			data, err := base64.StdEncoding.DecodeString(item.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to decode file %s: %w", name, err)
			}
			files = append(files, &FileUpload{
				Name:        item.Filename,
				Data:        data,
				ContentType: item.ContentType,
				Size:        len(data),
			})
		}
	}
	return files, nil
}

func (ctx *Context) Session(db *gorm.DB, name string) *TSession {
	return &TSession{
		DB:        db,
		Name:      name,
		SessionID: ctx.SessionID,
	}
}

// Maximum input size limits for security
const (
	MaxBodySize      = 10 * 1024 * 1024 // 10MB max request body
	MaxFieldNameLen  = 256              // Max field name length
	MaxFieldValueLen = 1024 * 1024      // 1MB max field value
	MaxFieldCount    = 1000             // Max number of fields
)

// validateInputSafety performs comprehensive input validation
func validateInputSafety(data []BodyItem) error {
	if len(data) > MaxFieldCount {
		return fmt.Errorf("too many fields: %d exceeds maximum of %d", len(data), MaxFieldCount)
	}

	for i, item := range data {
		// Validate field name length and content
		if len(item.Name) > MaxFieldNameLen {
			return fmt.Errorf("field name too long at index %d: %d exceeds maximum of %d", i, len(item.Name), MaxFieldNameLen)
		}

		if item.Name == "" {
			return fmt.Errorf("empty field name at index %d", i)
		}

		// Validate field name contains only safe characters
		for _, r := range item.Name {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '.' && r != '[' && r != ']' && r != '_' {
				return fmt.Errorf("unsafe character in field name at index %d: '%c'", i, r)
			}
		}

		// Validate field value length
		if len(item.Value) > MaxFieldValueLen {
			return fmt.Errorf("field value too long at index %d: %d exceeds maximum of %d", i, len(item.Value), MaxFieldValueLen)
		}

		// Validate type field
		if len(item.Type) > 64 {
			return fmt.Errorf("field type too long at index %d: %d exceeds maximum of 64", i, len(item.Type))
		}

		// Validate filename length (for file uploads)
		if len(item.Filename) > 512 {
			return fmt.Errorf("filename too long at index %d: %d exceeds maximum of 512", i, len(item.Filename))
		}

		// Validate content type length (for file uploads)
		if len(item.ContentType) > 256 {
			return fmt.Errorf("content type too long at index %d: %d exceeds maximum of 256", i, len(item.ContentType))
		}
	}

	return nil
}

// validateNumericInput validates numeric input with bounds checking and returns parsed value
func validateNumericInput(value string, inputType string) (any, error) {
	switch inputType {
	case "int":
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 20 { // Prevent overflow attacks
			return nil, fmt.Errorf("integer value too long: %d characters", len(cleanedValue))
		}
		n, err := strconv.ParseInt(cleanedValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return int(n), nil

	case "int64":
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 20 {
			return nil, fmt.Errorf("int64 value too long: %d characters", len(cleanedValue))
		}
		n, err := strconv.ParseInt(cleanedValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return n, nil

	case "uint":
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 20 {
			return nil, fmt.Errorf("uint value too long: %d characters", len(cleanedValue))
		}
		n, err := strconv.ParseUint(cleanedValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint(n), nil

	case "number":
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 20 {
			return nil, fmt.Errorf("number value too long: %d characters", len(cleanedValue))
		}
		n, err := strconv.Atoi(cleanedValue)
		if err != nil {
			return nil, err
		}
		return n, nil

	case "decimal", "float64":
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 50 { // Allow longer decimals but still bounded
			return nil, fmt.Errorf("decimal value too long: %d characters", len(cleanedValue))
		}
		f, err := strconv.ParseFloat(cleanedValue, 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	return nil, nil
}

func (ctx *Context) Body(output any) error {
	contentType := ctx.Request.Header.Get("Content-Type")

	// Handle multipart/form-data (file uploads)
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(MaxBodySize); err != nil {
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}

		// Process form values (non-file fields)
		for name, values := range ctx.Request.MultipartForm.Value {
			if len(values) == 0 {
				continue
			}
			value := values[0]

			structFieldValue, err := PathValue(output, name)
			if err != nil {
				fmt.Printf("Warning: Error getting field %s: %v\n", name, err)
				continue
			}

			if !structFieldValue.CanSet() {
				fmt.Printf("Warning: Cannot set field %s\n", name)
				continue
			}

			// Try to set the value based on the field type
			if err := setFieldValue(structFieldValue, value); err != nil {
				fmt.Printf("Warning: Error setting field %s: %v\n", name, err)
			}
		}

		// Note: File fields are NOT processed here - they must be accessed via ctx.Request.FormFile()
		// This is intentional as file handling requires special treatment in the handler

		return nil
	}

	// Handle JSON (original behavior)
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, MaxBodySize))
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Check if body was truncated due to size limit
	if len(body) >= MaxBodySize {
		return fmt.Errorf("request body too large: exceeds %d bytes", MaxBodySize)
	}

	var data []BodyItem
	if len(body) > 0 {
		err = json.Unmarshal(body, &data)
		if err != nil {
			return fmt.Errorf("failed to parse JSON body: %w", err)
		}
	}

	// Validate input safety
	if err := validateInputSafety(data); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	var firstErr error
	for i, item := range data {
		structFieldValue, err := PathValue(output, item.Name)
		if err != nil {
			fmt.Printf("Warning: Error getting field %s at index %d: %v\n", item.Name, i, err)
			continue
		}

		if !structFieldValue.CanSet() {
			fmt.Printf("Warning: Cannot set field %s at index %d\n", item.Name, i)
			continue
		}

		if err := setFieldValue(structFieldValue, item.Value); err != nil {
			fmt.Printf("Warning: Error setting field %s at index %d: %v\n", item.Name, i, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// parseTimeValue tries multiple time formats to parse a string value into time.Time
func parseTimeValue(value string) (time.Time, error) {
	formats := []string{
		"2006-01-02",                    // HTML date input
		"2006-01-02T15:04",              // HTML datetime-local
		"15:04",                         // HTML time input
		"2006-01-02 15:04:05 -0700 MST", // Go full timestamp with timezone abbreviation
		"2006-01-02 15:04:05 -0700 UTC", // Go full timestamp with UTC
		time.RFC3339,                    // ISO 8601
		time.RFC3339Nano,                // ISO 8601 with nanoseconds
	}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", value)
}

// setFieldValue sets a struct field value by inferring the type from the Go struct field
func setFieldValue(structFieldValue *reflect.Value, value string) error {
	fieldType := structFieldValue.Type()
	fieldKind := structFieldValue.Kind()

	// Direct string assignment
	if fieldKind == reflect.String {
		structFieldValue.SetString(value)
		return nil
	}

	// Handle type aliases (e.g., type MyString string, type Skeleton string)
	if fieldType == reflect.TypeFor[Skeleton]() {
		structFieldValue.Set(reflect.ValueOf(Skeleton(value)))
		return nil
	}

	// Handle gorm.DeletedAt specially
	if fieldType == reflect.TypeFor[gorm.DeletedAt]() {
		if value == "" {
			structFieldValue.Set(reflect.ValueOf(gorm.DeletedAt{}))
			return nil
		}
		t, err := parseTimeValue(value)
		if err != nil {
			return fmt.Errorf("error parsing date for DeletedAt: %w", err)
		}
		structFieldValue.Set(reflect.ValueOf(gorm.DeletedAt{Time: t, Valid: true}))
		return nil
	}

	// Handle time.Time
	if fieldType == reflect.TypeFor[time.Time]() {
		if value == "" {
			structFieldValue.Set(reflect.ValueOf(time.Time{}))
			return nil
		}
		t, err := parseTimeValue(value)
		if err != nil {
			return fmt.Errorf("error parsing time: %w", err)
		}
		structFieldValue.Set(reflect.ValueOf(t))
		return nil
	}

	// Handle boolean types
	if fieldKind == reflect.Bool {
		if value != "true" && value != "false" {
			return fmt.Errorf("invalid boolean value: %s (must be 'true' or 'false')", value)
		}
		structFieldValue.SetBool(value == "true")
		return nil
	}

	// Handle signed integers
	if fieldKind >= reflect.Int && fieldKind <= reflect.Int64 {
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 20 {
			return fmt.Errorf("integer value too long: %d characters", len(cleanedValue))
		}
		n, err := strconv.ParseInt(cleanedValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %w", err)
		}
		// Check bounds for specific int types
		switch fieldKind {
		case reflect.Int8:
			if n < -128 || n > 127 {
				return fmt.Errorf("value %d out of range for int8", n)
			}
		case reflect.Int16:
			if n < -32768 || n > 32767 {
				return fmt.Errorf("value %d out of range for int16", n)
			}
		case reflect.Int32:
			if n < -2147483648 || n > 2147483647 {
				return fmt.Errorf("value %d out of range for int32", n)
			}
		}
		structFieldValue.SetInt(n)
		return nil
	}

	// Handle unsigned integers
	if fieldKind >= reflect.Uint && fieldKind <= reflect.Uint64 {
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 20 {
			return fmt.Errorf("unsigned integer value too long: %d characters", len(cleanedValue))
		}
		n, err := strconv.ParseUint(cleanedValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %w", err)
		}
		// Check bounds for specific uint types
		switch fieldKind {
		case reflect.Uint8:
			if n > 255 {
				return fmt.Errorf("value %d out of range for uint8", n)
			}
		case reflect.Uint16:
			if n > 65535 {
				return fmt.Errorf("value %d out of range for uint16", n)
			}
		case reflect.Uint32:
			if n > 4294967295 {
				return fmt.Errorf("value %d out of range for uint32", n)
			}
		}
		structFieldValue.SetUint(n)
		return nil
	}

	// Handle floating point numbers
	if fieldKind == reflect.Float32 || fieldKind == reflect.Float64 {
		cleanedValue := strings.ReplaceAll(value, "_", "")
		if len(cleanedValue) > 50 {
			return fmt.Errorf("float value too long: %d characters", len(cleanedValue))
		}
		f, err := strconv.ParseFloat(cleanedValue, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %w", err)
		}
		// Check bounds for float32
		if fieldKind == reflect.Float32 {
			if f > 3.40282346638528859811704183484516925440e+38 || f < -3.40282346638528859811704183484516925440e+38 {
				return fmt.Errorf("value %g out of range for float32", f)
			}
		}
		structFieldValue.SetFloat(f)
		return nil
	}

	// Handle pointer types
	if fieldKind == reflect.Ptr {
		elemType := fieldType.Elem()
		// Create a new value of the element type
		elemValue := reflect.New(elemType).Elem()
		// Recursively set the element value
		if err := setFieldValue(&elemValue, value); err != nil {
			return err
		}
		structFieldValue.Set(elemValue.Addr())
		return nil
	}

	// Try type conversion for other types (e.g., type aliases)
	val := reflect.ValueOf(value)
	if val.Type().ConvertibleTo(fieldType) {
		structFieldValue.Set(val.Convert(fieldType))
		return nil
	}

	return fmt.Errorf("cannot convert string to %s", fieldType)
}

func (ctx *Context) Action(uid string, action Callable) **Callable {
	if ctx.App == nil {
		panic("App is nil, cannot register component. Did you set the App field in Context?")
	}

	return ctx.App.Action(uid, action)
}

func (ctx *Context) Callable(action Callable) **Callable {
	if ctx.App == nil {
		panic("App is nil, cannot create callable. Did you set the App field in Context?")
	}

	return ctx.App.Callable(action)
}

// htmlToJSElement converts an HTML string to a JSElement JSON structure
func htmlToJSElement(htmlStr string) (*JSElement, error) {
	// Parse the HTML string
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, err
	}

	// Find the first meaningful content element (skip html, head, body, DOCTYPE, comments)
	var findContentElement func(*html.Node) *html.Node
	findContentElement = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode {
			// Skip document structure tags - we want the actual content
			if n.Data == "html" || n.Data == "head" || n.Data == "body" {
				// Look inside these tags for content
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if elem := findContentElement(c); elem != nil {
						return elem
					}
				}
				return nil
			}
			// This is a content element
			return n
		}
		// For non-element nodes, keep searching
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if elem := findContentElement(c); elem != nil {
				return elem
			}
		}
		return nil
	}

	rootElement := findContentElement(doc)
	if rootElement == nil {
		// If no element found, return a text node wrapped in a span
		return &JSElement{
			T: "span",
			C: []interface{}{htmlStr},
		}, nil
	}

	return nodeToJSElement(rootElement), nil
}

// nodeToJSElement recursively converts an html.Node to JSElement
func nodeToJSElement(n *html.Node) *JSElement {
	if n.Type != html.ElementNode {
		return nil
	}

	elem := &JSElement{
		T: n.Data,
		A: make(map[string]string),
		E: make(map[string]*JSEvent),
		C: []interface{}{},
	}

	// Process attributes
	for _, attr := range n.Attr {
		switch attr.Key {
		case "onclick":
			if event := parseEventHandler(attr.Val, "click"); event != nil {
				if event.Act == "raw" {
					// Keep raw JavaScript as attribute so querySelector('[onclick]') works
					elem.A["onclick"] = attr.Val
				} else {
					elem.E["click"] = event
				}
			}
		case "onchange":
			if event := parseEventHandler(attr.Val, "change"); event != nil {
				if event.Act == "raw" {
					// Keep raw JavaScript as attribute
					elem.A["onchange"] = attr.Val
				} else {
					elem.E["change"] = event
				}
			}
		case "onsubmit":
			if event := parseEventHandler(attr.Val, "submit"); event != nil {
				if event.Act == "raw" {
					// Keep raw JavaScript as attribute
					elem.A["onsubmit"] = attr.Val
				} else {
					elem.E["submit"] = event
				}
			}
		case "disabled", "readonly", "checked", "selected", "required":
			// Boolean attributes - store with empty string or the value if present
			elem.A[attr.Key] = attr.Val
		default:
			elem.A[attr.Key] = attr.Val
		}
	}

	// Clean up empty maps
	if len(elem.A) == 0 {
		elem.A = nil
	}
	if len(elem.E) == 0 {
		elem.E = nil
	}

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text := c.Data
			if text != "" {
				elem.C = append(elem.C, text)
			}
		} else if c.Type == html.ElementNode {
			if child := nodeToJSElement(c); child != nil {
				elem.C = append(elem.C, child)
			}
		}
	}

	if len(elem.C) == 0 {
		elem.C = nil
	}

	return elem
}

// parseEventHandler parses event handler strings like "__post(...)" or "__submit(...)"
// and converts them to JSEvent structures. For raw JavaScript handlers, returns a JSEvent
// with Act="raw" to preserve client-side interactivity.
func parseEventHandler(handler string, eventType string) *JSEvent {
	handler = strings.TrimSpace(handler)

	if handler == "" {
		return nil
	}

	// Match __post(event, "swap", "target_id", "path", values)
	rePost := regexp.MustCompile(`__post\s*\(\s*event\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*(\[.*?\])\s*\)`)
	if matches := rePost.FindStringSubmatch(handler); len(matches) == 5 {
		event := &JSEvent{
			Act:  "post",
			Swap: matches[1],
			Tgt:  matches[2],
			Path: matches[3],
		}
		// Parse values array
		var vals []BodyItem
		if err := json.Unmarshal([]byte(matches[4]), &vals); err == nil {
			event.Vals = vals
		}
		return event
	}

	// Match __submit(event, "swap", "target_id", "path", values)
	reSubmit := regexp.MustCompile(`__submit\s*\(\s*event\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*(\[.*?\])\s*\)`)
	if matches := reSubmit.FindStringSubmatch(handler); len(matches) == 5 {
		event := &JSEvent{
			Act:  "form",
			Swap: matches[1],
			Tgt:  matches[2],
			Path: matches[3],
		}
		// Parse values array
		var vals []BodyItem
		if err := json.Unmarshal([]byte(matches[4]), &vals); err == nil {
			event.Vals = vals
		}
		return event
	}

	// For raw JavaScript handlers (like those used in IRadioButtons, IRadioDiv),
	// preserve them so client-side interactivity works after DOM updates
	return &JSEvent{
		Act: "raw",
		JS:  handler,
	}
}

func (ctx *Context) Post(as ActionType, swap Swap, action *Action) string {
	ctx.App.storedMu.Lock()
	path, ok := ctx.App.stored[action.Method]
	ctx.App.storedMu.Unlock()

	if !ok {
		funcName := reflect.ValueOf(*action.Method).String()
		panic(fmt.Sprintf("Function '%s' probably not registered. Cannot make call to this function.", funcName))
	}

	var body []BodyItem

	for _, item := range action.Values {
		v := reflect.ValueOf(item)

		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		for i := range v.NumField() {
			field := v.Field(i)
			fieldName := v.Type().Field(i).Name
			fieldType := field.Type().Name()
			fieldValue := field.Interface()

			// Handle time.Time specially to avoid parsing issues
			if field.Type().String() == "time.Time" {
				if t, ok := fieldValue.(time.Time); ok {
					fieldValue = t.Format("2006-01-02 15:04:05 -0700 UTC")
				}
			}

			body = append(body, BodyItem{
				Name:  fieldName,
				Type:  fieldType,
				Value: fmt.Sprintf("%v", fieldValue),
			})
		}
	}

	values := "[]"

	if len(body) > 0 {
		temp, err := json.Marshal(body)

		if err == nil {
			values = string(temp)
		}
	}

	if as == FORM {
		return Trim(fmt.Sprintf(`__submit(event, "%s", "%s", "%s", %s) `, escapeJS(string(swap)), escapeJS(action.Target.ID), escapeJS(path), values))
	}

	return Trim(fmt.Sprintf(`__post(event, "%s", "%s", "%s", %s) `, escapeJS(string(swap)), escapeJS(action.Target.ID), escapeJS(path), values))
}

type Actions struct {
	Render  func(target Attr) string
	Replace func(target Attr) string
	Append  func(target Attr) string
	Prepend func(target Attr) string
	None    func() string
	// AsSubmit func(target Attr, swap ...Swap) Attr
	// AsClick  func(target Attr, swap ...Swap) Attr
}

type Submits struct {
	Render  func(target Attr) Attr
	Replace func(target Attr) Attr
	Append  func(target Attr) Attr
	Prepend func(target Attr) Attr
	None    func() Attr
}

// func swapize(swap ...Swap) Swap {
// 	if len(swap) > 0 {
// 		return swap[0]
// 	}
// 	return INLINE
// }

func (ctx *Context) Submit(method Callable, values ...any) Submits {
	callable := ctx.Callable(method)

	return Submits{
		Render: func(target Attr) Attr {
			return Attr{OnSubmit: ctx.Post(FORM, INLINE, &Action{Method: *callable, Target: target, Values: values})}
		},
		Replace: func(target Attr) Attr {
			return Attr{OnSubmit: ctx.Post(FORM, OUTLINE, &Action{Method: *callable, Target: target, Values: values})}
		},
		Append: func(target Attr) Attr {
			return Attr{OnSubmit: ctx.Post(FORM, APPEND, &Action{Method: *callable, Target: target, Values: values})}
		},
		Prepend: func(target Attr) Attr {
			return Attr{OnSubmit: ctx.Post(FORM, PREPEND, &Action{Method: *callable, Target: target, Values: values})}
		},
		None: func() Attr {
			return Attr{OnSubmit: ctx.Post(FORM, NONE, &Action{Method: *callable, Values: values})}
		},
	}
}

func (ctx *Context) Click(method Callable, values ...any) Submits {
	callable := ctx.Callable(method)

	return Submits{
		Render: func(target Attr) Attr {
			return Attr{OnClick: ctx.Post(POST, INLINE, &Action{Method: *callable, Target: target, Values: values})}
		},
		Replace: func(target Attr) Attr {
			return Attr{OnClick: ctx.Post(POST, OUTLINE, &Action{Method: *callable, Target: target, Values: values})}
		},
		Append: func(target Attr) Attr {
			return Attr{OnClick: ctx.Post(POST, APPEND, &Action{Method: *callable, Target: target, Values: values})}
		},
		Prepend: func(target Attr) Attr {
			return Attr{OnClick: ctx.Post(POST, PREPEND, &Action{Method: *callable, Target: target, Values: values})}
		},
		None: func() Attr {
			return Attr{OnClick: ctx.Post(POST, NONE, &Action{Method: *callable, Values: values})}
		},
	}
}

func (ctx *Context) Send(method Callable, values ...any) Actions {
	callable := ctx.Callable(method)

	return Actions{
		Render: func(target Attr) string {
			return ctx.Post(FORM, INLINE, &Action{Method: *callable, Target: target, Values: values})
		},
		Replace: func(target Attr) string {
			return ctx.Post(FORM, OUTLINE, &Action{Method: *callable, Target: target, Values: values})
		},
		Append: func(target Attr) string {
			return ctx.Post(FORM, APPEND, &Action{Method: *callable, Target: target, Values: values})
		},
		Prepend: func(target Attr) string {
			return ctx.Post(FORM, PREPEND, &Action{Method: *callable, Target: target, Values: values})
		},
		None: func() string {
			return ctx.Post(FORM, NONE, &Action{Method: *callable, Values: values})
		},
		// AsSubmit: func(target Attr, swap ...Swap) Attr {
		// 	return Attr{OnSubmit: ctx.Post(FORM, swapize(swap...), &Action{Method: *method, Target: target, Values: values})}
		// },
		// AsClick: func(target Attr, swap ...Swap) Attr {
		// 	return Attr{OnClick: ctx.Post(FORM, swapize(swap...), &Action{Method: *method, Target: target, Values: values})}
		// },
	}
}

func (ctx *Context) Call(method Callable, values ...any) Actions {
	callable := ctx.Callable(method)

	return Actions{
		Render: func(target Attr) string {
			return ctx.Post(POST, INLINE, &Action{Method: *callable, Target: target, Values: values})
		},
		Replace: func(target Attr) string {
			return ctx.Post(POST, OUTLINE, &Action{Method: *callable, Target: target, Values: values})
		},
		Append: func(target Attr) string {
			return ctx.Post(POST, APPEND, &Action{Method: *callable, Target: target, Values: values})
		},
		Prepend: func(target Attr) string {
			return ctx.Post(POST, PREPEND, &Action{Method: *callable, Target: target, Values: values})
		},
		None: func() string {
			return ctx.Post(POST, NONE, &Action{Method: *callable, Values: values})
		},
		// AsSubmit: func(target Attr, swap ...Swap) Attr {
		// 	return Attr{OnSubmit: ctx.Post(POST, swapize(swap...), &Action{Method: *method, Target: target, Values: values})}
		// },
		// AsClick: func(target Attr, swap ...Swap) Attr {
		// 	return Attr{OnClick: ctx.Post(POST, swapize(swap...), &Action{Method: *method, Target: target, Values: values})}
		// },
	}
}

func (ctx *Context) Load(href string) Attr {
	return Attr{
		Href:    href,
		OnClick: Trim(fmt.Sprintf(`event.preventDefault();if(window.__router&&window.__router.navigate){window.__router.navigate("%s");}else{__load("%s");}`, escapeJS(href), escapeJS(href))),
	}
}

// Reload adds a reload operation that will reload the page on the client
func (ctx *Context) Reload() string {
	ctx.ops = append(ctx.ops, &JSPatchOp{
		Op: "reload",
	})

	return ""
}

// Redirect adds a redirect operation that will navigate to the specified URL
func (ctx *Context) Redirect(href string) string {
	ctx.ops = append(ctx.ops, &JSPatchOp{
		Op:   "redirect",
		Href: href,
	})

	return ""
}

// Deferred fragments removed. The previous ctx.Defer(...) builder and helpers
// are no longer available.

func displayMessage(ctx *Context, message string, variant string) {
	// Add notification operation to context
	ctx.ops = append(ctx.ops, &JSPatchOp{
		Op:      "notify",
		Msg:     message,
		Variant: variant,
	})
}

// displayError renders an error toast with a Reload button.
func displayError(ctx *Context, message string) {
	ctx.ops = append(ctx.ops, &JSPatchOp{
		Op:      "notify",
		Msg:     message,
		Variant: "error-reload",
	})
}

func (ctx *Context) Success(message string) string {
	displayMessage(ctx, message, "success")

	return ""
}

func (ctx *Context) Error(message string) string {
	displayMessage(ctx, message, "error")

	return ""
}

// ErrorReload shows an error toast with a Reload button.
func (ctx *Context) ErrorReload(message string) { displayError(ctx, message) }

func (ctx *Context) Info(message string) string {
	displayMessage(ctx, message, "info")

	return ""
}

// Title updates the page title dynamically
func (ctx *Context) Title(title string) {
	ctx.ops = append(ctx.ops, &JSPatchOp{
		Op:    "title",
		Title: title,
	})
}

// SetSecurityHeaders sets comprehensive security headers
func (ctx *Context) SetSecurityHeaders() {
	headers := ctx.Response.Header()

	// Content Security Policy (if not already set)
	if headers.Get("Content-Security-Policy") == "" {
		ctx.SetDefaultCSP()
	}

	// HTTP Strict Transport Security - enforce HTTPS
	headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

	// X-Frame-Options - prevent clickjacking
	headers.Set("X-Frame-Options", "DENY")

	// X-Content-Type-Options - prevent MIME type sniffing
	headers.Set("X-Content-Type-Options", "nosniff")

	// X-XSS-Protection - legacy XSS protection for older browsers
	headers.Set("X-XSS-Protection", "1; mode=block")

	// Referrer Policy - control referrer information
	headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Permissions Policy - restrict dangerous browser features
	headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
}

// SetCustomSecurityHeaders allows fine-grained control over security headers
func (ctx *Context) SetCustomSecurityHeaders(options SecurityHeaderOptions) {
	headers := ctx.Response.Header()

	if options.CSP != "" {
		headers.Set("Content-Security-Policy", options.CSP)
	}

	if options.HSTS != "" {
		headers.Set("Strict-Transport-Security", options.HSTS)
	} else if options.EnableHSTS {
		headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	if options.FrameOptions != "" {
		headers.Set("X-Frame-Options", options.FrameOptions)
	}

	if options.ContentTypeOptions {
		headers.Set("X-Content-Type-Options", "nosniff")
	}

	if options.XSSProtection != "" {
		headers.Set("X-XSS-Protection", options.XSSProtection)
	}

	if options.ReferrerPolicy != "" {
		headers.Set("Referrer-Policy", options.ReferrerPolicy)
	}

	if options.PermissionsPolicy != "" {
		headers.Set("Permissions-Policy", options.PermissionsPolicy)
	}
}

type SecurityHeaderOptions struct {
	CSP                string
	HSTS               string
	EnableHSTS         bool
	FrameOptions       string
	ContentTypeOptions bool
	XSSProtection      string
	ReferrerPolicy     string
	PermissionsPolicy  string
}

func (ctx *Context) DownloadAs(file *io.Reader, contentType string, name string) error {
	// Read the file content into a byte slice
	fileBytes, err := io.ReadAll(*file)
	if err != nil {
		log.Println(err)
		return err
	}

	// Encode the byte slice to a base64 string
	fileBase64 := base64.StdEncoding.EncodeToString(fileBytes)

	// Add download operation
	ctx.ops = append(ctx.ops, &JSPatchOp{
		Op:          "download",
		Data:        fileBase64,
		ContentType: contentType,
		Filename:    name,
	})

	return nil
}

func (ctx *Context) Translate(message string, val ...any) string {
	return fmt.Sprintf(message, val...)
}

// RandomString generates a cryptographically secure random string
func RandomString(n ...int) string {
	length := 20
	if len(n) > 0 && n[0] > 0 {
		length = n[0]
	}

	// Use base64 encoding for efficiency and URL safety
	// We need 3/4 * length bytes to get length characters after base64 encoding
	byteLength := ((length * 3) + 3) / 4

	b := make([]byte, byteLength)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based generation if crypto/rand fails
		// This is not ideal but ensures the function doesn't panic
		log.Printf("Warning: crypto/rand failed (%v), falling back to time-based generation\n", err)
		return fmt.Sprintf("fallback_%d_%d", time.Now().UnixNano(), length)
	}

	// Encode to base64 and trim to exact length
	encoded := base64.URLEncoding.EncodeToString(b)
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	// Replace any remaining padding characters with alphanumeric
	encoded = strings.ReplaceAll(encoded, "=", "X")
	encoded = strings.ReplaceAll(encoded, "-", "Y")
	encoded = strings.ReplaceAll(encoded, "_", "Z")

	return encoded
}

// SetCSP sets Content Security Policy headers to help prevent XSS attacks
func (ctx *Context) SetCSP(policy string) {
	ctx.Response.Header().Set("Content-Security-Policy", policy)
}

// SetDefaultCSP sets a restrictive CSP that allows only same-origin scripts and styles
func (ctx *Context) SetDefaultCSP() {
	policy := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss: ws:; frame-ancestors 'none';"
	ctx.SetCSP(policy)
}

// securityHeadersMiddleware adds comprehensive security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers before processing the request
		headers := w.Header()

		// HTTP Strict Transport Security - enforce HTTPS
		if isSecure(r) {
			headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// X-Frame-Options - prevent clickjacking
		headers.Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options - prevent MIME type sniffing
		headers.Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection - legacy XSS protection for older browsers
		headers.Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy - control referrer information
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - restrict dangerous browser features
		headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	if f, ok := w.writer.(*gzip.Writer); ok {
		_ = f.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacker not supported")
	}
	return hj.Hijack()
}

func (w *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(strings.ToLower(strings.Join(r.Header["Upgrade"], " ")), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		headers := w.Header()
		headers.Set("Content-Encoding", "gzip")
		headers.Add("Vary", "Accept-Encoding")
		headers.Del("Content-Length")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
	})
}

func cacheControlMiddleware(next http.Handler, maxAge time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Cache-Control header
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds())))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

type PWAIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"` // "any", "maskable", or "any maskable"
}

type PWAConfig struct {
	Name                  string    `json:"name"`
	ShortName             string    `json:"short_name"`
	ID                    string    `json:"id,omitempty"` // App ID - defaults to StartURL if empty
	Description           string    `json:"description,omitempty"`
	ThemeColor            string    `json:"theme_color,omitempty"`
	BackgroundColor       string    `json:"background_color,omitempty"`
	Display               string    `json:"display,omitempty"`
	StartURL              string    `json:"start_url,omitempty"`
	Icons                 []PWAIcon `json:"icons,omitempty"`
	GenerateServiceWorker bool      `json:"-"`
	CacheAssets           []string  `json:"-"` // Asset URLs to pre-cache, e.g., ["/assets/style.css"]
	OfflinePage           string    `json:"-"` // Optional offline fallback page URL
}

type Route struct {
	Path       string
	Title      string
	Handler    *Callable
	Pattern    string   // original pattern like "/vehicles/edit/{id}"
	Segments   []string // split segments for matching
	ParamNames []string // names of parameters in order
	HasParams  bool     // whether this route has path parameters
}

// CustomRoute represents a custom HTTP handler registered with the app
type CustomRoute struct {
	Method  string           // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path    string           // URL path
	Handler http.HandlerFunc // The handler function
}

type App struct {
	Lanugage          string
	ContentID         Attr
	BasePath          string // URL prefix for path-prefix mounting (e.g., "/admin")
	HTMLBody          func(string) string
	HTMLHead          []string
	DebugEnabled      bool
	pwaConfig         *PWAConfig
	pwaManifest       []byte
	swCacheKey        string // Generated on startup for cache versioning
	mux               *http.ServeMux
	muxOnce           sync.Once
	storedMu          sync.Mutex
	stored            map[*Callable]string
	sessMu            sync.Mutex
	sessions          map[string]*sessRec
	captchaSessions   map[string]*CaptchaSession
	captchaSessionsMu sync.RWMutex
	wsMu              sync.RWMutex
	wsClients         map[*websocket.Conn]*wsState
	assetHandlers     map[string]http.Handler
	routes            map[string]*Route // path → Route
	routesMu          sync.RWMutex      // mutex for route maps
	routesManifest    string
	routesCached      bool
	layout            Callable       // persistent layout function
	customRoutes      []*CustomRoute // custom HTTP handlers
	customRoutesMu    sync.RWMutex   // mutex for custom routes
}

type sessRec struct {
	lastSeen time.Time
	targets  map[string]func()
}

type wsState struct {
	lastPong time.Time
	sid      string
}

func (app *App) Register(httpMethod string, path string, method *Callable) string {
	if path == "" || method == nil {
		panic("Path and Method cannot be empty")
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(*method).Pointer()).Name()

	if funcName == "" {
		panic("Method cannot be empty")
	}

	_, ok := app.stored[method]
	if ok {
		panic("Method already registered: " + funcName)
	}

	for _, value := range app.stored {
		if value == path {
			panic("Path already registered: " + path)
		}
	}

	app.storedMu.Lock()
	app.stored[method] = path
	app.storedMu.Unlock()

	// fmt.Println("Registering: ", httpMethod, path, " -> ", funcName)

	return path
}

// Layout sets the persistent layout function that wraps all pages.
// The layout should render a content slot element with id="__content__".
func (app *App) Layout(handler Callable) {
	app.layout = handler
}

// parseRoutePattern parses a route path pattern and extracts segment information.
// Returns segments, parameter names, and whether the route has parameters.
func parseRoutePattern(pattern string) (segments []string, paramNames []string, hasParams bool) {
	segments = strings.Split(strings.Trim(pattern, "/"), "/")
	for _, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			paramName := strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}")
			paramNames = append(paramNames, paramName)
			hasParams = true
		}
	}
	return segments, paramNames, hasParams
}

// matchRoutePattern matches an actual path against a route pattern and extracts parameters.
// Returns the extracted parameters map if match succeeds, nil otherwise.
func matchRoutePattern(actualPath string, route *Route) map[string]string {
	if !route.HasParams {
		// Exact match only
		if actualPath == route.Path {
			return nil
		}
		return nil
	}

	actualSegments := strings.Split(strings.Trim(actualPath, "/"), "/")
	if len(actualSegments) != len(route.Segments) {
		return nil
	}

	params := make(map[string]string)
	for i, patternSeg := range route.Segments {
		actualSeg := actualSegments[i]
		if strings.HasPrefix(patternSeg, "{") && strings.HasSuffix(patternSeg, "}") {
			// This is a parameter segment
			paramName := strings.TrimSuffix(strings.TrimPrefix(patternSeg, "{"), "}")
			params[paramName] = actualSeg
		} else if patternSeg != actualSeg {
			// Literal segment doesn't match
			return nil
		}
	}

	return params
}

// matchRoute finds a route that matches the given path, trying exact match first, then pattern matching.
// Returns the matched route and extracted parameters.
func (app *App) matchRoute(path string) (*Route, map[string]string) {
	app.routesMu.RLock()
	defer app.routesMu.RUnlock()

	// First try exact match
	if route, exists := app.routes[path]; exists {
		return route, nil
	}

	// Then try pattern matching
	for _, route := range app.routes {
		if params := matchRoutePattern(path, route); params != nil {
			return route, params
		}
	}

	return nil, nil
}

func (app *App) routeManifestScript() string {
	app.routesMu.RLock()
	if app.routesCached {
		cached := app.routesManifest
		app.routesMu.RUnlock()
		return cached
	}
	app.routesMu.RUnlock()

	app.routesMu.Lock()
	defer app.routesMu.Unlock()

	if app.routesCached {
		return app.routesManifest
	}

	routeManifest := make(map[string]interface{}, len(app.routes))
	for path, route := range app.routes {
		if route.HasParams {
			routeManifest[path] = map[string]interface{}{
				"path":    path,
				"pattern": true,
			}
		} else {
			routeManifest[path] = path
		}
	}

	manifestJSON, err := json.Marshal(routeManifest)
	if err != nil {
		log.Printf("Error marshaling route manifest: %v", err)
		manifestJSON = []byte("{}")
	}

	app.routesManifest = fmt.Sprintf(`<script>window.__routes = %s;</script>`, string(manifestJSON))
	app.routesCached = true

	return app.routesManifest
}

// Page registers a route with a title and handler.
// Usage: Page("/", "Page Title", handler)
// Supports path parameters: Page("/vehicles/edit/{id}", "Edit Vehicle", handler)
func (app *App) Page(path string, title string, handler Callable) {
	if path == "" {
		panic("Page path cannot be empty")
	}
	if title == "" {
		panic("Page title cannot be empty")
	}

	app.routesMu.Lock()
	defer app.routesMu.Unlock()

	// Check if path already exists
	if _, exists := app.routes[path]; exists {
		panic(fmt.Sprintf("Page path already registered: %s", path))
	}

	// Parse route pattern to extract segments and parameters
	segments, paramNames, hasParams := parseRoutePattern(path)

	// Create route
	route := &Route{
		Path:       path,
		Title:      title,
		Handler:    &handler,
		Pattern:    path,
		Segments:   segments,
		ParamNames: paramNames,
		HasParams:  hasParams,
	}

	// Store route
	app.routes[path] = route
	app.routesCached = false
}

// Custom registers a standard http.HandlerFunc for a specific path and HTTP method.
// This allows integrating regular HTTP endpoints (like REST APIs) alongside g-sui pages.
// Custom handlers are checked before g-sui routes, so they take priority.
//
// Example:
//
//	app.Custom("GET", "/api/health", func(w http.ResponseWriter, r *http.Request) {
//	    w.Header().Set("Content-Type", "application/json")
//	    w.Write([]byte(`{"status": "ok"}`))
//	})
//
//	app.Custom("POST", "/api/users", createUserHandler)
func (app *App) Custom(method string, path string, handler http.HandlerFunc) {
	if method == "" {
		panic("Custom: method cannot be empty")
	}
	if path == "" {
		panic("Custom: path cannot be empty")
	}
	if handler == nil {
		panic("Custom: handler cannot be nil")
	}

	app.customRoutesMu.Lock()
	defer app.customRoutesMu.Unlock()

	// Check for duplicate registration
	for _, route := range app.customRoutes {
		if route.Method == method && route.Path == path {
			panic(fmt.Sprintf("Custom: route %s %s already registered", method, path))
		}
	}

	app.customRoutes = append(app.customRoutes, &CustomRoute{
		Method:  strings.ToUpper(method),
		Path:    path,
		Handler: handler,
	})
}

// GET registers a custom handler for GET requests.
// Shorthand for Custom("GET", path, handler).
func (app *App) GET(path string, handler http.HandlerFunc) {
	app.Custom("GET", path, handler)
}

// POST registers a custom handler for POST requests.
// Shorthand for Custom("POST", path, handler).
func (app *App) POST(path string, handler http.HandlerFunc) {
	app.Custom("POST", path, handler)
}

// PUT registers a custom handler for PUT requests.
// Shorthand for Custom("PUT", path, handler).
func (app *App) PUT(path string, handler http.HandlerFunc) {
	app.Custom("PUT", path, handler)
}

// DELETE registers a custom handler for DELETE requests.
// Shorthand for Custom("DELETE", path, handler).
func (app *App) DELETE(path string, handler http.HandlerFunc) {
	app.Custom("DELETE", path, handler)
}

// PATCH registers a custom handler for PATCH requests.
// Shorthand for Custom("PATCH", path, handler).
func (app *App) PATCH(path string, handler http.HandlerFunc) {
	app.Custom("PATCH", path, handler)
}

// Debug enables or disables server debug logging.
// When enabled, debug logs are printed with the "gsui:" prefix.
func (app *App) Debug(enable bool) {
	app.DebugEnabled = enable
}

func (app *App) debugf(format string, args ...any) {
	if !app.DebugEnabled {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	log.Printf("gsui: "+format, args...)
}

func (app *App) Action(uid string, action Callable) **Callable {
	if !strings.HasPrefix(uid, eventPath) {
		uid = eventPath + uid
	}

	uid = strings.ToLower(uid)

	for key, value := range app.stored {
		if value == uid {
			return &key
		}
	}

	found := &action
	app.Register("POST", uid, found)

	return &found
}

func (app *App) Callable(action Callable) **Callable {
	uid := runtime.FuncForPC(reflect.ValueOf(action).Pointer()).Name()
	uid = strings.ToLower(uid)
	uid = reRemoveChars.ReplaceAllString(uid, "")
	uid = reReplaceChars.ReplaceAllString(uid, "-")

	if !strings.HasPrefix(uid, eventPath) {
		uid = eventPath + uid
	}

	// Check if already registered - update the callable if found
	app.storedMu.Lock()
	for key, value := range app.stored {
		if value == uid {
			// Update the callable to the new instance's method
			// This ensures stateful handlers (like collate methods) use the latest instance
			*key = action
			app.storedMu.Unlock()
			return &key
		}
	}
	app.storedMu.Unlock()

	found := &action
	app.Register("POST", uid, found)

	return &found
}

// mimeTypeMiddleware wraps a handler to set correct MIME types for common web file types
func mimeTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		switch ext {
		// JavaScript/TypeScript
		case ".js", ".mjs":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".ts", ".mts":
			w.Header().Set("Content-Type", "text/typescript; charset=utf-8")
		case ".jsx":
			w.Header().Set("Content-Type", "text/jsx; charset=utf-8")
		case ".tsx":
			w.Header().Set("Content-Type", "text/tsx; charset=utf-8")
		// Stylesheets
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		// JSON
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// HTML
		case ".html", ".htm":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// XML
		case ".xml":
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		// Images
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".webp":
			w.Header().Set("Content-Type", "image/webp")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".avif":
			w.Header().Set("Content-Type", "image/avif")
		// Fonts
		case ".woff":
			w.Header().Set("Content-Type", "font/woff")
		case ".woff2":
			w.Header().Set("Content-Type", "font/woff2")
		case ".ttf":
			w.Header().Set("Content-Type", "font/ttf")
		case ".otf":
			w.Header().Set("Content-Type", "font/otf")
		case ".eot":
			w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
		// Audio/Video
		case ".mp3":
			w.Header().Set("Content-Type", "audio/mpeg")
		case ".wav":
			w.Header().Set("Content-Type", "audio/wav")
		case ".ogg":
			w.Header().Set("Content-Type", "audio/ogg")
		case ".mp4":
			w.Header().Set("Content-Type", "video/mp4")
		case ".webm":
			w.Header().Set("Content-Type", "video/webm")
		// Documents
		case ".pdf":
			w.Header().Set("Content-Type", "application/pdf")
		// Web manifest
		case ".webmanifest":
			w.Header().Set("Content-Type", "application/manifest+json")
		// Source maps
		case ".map":
			w.Header().Set("Content-Type", "application/json")
		// WebAssembly
		case ".wasm":
			w.Header().Set("Content-Type", "application/wasm")
		// Plain text
		case ".txt", ".md":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

func (app *App) Assets(assets embed.FS, path string, maxAge time.Duration) {
	path = strings.TrimPrefix(path, "/")
	handler := http.FileServer(http.FS(assets))
	wrappedHandler := mimeTypeMiddleware(cacheControlMiddleware(handler, maxAge))

	// Initialize the map if it doesn't exist
	if app.assetHandlers == nil {
		app.assetHandlers = make(map[string]http.Handler)
	}

	// Store the handler for this path prefix
	app.assetHandlers["/"+path+"/"] = wrappedHandler
}

// Favicon serves a favicon file from the embedded filesystem at /favicon.ico.
// The path parameter should be the path to the favicon file in the embed.FS
// (e.g., "assets/favicon.ico", "assets/favicon.svg").
// Defaults to "favicon.ico" if not provided.
func (app *App) Favicon(assets embed.FS, path string, maxAge time.Duration) {
	if path == "" {
		path = "favicon.ico"
	}

	app.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		file, err := assets.ReadFile(path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Set proper content type based on file extension.
		// Some browsers (and Go's DetectContentType) don't reliably detect SVG,
		// so prefer extension-based mapping for correctness.
		switch strings.ToLower(filepath.Ext(path)) {
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".webp":
			w.Header().Set("Content-Type", "image/webp")
		case ".avif":
			w.Header().Set("Content-Type", "image/avif")
		default:
			// Fallback to detection if unknown extension
			w.Header().Set("Content-Type", http.DetectContentType(file))
		}

		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds())))
		w.Write(file)
	})
}

func makeContext(app *App, r *http.Request, w http.ResponseWriter) *Context {
	var sessionID string

	// Prefer tsui__sid (t-sui compatible); fallback to legacy session_id if present
	if c, err := r.Cookie("tsui__sid"); err == nil {
		sessionID = c.Value
	} else if c2, err2 := r.Cookie("session_id"); err2 == nil {
		sessionID = c2.Value
	} else {
		sessionID = RandomString(30)
	}

	// Ensure cookie is set with Lax and conditional Secure
	if sessionID == "" {
		sessionID = RandomString(30)
	}
	// Always (re)issue cookie so attributes are updated
	http.SetCookie(w, &http.Cookie{
		Name:     "tsui__sid",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure(r),
		SameSite: http.SameSiteLaxMode,
	})

	// Track last-seen per session
	if app != nil {
		app.sessMu.Lock()
		if app.sessions == nil {
			app.sessions = make(map[string]*sessRec)
		}
		rec := app.sessions[sessionID]
		if rec == nil {
			rec = &sessRec{lastSeen: time.Now(), targets: make(map[string]func())}
			app.sessions[sessionID] = rec
		} else {
			rec.lastSeen = time.Now()
		}
		app.sessMu.Unlock()
	}

	return &Context{
		App:       app,
		Request:   r,
		Response:  w,
		SessionID: sessionID,
		append:    []string{},
	}
}

// makeContextForWS creates a Context from WebSocket data
func makeContextForWS(app *App, sessionID string, vals []BodyItem) (*Context, error) {
	// Separate file items from regular items
	var fileItems []BodyItem
	var regularItems []BodyItem
	for _, item := range vals {
		if item.Type == "file" {
			fileItems = append(fileItems, item)
		} else {
			regularItems = append(regularItems, item)
		}
	}

	// Create JSON body for regular fields only
	bodyJSON, err := json.Marshal(regularItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body items: %w", err)
	}

	r := &http.Request{
		Method: "POST",
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewReader(bodyJSON)),
	}
	r.Header.Set("Content-Type", "application/json")

	w := &wsResponseWriter{}
	ctx := makeContext(app, r, w)
	ctx.fileItems = fileItems // Store file items for ctx.File()/ctx.Files()

	if sessionID != "" {
		ctx.SessionID = sessionID
	}

	return ctx, nil
}

// wsResponseWriter is a minimal ResponseWriter for WebSocket contexts
type wsResponseWriter struct {
	header http.Header
}

func (w *wsResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}
func (w *wsResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (w *wsResponseWriter) WriteHeader(int)           {}

// setup registers the WebSocket and main handler on app.mux. It is
// idempotent — subsequent calls are no-ops thanks to sync.Once.
// Routes are always registered at "/" on the internal mux; when mounted
// on an external mux via Mount(), StripPrefix removes the BasePath.
func (app *App) setup() {
	app.muxOnce.Do(func() {
		// If BasePath is set, regenerate the WS script tag in HTMLHead
		if app.BasePath != "" {
			newScript := Script(__stringify, __loader, __offline, __error, __notify, __e, __engine, __post, __submit, __load, __router, __theme, __cfmt, __debounce, __clipboard, __capi, __cel, __cregister, __cfilter, __cfilterbar, __cpagination, __caction, __ctable, __cchart, __ckpi, __cfileupload, __confirm, __toast, __cmodal, __cbadge, __clientScript, wsScript(app.BasePath))
			// Replace the last entry in HTMLHead (which is the script tag)
			if len(app.HTMLHead) > 0 {
				app.HTMLHead[len(app.HTMLHead)-1] = newScript
			}
		}
		app.initWS()
		app.mux.Handle("/", app.buildHandler())
	})
}

func (app *App) Listen(port string) {
	log.Println("Listening on http://0.0.0.0" + port)

	// Start session sweeper in background
	app.StartSweeper()

	app.setup()

	if err := http.ListenAndServe(port, app.mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println("Error:", err)
	}
}

// Handler returns an http.Handler that can be used in custom server configurations.
// This allows wrapping the g-sui handler with custom middleware or integrating
// with existing HTTP server setups.
//
// Example:
//
//	app := ui.MakeApp("en")
//	app.Page("/", "Home", homeHandler)
//	app.StartSweeper()
//
//	// Use with custom server
//	server := &http.Server{
//	    Addr:    ":8080",
//	    Handler: app.Handler(),
//	}
//	server.ListenAndServe()
func (app *App) Handler() http.Handler {
	app.setup()
	return app.mux
}

// Mount registers the app's routes on an external mux under the given prefix.
// This enables running multiple g-sui apps on a single HTTP server.
//
// Example:
//
//	mux := http.NewServeMux()
//	adminApp := ui.MakeApp("en")
//	adminApp.Page("/", "Admin", adminHandler)
//	adminApp.Mount("/admin", mux)
//
//	portalApp := ui.MakeApp("en")
//	portalApp.Page("/", "Portal", portalHandler)
//	portalApp.Mount("/portal", mux)
//
//	http.ListenAndServe(":8080", mux)
func (app *App) Mount(prefix string, mux *http.ServeMux) {
	app.BasePath = prefix
	app.StartSweeper()
	app.setup()
	mux.Handle(prefix+"/", http.StripPrefix(prefix, app.mux))
}

// buildHandler creates the main HTTP handler for the app.
// This is used internally by both Listen() and Handler().
func (app *App) buildHandler() http.Handler {
	bodyTemplate := app.HTMLBody("")
	bodyTemplate = strings.ReplaceAll(bodyTemplate, "__lang__", app.Lanugage)
	bodyTemplate = Trim(bodyTemplate)

	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.URL.Path

		if strings.Contains(strings.Join(r.Header["Upgrade"], " "), "websocket") {
			// Let explicit WS handlers handle upgrades
			http.NotFound(w, r)
			return
		}

		// Check if this is an asset request
		for prefix, handler := range app.assetHandlers {
			if strings.HasPrefix(value, prefix) {
				handler.ServeHTTP(w, r)
				return
			}
		}

		// Check custom handlers first (before g-sui routes)
		app.customRoutesMu.RLock()
		for _, route := range app.customRoutes {
			if route.Method == r.Method && route.Path == value {
				app.customRoutesMu.RUnlock()
				route.Handler(w, r)
				return
			}
		}
		app.customRoutesMu.RUnlock()

		// For non-custom routes, only allow GET, POST, HEAD
		if !strings.Contains("GET POST HEAD", r.Method) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Handle POST requests for SPA routing (fetch page content as JSON)
		if r.Method == "POST" {
			// Try to get actual path from request body or query string
			var actualPath string
			if r.Body != nil {
				var bodyData struct {
					Path string `json:"path"`
				}
				if err := json.NewDecoder(r.Body).Decode(&bodyData); err == nil && bodyData.Path != "" {
					actualPath = bodyData.Path
				}
			}

			// Fallback: use the path from URL if not in body
			if actualPath == "" {
				actualPath = r.URL.Query().Get("path")
			}

			// If still empty, use the URL path
			if actualPath == "" {
				actualPath = value
			}

			// Parse query parameters from actualPath and strip for route matching
			pathForMatching := actualPath
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

			// Use pattern matching to find route and extract parameters
			route, pathParams := app.matchRoute(pathForMatching)

			if route == nil {
				// Not a page route, check for special routes (manifest, service worker)
				app.storedMu.Lock()
				for found, path := range app.stored {
					if path == value {
						app.storedMu.Unlock()
						ctx := makeContext(app, r, w)
						defer func() {
							if rec := recover(); rec != nil {
								log.Println("handler panic recovered:", rec)
								w.WriteHeader(http.StatusInternalServerError)
								w.Write([]byte(app.devErrorPage()))
							}
						}()
						html := (*found)(ctx)
						if len(html) > 0 {
							w.Write([]byte(html))
						}
						return
					}
				}
				app.storedMu.Unlock()

				http.Error(w, "Page not found", http.StatusNotFound)
				return
			}

			ctx := makeContext(app, r, w)
			// Set path parameters in context
			if pathParams != nil {
				ctx.pathParams = pathParams
			}
			// Set query parameters in context (from SPA navigation path)
			if queryParams != nil {
				ctx.queryParams = queryParams
			}

			defer func() {
				if rec := recover(); rec != nil {
					log.Println("handler panic recovered:", rec)
					w.WriteHeader(http.StatusInternalServerError)
					errorHTML := app.devErrorPage()
					jsElement, err := htmlToJSElement(errorHTML)
					if err != nil {
						jsElement = &JSElement{T: "span", C: []interface{}{"Error"}}
					}
					response := JSHTTPResponse{El: jsElement}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					json.NewEncoder(w).Encode(response)
				}
			}()

			app.debugf("page fetch %s", route.Path)
			html := (*route.Handler)(ctx)
			if len(ctx.append) > 0 {
				html += strings.Join(ctx.append, "")
			}

			// Convert HTML to JSON element structure
			jsElement, err := htmlToJSElement(html)
			if err != nil {
				log.Printf("Error converting HTML to JSON: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsElement = &JSElement{T: "span", C: []interface{}{"Error"}}
			}

			// Add title operation
			ops := ctx.ops
			if route.Title != "" {
				ops = append(ops, &JSPatchOp{
					Op:    "title",
					Title: route.Title,
				})
			}

			response := JSHTTPResponse{El: jsElement, Ops: ops}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}

			return
		}

		// Handle GET requests for pages
		if r.Method == "GET" {
			// Try to match route (exact or pattern)
			route, pathParams := app.matchRoute(value)

			// Only render shell for registered routes
			if route != nil {
				ctx := makeContext(app, r, w)
				// Set path parameters in context
				if pathParams != nil {
					ctx.pathParams = pathParams
				}

				defer func() {
					if rec := recover(); rec != nil {
						log.Println("handler panic recovered:", rec)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(app.devErrorPage()))
					}
				}()

				manifestScript := app.routeManifestScript()

				var bodyHTML string
				if app.layout != nil {
					// Render with layout shell
					bodyHTML = app.layout(ctx) + manifestScript
				} else {
					// No layout - render page content directly in a content div
					pageHTML := (*route.Handler)(ctx)
					bodyHTML = fmt.Sprintf(`<div id="%s">%s</div>%s`, app.ContentID.ID, pageHTML, manifestScript)
				}

				// Build full HTML page
				head := []string{
					fmt.Sprintf(`<title>%s</title>`, route.Title),
				}
				head = append(head, app.HTMLHead...)
				// Ensure Material Icons CSS is applied
				head = append(head, `<style>.material-icons{font-family:'Material Icons';font-weight:normal;font-style:normal;font-size:24px;line-height:1;letter-spacing:normal;text-transform:none;display:inline-block;white-space:nowrap;word-wrap:normal;direction:ltr;-webkit-font-feature-settings:'liga';-webkit-font-smoothing:antialiased;}</style>`)

				html := bodyTemplate
				html = strings.ReplaceAll(html, "__head__", strings.Join(head, " "))
				html = strings.ReplaceAll(html, "__body__", bodyHTML)

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(html))
				return
			}
		}

		// Callable actions are now handled via WebSocket, not HTTP POST
		// HTTP POST is used for SPA routing to fetch page content as JSON

		// Serve static handlers (manifest, service worker) via GET
		if r.Method == "GET" && (value == "/manifest.webmanifest" || value == "/sw.js") {
			app.storedMu.Lock()
			for found, path := range app.stored {
				if path == value {
					app.storedMu.Unlock()
					ctx := makeContext(app, r, w)
					defer func() {
						if rec := recover(); rec != nil {
							log.Println("handler panic recovered:", rec)
							w.WriteHeader(http.StatusInternalServerError)
							w.Write([]byte(app.devErrorPage()))
						}
					}()
					html := (*found)(ctx)
					if len(html) > 0 {
						w.Write([]byte(html))
					}
					return
				}
			}
			app.storedMu.Unlock()
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	return securityHeadersMiddleware(gzipMiddleware(mainHandler))
}

// TestHandler returns an http.Handler that uses g-sui's routing logic.
// This is intended for testing purposes to allow test servers to
// use the same routing as production without starting a real HTTP server.
func (app *App) TestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains("GET POST", r.Method) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if strings.Contains(strings.Join(r.Header["Upgrade"], " "), "websocket") {
			http.NotFound(w, r)
			return
		}

		// Callable actions are now handled via WebSocket, not HTTP POST.
		// TestHandler only serves static handlers (manifest, service worker).
		if r.Method == "GET" && (r.URL.Path == "/manifest.webmanifest" || r.URL.Path == "/sw.js") {
			app.storedMu.Lock()
			for found, routePath := range app.stored {
				if routePath == r.URL.Path {
					app.storedMu.Unlock()
					ctx := makeContext(app, r, w)

					defer func() {
						if rec := recover(); rec != nil {
							log.Println("handler panic recovered:", rec)
							w.WriteHeader(http.StatusInternalServerError)
							w.Write([]byte(app.devErrorPage()))
						}
					}()

					html := (*found)(ctx)
					if len(html) > 0 {
						w.Write([]byte(html))
					}

					return
				}
			}
			app.storedMu.Unlock()
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})
}

// PWA enables Progressive Web App capabilities.
// Call this to generate manifest.webmanifest and optionally a service worker.
// The manifest will be served at /manifest.webmanifest
// The service worker will be served at /sw.js
func (app *App) PWA(config PWAConfig) {
	if config.StartURL == "" {
		config.StartURL = "/"
	}
	if config.Display == "" {
		config.Display = "standalone"
	}
	// Default ID to StartURL if not specified (per PWA best practices)
	if config.ID == "" {
		config.ID = config.StartURL
	}

	app.pwaConfig = &config

	// Generate manifest JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("gsui: failed to generate pwa manifest: %v", err)
	} else {
		app.pwaManifest = data
	}

	// Add HTML head tags
	app.HTMLHead = append(app.HTMLHead,
		`<link rel="manifest" href="/manifest.webmanifest">`,
		`<meta name="mobile-web-app-capable" content="yes">`,
		`<meta name="apple-mobile-web-app-capable" content="yes">`,
		`<meta name="apple-mobile-web-app-status-bar-style" content="default">`,
	)

	if config.ThemeColor != "" {
		app.HTMLHead = append(app.HTMLHead, fmt.Sprintf(`<meta name="theme-color" content="%s">`, config.ThemeColor))
	}

	// Register manifest route using the framework's routing system
	manifestHandler := func(ctx *Context) string {
		ctx.Response.Header().Set("Content-Type", "application/manifest+json")
		ctx.Response.Header().Set("Cache-Control", "public, max-age=3600")
		ctx.Response.Write(app.pwaManifest)
		return ""
	}
	app.storedMu.Lock()
	app.stored[&manifestHandler] = "/manifest.webmanifest"
	app.storedMu.Unlock()

	// Register service worker route if requested
	if config.GenerateServiceWorker {
		// Generate unique cache key based on startup time
		app.swCacheKey = fmt.Sprintf("app-%d", time.Now().Unix())

		app.HTMLHead = append(app.HTMLHead,
			`<script>
                if ('serviceWorker' in navigator) {
                    window.addEventListener('load', () => {
                        navigator.serviceWorker.register('/sw.js');
                    });
                }
            </script>`,
		)

		swHandler := func(ctx *Context) string {
			ctx.Response.Header().Set("Content-Type", "application/javascript")
			ctx.Response.Header().Set("Cache-Control", "no-cache")

			// Build assets array for JS
			assetsJSON := "[]"
			if len(config.CacheAssets) > 0 {
				if data, err := json.Marshal(config.CacheAssets); err == nil {
					assetsJSON = string(data)
				}
			}

			// Determine offline fallback
			offlineFallback := "'/'"
			if config.OfflinePage != "" {
				offlineFallback = fmt.Sprintf("'%s'", config.OfflinePage)
			}

			sw := fmt.Sprintf(`
const CACHE_NAME = '%s';
const ASSETS_TO_CACHE = %s;

// Install: pre-cache assets and skip waiting
self.addEventListener('install', event => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(cache => cache.addAll(ASSETS_TO_CACHE))
            .then(() => self.skipWaiting())
    );
});

// Activate: cleanup old caches and claim clients
self.addEventListener('activate', event => {
    event.waitUntil(
        caches.keys()
            .then(names => Promise.all(
                names.filter(n => n !== CACHE_NAME).map(n => caches.delete(n))
            ))
            .then(() => self.clients.claim())
    );
});

// Fetch: network-first for pages, cache-first for assets
self.addEventListener('fetch', event => {
    const req = event.request;

    // Skip non-GET requests
    if (req.method !== 'GET') return;

    // Skip WebSocket upgrades
    if (req.headers.get('Upgrade') === 'websocket') return;

    // Navigation (pages): network-first with offline fallback
    if (req.mode === 'navigate') {
        event.respondWith(
            fetch(req).catch(() => caches.match(%s) || caches.match('/'))
        );
        return;
    }

    // Assets: cache-first, then network
    event.respondWith(
        caches.match(req).then(cached => cached || fetch(req))
    );
});
`, app.swCacheKey, assetsJSON, offlineFallback)

			ctx.Response.Write([]byte(sw))
			return ""
		}
		app.storedMu.Lock()
		app.stored[&swHandler] = "/sw.js"
		app.storedMu.Unlock()
	}
}

// initWS registers the WebSocket endpoint for server-initiated patches.
func (app *App) initWS() {
	app.wsMu.Lock()
	if app.wsClients == nil {
		app.wsClients = make(map[*websocket.Conn]*wsState)
	}
	app.wsMu.Unlock()

	app.mux.Handle("/__ws", websocket.Handler(func(ws *websocket.Conn) {
		// Allow large payloads for file uploads (10MB)
		ws.MaxPayloadBytes = 10 * 1024 * 1024

		// Register
		st := &wsState{lastPong: time.Now()}
		// Resolve session id from handshake cookies
		if req := ws.Request(); req != nil {
			if c, err := req.Cookie("tsui__sid"); err == nil {
				st.sid = c.Value
			}
		}
		app.wsMu.Lock()
		app.wsClients[ws] = st
		app.wsMu.Unlock()
		defer func() {
			app.wsMu.Lock()
			delete(app.wsClients, ws)
			app.wsMu.Unlock()
			_ = ws.Close()
		}()

		// Heartbeat goroutine: app-level ping and stale close
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(25 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_ = websocket.Message.Send(ws, `{"type":"ping"}`)
					// touch session last-seen
					if st.sid != "" {
						app.sessMu.Lock()
						if rec := app.sessions[st.sid]; rec != nil {
							rec.lastSeen = time.Now()
						}
						app.sessMu.Unlock()
					}
					app.wsMu.RLock()
					last := st.lastPong
					app.wsMu.RUnlock()
					if time.Since(last) > 75*time.Second {
						_ = ws.Close()
						return
					}
				case <-done:
					return
				}
			}
		}()

		// Receive loop: handle ping/pong from client, invalid target notices, and callable actions
		// Message buffer for handling WebSocket frame fragmentation (large messages may be split)
		var messageBuffer strings.Builder
		for {
			var s string
			if err := websocket.Message.Receive(ws, &s); err != nil {
				close(done)
				return
			}

			// Check if this looks like the start of a new JSON message
			trimmed := strings.TrimSpace(s)
			if strings.HasPrefix(trimmed, "{") {
				// Start of new message - reset buffer
				messageBuffer.Reset()
				messageBuffer.WriteString(s)
			} else if messageBuffer.Len() > 0 {
				// Continuation of previous message (WebSocket frame fragmentation)
				messageBuffer.WriteString(s)
			}

			// Try to parse as JSON
			fullMessage := messageBuffer.String()
			if fullMessage == "" {
				fullMessage = s
			}

			var obj map[string]any
			if err := json.Unmarshal([]byte(fullMessage), &obj); err != nil {
				// If buffer is getting too large, reset it to avoid memory issues
				if messageBuffer.Len() > 20*1024*1024 {
					messageBuffer.Reset()
				}
				continue
			}

			// Successfully parsed - reset buffer
			messageBuffer.Reset()

			t, _ := obj["type"].(string)
			if t == "" {
				continue
			}
			switch t {
			case "ping":
				_ = websocket.Message.Send(ws, `{"type":"pong"}`)
			case "pong":
				app.wsMu.Lock()
				st.lastPong = time.Now()
				app.wsMu.Unlock()
				if st.sid != "" {
					app.sessMu.Lock()
					if rec := app.sessions[st.sid]; rec != nil {
						rec.lastSeen = time.Now()
					}
					app.sessMu.Unlock()
				}
			case "invalid":
				id, _ := obj["id"].(string)
				if id != "" && st.sid != "" {
					app.sessMu.Lock()
					if rec := app.sessions[st.sid]; rec != nil {
						fn := rec.targets[id]
						delete(rec.targets, id)
						app.sessMu.Unlock()
						if fn != nil {
							func() { defer func() { recover() }(); fn() }()
						}
					} else {
						app.sessMu.Unlock()
					}
				}
			case "call":
				// Handle callable action request
				msgRaw := fullMessage
				go func(raw string) {
					var callMsg JSCallMessage
					defer func() {
						if rec := recover(); rec != nil {
							log.Printf("WebSocket call handler panic: %v", rec)
							// Send error response
							if callMsg.RID != "" {
								errorEl := &JSElement{T: "span", C: []interface{}{"Error"}}
								errorResp := JSResponseMessage{
									Type: "response",
									RID:  callMsg.RID,
									El:   errorEl,
									Ops:  []*JSPatchOp{{Op: "notify", Msg: "An error occurred", Variant: "error"}},
								}
								if data, err := json.Marshal(errorResp); err == nil {
									_ = websocket.Message.Send(ws, string(data))
								}
							}
						}
					}()

					if err := json.Unmarshal([]byte(raw), &callMsg); err != nil {
						log.Printf("Failed to parse call message: %v", err)
						// Try to extract RID from raw message to send error response
						var partial struct {
							RID string `json:"rid"`
						}
						if json.Unmarshal([]byte(raw), &partial) == nil && partial.RID != "" {
							errorEl := &JSElement{T: "span", C: []interface{}{"Error parsing request"}}
							errorResp := JSResponseMessage{
								Type: "response",
								RID:  partial.RID,
								El:   errorEl,
								Ops:  []*JSPatchOp{{Op: "notify", Msg: "Failed to process request: " + err.Error(), Variant: "error"}},
							}
							if data, err := json.Marshal(errorResp); err == nil {
								_ = websocket.Message.Send(ws, string(data))
							}
						}
						return
					}

					// Look up callable from stored map
					app.storedMu.Lock()
					var found *Callable
					for storedCallable, path := range app.stored {
						if path == callMsg.Path {
							found = storedCallable
							break
						}
					}
					app.storedMu.Unlock()

					if found == nil {
						log.Printf("Callable not found for path: %s", callMsg.Path)
						errorEl := &JSElement{T: "span", C: []interface{}{"Not found"}}
						errorResp := JSResponseMessage{
							Type: "response",
							RID:  callMsg.RID,
							El:   errorEl,
							Ops:  []*JSPatchOp{{Op: "notify", Msg: "Action not found", Variant: "error"}},
						}
						if data, err := json.Marshal(errorResp); err == nil {
							_ = websocket.Message.Send(ws, string(data))
						}
						return
					}

					// Create context from WebSocket data
					ctx, err := makeContextForWS(app, st.sid, callMsg.Vals)
					if err != nil {
						log.Printf("Failed to create context: %v", err)
						errorEl := &JSElement{T: "span", C: []interface{}{"Error"}}
						errorResp := JSResponseMessage{
							Type: "response",
							RID:  callMsg.RID,
							El:   errorEl,
							Ops:  []*JSPatchOp{{Op: "notify", Msg: "Failed to process request", Variant: "error"}},
						}
						if data, err := json.Marshal(errorResp); err == nil {
							_ = websocket.Message.Send(ws, string(data))
						}
						return
					}

					// Execute callable
					html := (*found)(ctx)
					if len(ctx.append) > 0 {
						html += strings.Join(ctx.append, "")
					}

					// Convert HTML to JSON element
					jsElement, err := htmlToJSElement(html)
					if err != nil {
						log.Printf("Error converting HTML to JSON: %v", err)
						jsElement = &JSElement{T: "span", C: []interface{}{"Error"}}
					}

					// Send response
					resp := JSResponseMessage{
						Type: "response",
						RID:  callMsg.RID,
						El:   jsElement,
						Ops:  ctx.ops,
					}
					if data, err := json.Marshal(resp); err == nil {
						_ = websocket.Message.Send(ws, string(data))
					} else {
						log.Printf("Failed to marshal response: %v", err)
					}
				}(msgRaw)
			}
		}
	}))
}

// sendPatch broadcasts a patch message to all connected WS clients.
func (app *App) sendPatch(ops []*JSPatchOp) {
	msg := JSPatchMessage{
		Type: "patch",
		Ops:  ops,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	app.wsMu.RLock()
	for ws := range app.wsClients {
		go func(c *websocket.Conn) {
			defer func() { recover() }()
			if err := websocket.Message.Send(c, string(data)); err != nil {
				log.Printf("Warning: Failed to send patch to client: %v", err)
			}
		}(ws)
	}
	app.wsMu.RUnlock()
}

// Patch patches using a TargetSwap descriptor (id + swap) and pushes to WS clients.
func (ctx *Context) Patch(ts TargetSwap, html string, clear ...func()) {
	if ctx == nil || ctx.App == nil {
		return
	}

	// per-session clear callback registration
	if len(clear) > 0 && clear[0] != nil {
		ctx.App.sessMu.Lock()
		if ctx.App.sessions == nil {
			ctx.App.sessions = make(map[string]*sessRec)
		}
		rec := ctx.App.sessions[ctx.SessionID]
		if rec == nil {
			rec = &sessRec{lastSeen: time.Now(), targets: make(map[string]func())}
			ctx.App.sessions[ctx.SessionID] = rec
		}
		rec.targets[ts.ID] = clear[0]
		ctx.App.sessMu.Unlock()
	}

	// Convert HTML to JSON element and create patch operation
	jsElement, err := htmlToJSElement(html)
	if err != nil {
		log.Printf("Error converting HTML to JSON for patch: %v", err)
		// Fallback to simple text element
		jsElement = &JSElement{T: "span", C: []interface{}{"Error"}}
	}

	op := &JSPatchOp{
		Op:  string(ts.Swap),
		Tgt: ts.ID,
		El:  jsElement,
	}

	ctx.App.sendPatch([]*JSPatchOp{op})
}

// wrapJSForPatch is deprecated but kept for backwards compatibility
// Now we use JSON operations instead
func wrapJSForPatch(ts TargetSwap, jsCode string) string {
	var code string
	switch ts.Swap {
	case INLINE:
		code = fmt.Sprintf(`(function(){var t=document.getElementById('%s');if(t){t.innerHTML='';t.appendChild(%s);}})();`, escapeJS(ts.ID), jsCode)
	case OUTLINE:
		code = fmt.Sprintf(`(function(){var t=document.getElementById('%s');if(t){t.outerHTML='';var p=t.parentNode;if(p){p.replaceChild(%s,t);}}})();`, escapeJS(ts.ID), jsCode)
	case APPEND:
		code = fmt.Sprintf(`(function(){var t=document.getElementById('%s');if(t){t.appendChild(%s);}})();`, escapeJS(ts.ID), jsCode)
	case PREPEND:
		code = fmt.Sprintf(`(function(){var t=document.getElementById('%s');if(t){t.insertBefore(%s,t.firstChild);}})();`, escapeJS(ts.ID), jsCode)
	case NONE:
		code = fmt.Sprintf(`(function(){%s})();`, jsCode)
	default:
		code = jsCode
	}
	return code
}

// Render renders HTML inside the given target element (replaces innerHTML).
func (ctx *Context) Render(target Attr, html string) {
	ctx.Patch(target.Render(), html)
}

// Replace replaces the given target element with HTML (replaces outerHTML).
func (ctx *Context) Replace(target Attr, html string) {
	ctx.Patch(target.Replace(), html)
}

func (app *App) AutoRestart(enable bool) {
	if !enable {
		return
	}

	// Detect the directory of the user's main package (where main.main is defined).
	mainDir := detectMainDir()
	if mainDir == "" {
		// Fallback to current working directory
		if wd, err := os.Getwd(); err == nil {
			mainDir = wd
		} else {
			log.Println("[autorestart] cannot determine working directory:", err)
			return
		}
	}

	sanitizedRoot, err := sanitizeAutoRestartRoot(mainDir)
	if err != nil {
		log.Println("[autorestart] refusing to watch directory:", err)
		return
	}

	go watchAndRestart(sanitizedRoot)
}

// watchAndRestart watches the provided directory recursively for file changes
// and rebuilds + execs the binary in-place when a change is detected.
func watchAndRestart(root string) {
	safeRoot, err := sanitizeAutoRestartRoot(root)
	if err != nil {
		log.Println("[autorestart] invalid watch root:", err)
		return
	}
	root = safeRoot

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("[autorestart] watcher error:", err)
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Printf("[autorestart] watcher close error: %v", err)
		}
	}()

	// Add directories recursively
	addDirs := func() error {
		return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // ignore traversal errors
			}
			if d.IsDir() {
				resolvedPath := path
				if evalPath, err := filepath.EvalSymlinks(path); err == nil {
					resolvedPath = evalPath
				}
				if !isSubpath(root, resolvedPath) {
					return filepath.SkipDir
				}
				name := d.Name()
				if shouldSkipDir(name) {
					return filepath.SkipDir
				}
				if err := watcher.Add(path); err != nil {
					// Non-fatal: keep going
					return nil
				}
			}
			return nil
		})
	}

	if err := addDirs(); err != nil {
		log.Println("[autorestart] add dirs error:", err)
	}

	log.Printf("[autorestart] watching %s for changes...\n", root)

	// Debounce timer
	var (
		restartPending bool
		timer          *time.Timer
		mu             sync.Mutex
	)

	schedule := func() {
		mu.Lock()
		defer mu.Unlock()
		if restartPending {
			if timer != nil {
				timer.Reset(350 * time.Millisecond)
			}
			return
		}
		restartPending = true
		timer = time.AfterFunc(350*time.Millisecond, func() {
			// Build new binary in temp dir; then exec into it
			if err := rebuildAndExec(root); err != nil {
				log.Println("[autorestart] rebuild failed:", err)
				restartPending = false
			}
		})
	}

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Watch for new directories created (e.g., when adding packages)
			if ev.Has(fsnotify.Create) {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					resolvedName := ev.Name
					if evalName, err := filepath.EvalSymlinks(ev.Name); err == nil {
						resolvedName = evalName
					}
					if !isSubpath(root, resolvedName) {
						continue
					}
					// Add new directory and its children
					_ = filepath.WalkDir(ev.Name, func(p string, d os.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if d.IsDir() {
							resolvedPath := p
							if evalPath, err := filepath.EvalSymlinks(p); err == nil {
								resolvedPath = evalPath
							}
							if !isSubpath(root, resolvedPath) {
								return filepath.SkipDir
							}
							if shouldSkipDir(d.Name()) {
								return filepath.SkipDir
							}
							if err := watcher.Add(p); err != nil {
								log.Printf("[autorestart] failed to watch directory %s: %v", p, err)
							}
						}
						return nil
					})
					continue
				}
			}

			// Only react to relevant file changes
			if (ev.Has(fsnotify.Write) || ev.Has(fsnotify.Create) || ev.Has(fsnotify.Rename) || ev.Has(fsnotify.Remove)) && shouldTrigger(ev.Name) {
				log.Println("[autorestart] change:", ev.Name)
				schedule()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("[autorestart] watcher error:", err)
		}
	}
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules", ".idea", ".vscode", "dist", "build", "bin", ".DS_Store":
		return true
	}
	return strings.HasPrefix(name, ".")
}

func shouldTrigger(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".tmpl", ".html", ".css", ".js":
		return true
	}
	return false
}

// detectMainDir attempts to find the source directory of main.main
func detectMainDir() string {
	// First, try to inspect the stack for main.main when AutoRestart is called from user code
	for skip := 1; skip < 32; skip++ {
		pc, file, _, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn != nil && fn.Name() == "main.main" { // exact main package entrypoint
			return filepath.Dir(file)
		}
	}
	return ""
}

func sanitizeAutoRestartRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("empty autorestart root")
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	resolvedRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("autorestart root %q is not a directory", resolvedRoot)
	}

	moduleRoot, err := findModuleRoot(resolvedRoot)
	if err != nil {
		return "", err
	}

	if !isSubpath(moduleRoot, resolvedRoot) {
		return "", fmt.Errorf("autorestart root %q must stay within module root %q", resolvedRoot, moduleRoot)
	}

	return resolvedRoot, nil
}

func findModuleRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %q", start)
		}
		dir = parent
	}
}

func isSubpath(base, target string) bool {
	baseClean, err := filepath.Abs(base)
	if err != nil {
		return false
	}
	targetClean, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(baseClean, targetClean)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// rebuildAndExec builds the main package in root and re-execs into the new binary.
func rebuildAndExec(root string) error {
	safeRoot, err := sanitizeAutoRestartRoot(root)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "g-sui-build-*")
	if err != nil {
		return err
	}
	tmp := filepath.Join(tmpDir, fmt.Sprintf("g-sui-%d", time.Now().UnixNano()))
	goBin, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	cmd := exec.Command(goBin, "build", "-o", tmp, ".")
	cmd.Dir = safeRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			log.Printf("Warning: Failed to cleanup tmp dir %s: %v", tmpDir, rmErr)
		}
		return err
	}

	absTmp, err := filepath.Abs(tmp)
	if err != nil {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			log.Printf("Warning: Failed to cleanup tmp dir %s: %v", tmpDir, rmErr)
		}
		return err
	}

	// Replace current process with the new binary
	args := append([]string{absTmp}, os.Args[1:]...)
	env := os.Environ()

	// Best effort: exec on Unix, spawn+exit on Windows
	if runtime.GOOS == "windows" {
		c := exec.Command(absTmp, os.Args[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Start(); err != nil {
			if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
				log.Printf("Warning: Failed to cleanup tmp dir %s: %v", tmpDir, rmErr)
			}
			return err
		}
		// Exit current process to let the new one take over
		os.Exit(0)
		return nil
	}

	return syscall.Exec(absTmp, args, env)
}

func (app *App) Description(description string) {
	app.HTMLHead = append(app.HTMLHead, `<meta name="description" content="`+description+`">`)
}

func (app *App) HTML(title string, class string, body ...string) string {
	head := []string{
		`<title>` + title + `</title>`,
	}

	head = append(head, app.HTMLHead...)
	// Ensure Material Icons CSS is applied
	head = append(head, `<style>.material-icons{font-family:'Material Icons';font-weight:normal;font-style:normal;font-size:24px;line-height:1;letter-spacing:normal;text-transform:none;display:inline-block;white-space:nowrap;word-wrap:normal;direction:ltr;-webkit-font-feature-settings:'liga';-webkit-font-smoothing:antialiased;}</style>`)

	html := app.HTMLBody(class)
	html = strings.ReplaceAll(html, "__lang__", app.Lanugage)
	html = strings.ReplaceAll(html, "__head__", strings.Join(head, " "))

	html = strings.ReplaceAll(html, "__body__", strings.Join(body, ""))

	return Trim(html)
}

// devErrorPage returns a minimal standalone HTML page displayed on handler panics in dev.
// It tries to reconnect to the app WS at /__ws and reloads the page when the socket opens.
func (app *App) devErrorPage() string {
	wsPath := app.BasePath + "/__ws"
	return Trim(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Something went wrong…</title>
  <style>
    html,body{height:100%%}
    body{margin:0;display:flex;align-items:center;justify-content:center;background:#f3f4f6;font-family:system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#111827}
    .card{background:#fff;box-shadow:0 10px 25px rgba(0,0,0,.08);border-radius:14px;padding:28px 32px;border:1px solid rgba(0,0,0,.06);text-align:center}
    .title{font-size:20px;font-weight:600;margin-bottom:6px}
    .sub{font-size:14px;color:#6b7280}
  </style>
  </head>
  <body>
    <div class="card">
      <div class="title">Something went wrong…</div>
      <div class="sub">Waiting for server changes. Page will refresh when ready.</div>
    </div>
    <script>
      (function(){
        try {
          function connect(){
            var p=(location.protocol==='https:')?'wss://':'ws://';
            var ws=new WebSocket(p+location.host+'%s');
            ws.onopen=function(){ try{ location.reload(); } catch(_){} };
            ws.onclose=function(){ setTimeout(connect, 1000); };
            ws.onerror=function(){ try{ ws.close(); } catch(_){} };
          }
          connect();
        } catch(_){ /* noop */ }
      })();
    </script>
  </body>
</html>`, wsPath))
}

var __post = Trim(` 
    function __post(event, swap, target_id, path, values) {
		const el = event.target;
		const name = el.getAttribute("name");
		const type = el.getAttribute("type");
		const value = el.value;

		let body = values; 
		if (name != null) {
			body = body.filter(element => element.name !== name);
			body.push({ name, type, value });
		}

		var L = (function(){ try { return __loader.start(); } catch(_) { return { stop: function(){} }; } })();

		// Generate unique request ID
		var rid = 'req_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
		
		// Initialize pending requests map if needed
		if (!window.__gsuiPending) {
			window.__gsuiPending = {};
		}

		// Store pending request with callback
		window.__gsuiPending[rid] = function(data) {
			try {
				// Create element from JSON and apply to target
				var target = document.getElementById(target_id);
				if (target != null && data && data.el) {
					var element = __engine.create(data.el);
					if (element) {
						if (swap === "inline") {
							target.innerHTML = '';
							target.appendChild(element);
						} else if (swap === "outline") {
							if (target.parentNode) {
								target.parentNode.replaceChild(element, target);
							}
						} else if (swap === "append") {
							target.appendChild(element);
						} else if (swap === "prepend") {
							target.insertBefore(element, target.firstChild);
						}
					}
				}
				// Process additional operations (notifications, title, etc.)
				if (data && data.ops) {
					__engine.applyPatch(data.ops);
				}
			} catch(err) {
				try { console.error('__post callback error:', err); __error('Something went wrong ...'); } catch(__){}
			} finally {
				try { L.stop(); } catch(_){}
				delete window.__gsuiPending[rid];
			}
		};

		// Set timeout for stale requests (30 seconds)
		setTimeout(function() {
			if (window.__gsuiPending && window.__gsuiPending[rid]) {
				delete window.__gsuiPending[rid];
				try { L.stop(); } catch(_){}
				try { __error('Request timeout'); } catch(_){}
			}
		}, 30000);

		// Send WebSocket message
		try {
			var ws = (window).__gsuiWS;
			if (ws && ws.readyState === 1) {
				var msg = {
					type: 'call',
					rid: rid,
					act: 'post',
					path: path,
					swap: swap,
					tgt: target_id,
					vals: body
				};
				ws.send(JSON.stringify(msg));
			} else {
				// WebSocket not ready, fallback to error
				delete window.__gsuiPending[rid];
				try { L.stop(); } catch(_){}
				try { __error('Connection not available'); } catch(_){}
			}
		} catch(err) {
			delete window.__gsuiPending[rid];
			try { L.stop(); } catch(_){}
			try { console.error('__post error:', err); __error('Something went wrong ...'); } catch(__){}
		}
    }
`)

var __stringify = Trim(`
    function __stringify(values) {
        const result = {};

        values.forEach(item => {
            const nameParts = item.name.split('.');
            let currentObj = result;
        
            for (let i = 0; i < nameParts.length - 1; i++) {
                const part = nameParts[i];
                if (!currentObj[part]) {
                    currentObj[part] = {};
                }
                currentObj = currentObj[part];
            }
        
            const lastPart = nameParts[nameParts.length - 1];

            switch(item.type) {
                case 'date':
                case 'time':
                case 'Time':
                case 'datetime-local':
                    currentObj[lastPart] = new Date(item.value);    
                    break;
                case 'float64':
                    currentObj[lastPart] = parseFloat(item.value);
                    break;
                case 'bool':
                case 'checkbox':
                    currentObj[lastPart] = item.value === 'true';
                    break;
                default:
                    currentObj[lastPart] = item.value;
            }
        });

        return JSON.stringify(result);
    }
`)

var __submit = Trim(`
    function __submit(event, swap, target_id, path, values) {
		event.preventDefault(); 

		const el = event.target;
		const tag = el.tagName.toLowerCase();
		const form = tag === "form" ? el : el.closest("form");
		const id = form.getAttribute("id");
		let body = values; 

		// Capture clicked submit button's name/value using event.submitter (standard API)
		// Falls back to checking if event.target is the button itself
		var clickedSubmit = null;
		var submitter = event.submitter || (tag === "button" ? el : null);
		if (submitter && submitter.getAttribute("type") === "submit" && submitter.getAttribute("name")) {
			clickedSubmit = { name: submitter.getAttribute("name"), value: submitter.value || "" };
		}

		let found = Array.from(document.querySelectorAll('[form=' + id + '][name]'));

		if (found.length === 0) {
			found = Array.from(form.querySelectorAll('[name]'));
		};

		var L = (function(){ try { return __loader.start(); } catch(_) { return { stop: function(){} }; } })();

		// Process form fields and files
		var processFields = function(callback) {
			var processed = 0;
			var total = found.length;
			
			if (total === 0) {
				// Add clicked submit button value if present
				if (clickedSubmit) {
					body = body.filter(element => element.name !== clickedSubmit.name);
					body.push({ name: clickedSubmit.name, type: "submit", value: clickedSubmit.value });
				}
				callback(body);
				return;
			}

			// Helper to finalize and call callback
			var finalize = function() {
				if (clickedSubmit) {
					body = body.filter(element => element.name !== clickedSubmit.name);
					body.push({ name: clickedSubmit.name, type: "submit", value: clickedSubmit.value });
				}
				callback(body);
			};

			found.forEach((item) => {
				const name = item.getAttribute("name");
				const type = item.getAttribute("type");
				
				if (name == null) {
					processed++;
					if (processed === total) finalize();
					return;
				}

				// For radio buttons, only include the checked one
				if (type === "radio" && !item.checked) {
					processed++;
					if (processed === total) finalize();
					return;
				}

				// Skip submit buttons - only the clicked one should be included
				if (type === "submit") {
					processed++;
					if (processed === total) finalize();
					return;
				}

				if (type === "file" && item.files && item.files.length > 0) {
					// Read file as Base64
					var file = item.files[0];
					var reader = new FileReader();
					reader.onload = function(e) {
						var result = String(e.target.result || '');
						var base64 = result.split(',')[1] || result;
						body = body.filter(element => element.name !== name);
						body.push({ name: name, type: "file", value: base64, filename: file.name || name, content_type: file.type || '' });
						processed++;
						if (processed === total) finalize();
					};
					reader.onerror = function() {
						processed++;
						if (processed === total) finalize();
					};
					reader.readAsDataURL(item.files[0]);
				} else {
					let value = item.value;
					if (type === 'checkbox') {
						value = String(item.checked);
					}
					body = body.filter(element => element.name !== name);
					body.push({ name: name, type: type || 'text', value: value });
					processed++;
					if (processed === total) finalize();
				}
			});
		};

		// Generate unique request ID
		var rid = 'req_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
		
		// Initialize pending requests map if needed
		if (!window.__gsuiPending) {
			window.__gsuiPending = {};
		}

		// Store pending request with callback
		window.__gsuiPending[rid] = function(data) {
			try {
				// Create element from JSON and apply to target
				var target = document.getElementById(target_id);
				if (target != null && data && data.el) {
					var element = __engine.create(data.el);
					if (element) {
						if (swap === "inline") {
							target.innerHTML = '';
							target.appendChild(element);
						} else if (swap === "outline") {
							if (target.parentNode) {
								target.parentNode.replaceChild(element, target);
							}
						} else if (swap === "append") {
							target.appendChild(element);
						} else if (swap === "prepend") {
							target.insertBefore(element, target.firstChild);
						}
					}
				}
				// Process additional operations (notifications, title, etc.)
				if (data && data.ops) {
					__engine.applyPatch(data.ops);
				}
			} catch(err) {
				try { console.error('__submit callback error:', err); __error('Something went wrong ...'); } catch(__){}
			} finally {
				try { L.stop(); } catch(_){}
				delete window.__gsuiPending[rid];
			}
		};

		// Set timeout for stale requests (30 seconds)
		setTimeout(function() {
			if (window.__gsuiPending && window.__gsuiPending[rid]) {
				delete window.__gsuiPending[rid];
				try { L.stop(); } catch(_){}
				try { __error('Request timeout'); } catch(_){}
			}
		}, 30000);

		// Process fields and send WebSocket message
		processFields(function(finalBody) {
			try {
				var ws = (window).__gsuiWS;
				if (ws && ws.readyState === 1) {
					var msg = {
						type: 'call',
						rid: rid,
						act: 'form',
						path: path,
						swap: swap,
						tgt: target_id,
						vals: finalBody
					};
					ws.send(JSON.stringify(msg));
				} else {
					// WebSocket not ready, fallback to error
					delete window.__gsuiPending[rid];
					try { L.stop(); } catch(_){}
					try { __error('Connection not available'); } catch(_){}
				}
			} catch(err) {
				delete window.__gsuiPending[rid];
				try { L.stop(); } catch(_){}
				try { console.error('__submit error:', err); __error('Something went wrong ...'); } catch(__){}
			}
		});
    }
`)

var __load = Trim(`
    function __load(href) {
		// Prevent default navigation if event is available
		try {
			if (typeof event !== 'undefined' && event && event.preventDefault) {
				event.preventDefault();
			}
		} catch(_) {}

		// Start fetch immediately in background
		var loaderTimer = null;
		var loaderStarted = false;
		var L = null;
		
		// Set timer to show loader after 50ms if fetch is still pending
		loaderTimer = setTimeout(function() {
			if (!loaderStarted) {
				loaderStarted = true;
				try {
					L = (function(){ try { return __loader.start(); } catch(_) { return { stop: function(){} }; } })();
				} catch(_) {}
			}
		}, 50);

		fetch(href, {method: "GET"})
			.then(function(resp){ if(!resp.ok){ throw new Error('HTTP '+resp.status); } return resp.text(); })
			.then(function (html) {
				// Cancel loader timer if fetch completed quickly (< 50ms)
				if (loaderTimer) {
					clearTimeout(loaderTimer);
					loaderTimer = null;
				}
				
				// Only stop loader if it was actually started
				if (loaderStarted && L) {
					try { L.stop(); } catch(_) {}
				}

				const parser = new DOMParser();
				const doc = parser.parseFromString(html, 'text/html');

				document.title = doc.title;
				document.body.innerHTML = doc.body.innerHTML;

				const scripts = [...doc.body.querySelectorAll('script'), ...doc.head.querySelectorAll('script')];
				for (let i = 0; i < scripts.length; i++) {
					const newScript = document.createElement('script');
					newScript.textContent = scripts[i].textContent;
					document.body.appendChild(newScript);
				}

				window.history.pushState({}, doc.title, href);
			})
			.catch(function(_){ 
				// Cancel loader timer on error
				if (loaderTimer) {
					clearTimeout(loaderTimer);
					loaderTimer = null;
				}
				if (loaderStarted && L) {
					try { L.stop(); } catch(_) {}
				}
				try { __error('Something went wrong ...'); } catch(__){} 
			});
    }
`)

// __router: client-side SPA router for page navigation
var __router = Trim(`
    var __router = (function(){
        var routes = window.__routes || {};
        var contentId = '__content__';
        var initialized = false;
        
        function renderNotFound(path) {
            try {
                var target = document.getElementById(contentId);
                if (!target) return;
                target.innerHTML = '';
                var wrap = document.createElement('div');
                wrap.className = 'mx-auto max-w-3xl px-4 py-8';
                var title = document.createElement('h1');
                title.textContent = '404 - Page not found';
                title.className = 'text-2xl font-semibold mb-2';
                var text = document.createElement('p');
                text.textContent = 'No route matched ' + (path || '/');
                text.className = 'text-gray-600';
                wrap.appendChild(title);
                wrap.appendChild(text);
                target.appendChild(wrap);
                try { document.title = '404 - Page not found'; } catch(_) {}
            } catch(_) {}
        }

        function matchPattern(pathname, routeMap) {
            // Try to match against pattern routes
            for (var pattern in routeMap) {
                var routeInfo = routeMap[pattern];
                if (typeof routeInfo === 'object' && routeInfo.pattern) {
                    // Convert pattern to regex: /vehicles/edit/{id} -> /vehicles/edit/([^/]+)
                    var regexStr = pattern.replace(/\{[^}]+\}/g, '([^/]+)');
                    var regex = new RegExp('^' + regexStr + '$');
                    if (regex.test(pathname)) {
                        return routeInfo;
                    }
                }
            }
            return null;
        }

        function navigate(path, pushState) {
            // Preserve query string
            var queryIndex = path.indexOf('?');
            var pathname = queryIndex >= 0 ? path.substring(0, queryIndex) : path;
            var query = queryIndex >= 0 ? path.substring(queryIndex) : '';
            
            // Normalize pathname
            if (!pathname) pathname = '/';
            if (!pathname.startsWith('/')) pathname = '/' + pathname;

            var routeMap = window.__routes || routes || {};
            
            // Try exact match first
            var routeInfo = routeMap[pathname];
            
            // If not found, try pattern matching
            if (!routeInfo) {
                routeInfo = matchPattern(pathname, routeMap);
            }
            
            if (!routeInfo) {
                console.warn('Route not found:', pathname);
                renderNotFound(pathname);
                return;
            }
            
            var routePath = typeof routeInfo === 'string' ? routeInfo : routeInfo.path;
            
            var L = null;
            var loaderTimer = setTimeout(function() {
                try {
                    L = (function(){ try { return __loader.start(); } catch(_) { return { stop: function(){} }; } })();
                } catch(_) {}
            }, 50);
            
            // Send actual path to server for parameter extraction
            // Include query string in the path sent to server
            var pathToSend = pathname + query;
            fetch(routePath, {
                method: 'POST',
                body: JSON.stringify({path: pathToSend}),
                headers: {'Content-Type': 'application/json'}
            })
                .then(function(resp) {
                    if (resp.status === 404) {
                        renderNotFound(path);
                        return null;
                    }
                    if (!resp.ok) { throw new Error('HTTP ' + resp.status); }
                    return resp.json();
                })
                .then(function(data) {
                    if (!data) return;
                    clearTimeout(loaderTimer);
                    if (L) {
                        try { L.stop(); } catch(_) {}
                    }
                    
                    var target = document.getElementById(contentId);
                    if (target && data && data.el) {
                        var element = __engine.create(data.el);
                        if (element) {
                            target.innerHTML = '';
                            target.appendChild(element);
                        }
                    }
                    
                    // Handle operations (title, notifications, etc.)
                    if (data && data.ops) {
                        __engine.applyPatch(data.ops);
                    }
                    
                    // Update history
                    if (pushState !== false) {
                        try {
                            window.history.pushState({path: path}, '', path);
                        } catch(_) {}
                    }
                })
                .catch(function(err) {
                    clearTimeout(loaderTimer);
                    if (L) {
                        try { L.stop(); } catch(_) {}
                    }
                    try {
                        console.error('__router navigate error:', err);
                        __error('Failed to load page');
                    } catch(_) {}
                });
        }
        
        // Handle browser back/forward
        window.addEventListener('popstate', function(e) {
            navigate(window.location.pathname + window.location.search, false);
        });
        
        // Initialize on DOM ready
        function init() {
            if (initialized) return;
            initialized = true;
            
            // Navigate to current path on initial load (including query string)
            navigate(window.location.pathname + window.location.search, false);
        }
        
        // Expose navigate on window immediately for ctx.Load()
        var routerAPI = {
            navigate: navigate,
            routes: routes
        };
        
        try {
            window.__router = routerAPI;
        } catch(_) {}
        
        // Initialize navigation on DOM ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', init);
        } else {
            init();
        }
        
        return routerAPI;
    })();
`)

// __theme: initialize theme and expose setTheme(mode) on window.
// Applies html.dark class based on stored preference or system setting.
var __theme = Trim(`
    (function(){
        try {
            if (window.__gsuiThemeInit) { return; }
            window.__gsuiThemeInit = true;
            var doc = document.documentElement;
            function apply(mode){
                var m = mode;
                if (m === 'system') {
                    try { m = (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) ? 'dark' : 'light'; }
                    catch(_) { m = 'light'; }
                }
                if (m === 'dark') { try { doc.classList.add('dark'); doc.style.colorScheme = 'dark'; } catch(_){} }
                else { try { doc.classList.remove('dark'); doc.style.colorScheme = 'light'; } catch(_){} }
            }
            function set(mode){ try { localStorage.setItem('theme', mode); } catch(_){} apply(mode); }
            try { (window).setTheme = set; } catch(_){}
            try { (window).toggleTheme = function(){ var d = !!doc.classList.contains('dark'); set(d ? 'light' : 'dark'); }; } catch(_){}
            var init = 'system';
            try { init = localStorage.getItem('theme') || 'system'; } catch(_){}
            apply(init);
            try {
                if (window.matchMedia) {
                    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(){
                        var s = '';
                        try { s = localStorage.getItem('theme') || ''; } catch(_){ }
                        if (!s || s === 'system') { apply('system'); }
                    });
                }
            } catch(_){ }
        } catch(_){ }    })();
`)

// wsScript generates the WebSocket client JS with the given base path prefix.
func wsScript(basePath string) string {
	return Trim(fmt.Sprintf(`
    (function(){
        try {
            if (window.__gsuiWSInit) { return; }
            window.__gsuiWSInit = true;
            var appPing = 0;
            // Track whether the WS has ever been closed; used to trigger a full reload once it reconnects
            try { if (!(window).__gsuiHadClose) { (window).__gsuiHadClose = false; } } catch(_){ }
            function handlePatch(msg){
                try {
                    // Handle new JSON-based patches
                    if (msg.ops && msg.ops.length > 0) {
                        __engine.applyPatch(msg.ops);
                    }
                    // Fallback for old JS-based patches (backwards compatibility)
                    else if (msg.js) {
                        try { eval(String(msg.js||'')); } catch(_){ }
                    }
                } catch(_){ }
            }
            function connect(){
                var p=(location.protocol==='https:')?'wss://':'ws://';
                var ws = new WebSocket(p+location.host+'%s/__ws');
                try { (window).__gsuiWS = ws; } catch(_){ }
                ws.onopen = function(){
                    try { if (typeof __offline !== 'undefined') { __offline.hide(); } } catch(_){ }
                    try { if (appPing) { clearInterval(appPing); appPing = 0; } } catch(_){ }
                    // If the socket had previously closed, we just reconnected — reload to pick up new server state/binary
                    try { if ((window).__gsuiHadClose) { (window).__gsuiHadClose = false; try { location.reload(); return; } catch(__){} } } catch(_){ }
                    try {
                        ws.send(JSON.stringify({ type: 'ping', t: Date.now() }));
                        appPing = setInterval(function(){
                            try { ws.send(JSON.stringify({ type: 'ping', t: Date.now() })); } catch(_){ }
                        }, 30000);
                    } catch(_){ }
                };
                ws.onmessage = function(ev){
                    try {
                        var msg = JSON.parse(String(ev.data||'{}'));
                        if (!msg || !msg.type) { return; }
                        if (msg.type==='patch') { handlePatch(msg); }
                        else if (msg.type==='ping') {
                            try { ws.send(JSON.stringify({ type: 'pong', t: Date.now() })); } catch(_){ }
                        } else if (msg.type==='pong') { /* ignore */ }
                        else if (msg.type==='response') {
                            // Handle callable action response
                            try {
                                if (window.__gsuiPending && msg.rid && window.__gsuiPending[msg.rid]) {
                                    var callback = window.__gsuiPending[msg.rid];
                                    callback({ el: msg.el, ops: msg.ops });
                                }
                            } catch(err) {
                                try { console.error('Error handling response:', err); } catch(_){}
                            }
                        }
                    } catch(_){ }
                };
                ws.onclose = function(){
                    try { if (appPing) { clearInterval(appPing); appPing = 0; } } catch(_){ }
                    try { if (typeof __offline !== 'undefined') { __offline.show(); } } catch(_){ }
                    try { (window).__gsuiHadClose = true; } catch(_){ }
                    setTimeout(connect, 1500);
                };
                ws.onerror = function(){ try { ws.close(); } catch(_){ } };
            }
            connect();
        } catch(_){ }
    })();
`, basePath))
}

// __ws: default WebSocket client script (no base path prefix)
var __ws = wsScript("")

// __loader: shared loading overlay with delayed show and fade-out
var __loader = Trim(`
    var __loader = (function(){
        var S = { count: 0, t: 0, el: null };
        function build() {
            var overlay = document.createElement('div');
            overlay.className = 'fixed inset-0 z-50 flex items-center justify-center transition-opacity opacity-0';
            try { overlay.style.backdropFilter = 'blur(3px)'; } catch(_){}
            try { overlay.style.webkitBackdropFilter = 'blur(3px)'; } catch(_){}
            try { overlay.style.background = 'rgba(255,255,255,0.28)'; } catch(_){}
            try { overlay.style.pointerEvents = 'auto'; } catch(_){}
            var badge = document.createElement('div');
            badge.className = 'absolute top-3 left-3 flex items-center gap-2 rounded-full px-3 py-1 text-white shadow-lg ring-1 ring-white/30';
            badge.style.background = 'linear-gradient(135deg, #6366f1, #22d3ee)';
            var dot = document.createElement('span');
            dot.className = 'inline-block h-2.5 w-2.5 rounded-full bg-white/95 animate-pulse';
            var label = document.createElement('span');
            label.className = 'font-semibold tracking-wide';
            label.textContent = 'Loading…';
            var sub = document.createElement('span');
            sub.className = 'ml-1 text-white/85 text-xs';
            sub.textContent = 'Please wait';
            sub.style.color = 'rgba(255,255,255,0.9)';
            badge.appendChild(dot); badge.appendChild(label); badge.appendChild(sub);
            overlay.appendChild(badge);
            document.body.appendChild(overlay);
            try { requestAnimationFrame(function(){ overlay.style.opacity = '1'; }); } catch(_){}
            return overlay;
        }
        function start() {
            S.count = S.count + 1;
            if (S.el != null) { return { stop: stop }; }
            if (S.t) { return { stop: stop }; }
            S.t = setTimeout(function(){ S.t = 0; if (S.el == null) { S.el = build(); } }, 120);
            return { stop: stop };
        }
        function stop() {
            if (S.count > 0) { S.count = S.count - 1; }
            if (S.count !== 0) { return; }
            if (S.t) { try { clearTimeout(S.t); } catch(_){} S.t = 0; }
            if (S.el) {
                var el = S.el; S.el = null;
                try { el.style.opacity = '0'; } catch(_){}
                setTimeout(function(){ try { if (el && el.parentNode) { el.parentNode.removeChild(el); } } catch(_){} }, 160);
            }
        }
        return { start: start };
    })();
`)

// __offline: overlay shown when live WS disconnects; removed on reload
var __offline = Trim(`
    var __offline = (function(){
        var el = null;
        function show(){
            if (document.getElementById('__offline__')) { el = document.getElementById('__offline__'); return; }
            try { document.body.classList.add('pointer-events-none'); } catch(_){ }
            var overlay = document.createElement('div');
            overlay.id = '__offline__';
            overlay.style.position = 'fixed'; overlay.style.inset = '0'; overlay.style.zIndex = '60';
            overlay.style.pointerEvents = 'none'; overlay.style.opacity = '0'; overlay.style.transition = 'opacity 160ms ease-out';
            try { overlay.style.backdropFilter = 'blur(2px)'; } catch(_){ }
            try { overlay.style.webkitBackdropFilter = 'blur(2px)'; } catch(_){ }
            try { overlay.style.background = 'rgba(255,255,255,0.18)'; } catch(_){ }
            var badge = document.createElement('div');
            badge.className = 'absolute top-3 left-3 flex items-center gap-2 rounded-full px-3 py-1 text-white shadow-lg ring-1 ring-white/30';
            badge.style.background = 'linear-gradient(135deg, #ef4444, #ec4899)';
            var dot = document.createElement('span'); dot.className = 'inline-block h-2.5 w-2.5 rounded-full bg-white/95 animate-pulse';
            var label = document.createElement('span'); label.className = 'font-semibold tracking-wide'; label.textContent = 'Offline'; label.style.color = '#fff';
            var sub = document.createElement('span'); sub.className = 'ml-1 text-white/85 text-xs'; sub.textContent = 'Trying to reconnect…'; sub.style.color = 'rgba(255,255,255,0.9)';
            badge.appendChild(dot); badge.appendChild(label); badge.appendChild(sub);
            overlay.appendChild(badge);
            document.body.appendChild(overlay);
            try { requestAnimationFrame(function(){ overlay.style.opacity = '1'; }); } catch(_){ }
            el = overlay;
        }
        function hide(){
            try { document.body.classList.remove('pointer-events-none'); } catch(_){ }
            var o = document.getElementById('__offline__'); if (!o) { el = null; return; }
            try { o.style.opacity = '0'; } catch(_){ }
            setTimeout(function(){ try { if (o && o.parentNode) { o.parentNode.removeChild(o); } } catch(_){} }, 150);
            el = null;
        }
        return { show: show, hide: hide };
    })();
`)

// __e: DOM element creation helper for JS-based rendering
var __e = Trim(`
    function __e(tag, attrs, children) {
        var el = document.createElement(tag);
        if (attrs) {
            if (attrs.id) el.id = attrs.id;
            if (attrs.class) el.className = attrs.class;
            if (attrs.style) el.setAttribute('style', attrs.style);
            if (attrs.href) el.href = attrs.href;
            if (attrs.src) el.src = attrs.src;
            if (attrs.alt) el.alt = attrs.alt;
            if (attrs.title) el.title = attrs.title;
            if (attrs.type) el.type = attrs.type;
            if (attrs.name) el.name = attrs.name;
            if (attrs.value) el.value = attrs.value;
            if (attrs.placeholder) el.placeholder = attrs.placeholder;
            if (attrs.disabled) el.disabled = true;
            if (attrs.required) el.required = true;
            if (attrs.readonly) el.readOnly = true;
            if (attrs.checked) el.checked = true;
            if (attrs.selected) el.selected = true;
            if (attrs.min) el.min = attrs.min;
            if (attrs.max) el.max = attrs.max;
            if (attrs.step) el.step = attrs.step;
            if (attrs.rows) el.rows = attrs.rows;
            if (attrs.cols) el.cols = attrs.cols;
            if (attrs.width) el.width = attrs.width;
            if (attrs.height) el.height = attrs.height;
            if (attrs.pattern) el.pattern = attrs.pattern;
            if (attrs.autocomplete) el.autocomplete = attrs.autocomplete;
            if (attrs.for) el.htmlFor = attrs.for;
			if (attrs.form) el.setAttribute('form', attrs.form);
			if (attrs.target) el.target = attrs.target;
			if (attrs.action) el.action = attrs.action;
			if (attrs.method) el.method = attrs.method;
			if (attrs.onclick) el.setAttribute('onclick', attrs.onclick);
			if (attrs.onchange) el.setAttribute('onchange', attrs.onchange);
			if (attrs.onsubmit) el.setAttribute('onsubmit', attrs.onsubmit);
            for (var key in attrs) {
                if (key.indexOf('data-') === 0) {
                    el.setAttribute(key, attrs[key]);
                }
            }
        }
        if (children) {
            for (var i = 0; i < children.length; i++) {
                var child = children[i];
                if (child === null || child === undefined) continue;
                if (typeof child === 'string') {
                    el.appendChild(document.createTextNode(child));
                } else if (child.nodeType) {
                    el.appendChild(child);
                }
            }
        }
        return el;
    }
`)

// __engine: JSON DOM engine for declarative DOM updates
var __engine = Trim(`
    var __engine = (function(){
        var svgNS = 'http://www.w3.org/2000/svg';
        var svgTags = {svg:1,g:1,path:1,rect:1,circle:1,ellipse:1,line:1,polyline:1,polygon:1,text:1,tspan:1,defs:1,linearGradient:1,radialGradient:1,stop:1,clipPath:1,mask:1,use:1,symbol:1,marker:1,pattern:1,image:1,foreignObject:1};

        function create(json, parentSvg) {
            if (!json) return null;
            if (typeof json === 'string') {
                return document.createTextNode(json);
            }
            if (!json.t) return null;
            if (json.t === '__html') {
                var wrap = document.createElement('span');
                wrap.innerHTML = json.v || '';
                return wrap;
            }

            var isSvg = parentSvg || json.t === 'svg' || svgTags.hasOwnProperty(json.t);
            var el = isSvg ? document.createElementNS(svgNS, json.t) : document.createElement(json.t);
            
            // Apply attributes
            if (json.a) {
                for (var key in json.a) {
                    if (!json.a.hasOwnProperty(key)) continue;
                    var val = json.a[key];
                    if (!isSvg && key === 'class') {
                        el.className = val;
                    } else if (!isSvg && key === 'for') {
                        el.htmlFor = val;
                    } else if (!isSvg && key === 'readonly') {
                        el.readOnly = true;
                    } else if (!isSvg && key === 'disabled') {
                        el.disabled = true;
                    } else if (!isSvg && key === 'checked') {
                        el.checked = true;
                    } else if (!isSvg && key === 'selected') {
                        el.selected = true;
                    } else if (!isSvg && key === 'required') {
                        el.required = true;
                    } else {
                        el.setAttribute(key, val);
                    }
                }
            }
            
            // Bind events
            if (json.e) {
                bindEvents(el, json.e);
            }
            
            // Append children (propagate SVG context)
            if (json.c && json.c.length > 0) {
                for (var i = 0; i < json.c.length; i++) {
                    var child = create(json.c[i], isSvg);
                    if (child) el.appendChild(child);
                }
            }
            
            return el;
        }
        
        function bindEvents(el, events) {
            for (var eventType in events) {
                if (!events.hasOwnProperty(eventType)) continue;
                var evt = events[eventType];
                
                if (evt.act === 'post') {
                    el.addEventListener(eventType, (function(e) {
                        return function(event) {
                            __post(event, e.swap || 'inline', e.tgt || '', e.path || '', e.vals || []);
                        };
                    })(evt));
                } else if (evt.act === 'form') {
                    el.addEventListener(eventType, (function(e) {
                        return function(event) {
                            __submit(event, e.swap || 'inline', e.tgt || '', e.path || '', e.vals || []);
                        };
                    })(evt));
                } else if (evt.act === 'raw' && evt.js) {
                    el.addEventListener(eventType, (function(js) {
                        return function(event) {
                            try { (new Function('event', js))(event); } catch(_){ }
                        };
                    })(evt.js));
                }
            }
        }
        
        // Track consecutive misses per target - only send "invalid" after multiple misses
        // This handles race conditions during page load
        var missCount = {};
        
        function applyPatch(ops) {
            if (!ops || !ops.length) return;
            
            for (var i = 0; i < ops.length; i++) {
                var op = ops[i];
                
                // Handle raw JS for backwards compatibility
                if (op.js) {
                    try { eval(op.js); } catch(_){ }
                    continue;
                }
                
                // Handle notification operations
                if (op.op === 'notify') {
                    try { __notify(op.msg || '', op.variant || 'info'); } catch(_){ }
                    continue;
                }
                
                // Handle title operations
                if (op.op === 'title') {
                    try { document.title = op.title || ''; } catch(_){ }
                    continue;
                }
                
                // Handle reload operations
                if (op.op === 'reload') {
                    try { window.location.reload(); } catch(_){ }
                    continue;
                }
                
                // Handle redirect operations
                if (op.op === 'redirect') {
                    try { window.location.href = op.href || '/'; } catch(_){ }
                    continue;
                }
                
                // Handle download operations
                if (op.op === 'download') {
                    try {
                        var byteCharacters = atob(op.data || '');
                        var byteNumbers = new Array(byteCharacters.length);
                        for (var j = 0; j < byteCharacters.length; j++) {
                            byteNumbers[j] = byteCharacters.charCodeAt(j);
                        }
                        var byteArray = new Uint8Array(byteNumbers);
                        var blob = new Blob([byteArray], { type: op.content_type || 'application/octet-stream' });
                        var url = URL.createObjectURL(blob);
                        var a = document.createElement('a');
                        a.href = url;
                        a.download = op.filename || 'download';
                        a.click();
                        URL.revokeObjectURL(url);
                    } catch(_){ }
                    continue;
                }
                
                var target = document.getElementById(op.tgt);
                if (!target) {
                    if (op.tgt) {
                        // Track consecutive misses for this target
                        missCount[op.tgt] = (missCount[op.tgt] || 0) + 1;
                        
                        // Only send "invalid" after 2+ consecutive misses
                        // This handles race conditions during page load while still
                        // detecting when an element is truly gone
                        if (missCount[op.tgt] >= 2) {
                            delete missCount[op.tgt];
                            try {
                                var ws = (window).__gsuiWS;
                                if (ws && ws.readyState === 1) {
                                    ws.send(JSON.stringify({ type: 'invalid', id: op.tgt }));
                                }
                            } catch(_){ }
                        }
                    }
                    continue;
                }
                
                // Reset miss count on successful patch
                if (op.tgt) {
                    missCount[op.tgt] = 0;
                }
                
                var element = create(op.el);
                if (!element) continue;
                
                switch (op.op) {
                    case 'inline':
                        target.innerHTML = '';
                        target.appendChild(element);
                        break;
                    case 'outline':
                        if (target.parentNode) {
                            target.parentNode.replaceChild(element, target);
                        }
                        break;
                    case 'append':
                        target.appendChild(element);
                        break;
                    case 'prepend':
                        target.insertBefore(element, target.firstChild);
                        break;
                    case 'none':
                        // Just execute without modifying DOM
                        break;
                }
            }
        }
        
        return {
            create: create,
            applyPatch: applyPatch,
            bindEvents: bindEvents
        };
    })();
`)

// __notify: creates styled notification toasts
var __notify = Trim(`
    function __notify(msg, variant) {
        var box = document.getElementById('__messages__');
        if (!box) {
            box = document.createElement('div');
            box.id = '__messages__';
            box.style.position = 'fixed';
            box.style.top = '0';
            box.style.right = '0';
            box.style.padding = '8px';
            box.style.zIndex = '9999';
            box.style.pointerEvents = 'none';
            document.body.appendChild(box);
        }
        
        var n = document.createElement('div');
        n.style.display = 'flex';
        n.style.alignItems = 'center';
        n.style.gap = '10px';
        n.style.padding = '12px 16px';
        n.style.margin = '8px';
        n.style.borderRadius = '12px';
        n.style.minHeight = '44px';
        n.style.minWidth = '340px';
        n.style.maxWidth = '340px';
        n.style.boxShadow = '0 6px 18px rgba(0,0,0,0.08)';
        n.style.border = '1px solid';
        n.style.fontWeight = '600';
        
        var accent = '#4f46e5';
        var timeout = 5000;
        
        if (variant === 'success') {
            accent = '#16a34a';
            n.style.background = '#dcfce7';
            n.style.color = '#166534';
            n.style.borderColor = '#bbf7d0';
        } else if (variant === 'error' || variant === 'error-reload') {
            accent = '#dc2626';
            n.style.background = '#fee2e2';
            n.style.color = '#991b1b';
            n.style.borderColor = '#fecaca';
            if (variant === 'error-reload') {
                n.style.pointerEvents = 'auto';
                timeout = 88000;
            }
        } else {
            n.style.background = '#eef2ff';
            n.style.color = '#3730a3';
            n.style.borderColor = '#e0e7ff';
        }
        
        n.style.borderLeft = '4px solid ' + accent;
        
        var dot = document.createElement('span');
        dot.style.width = '10px';
        dot.style.height = '10px';
        dot.style.borderRadius = '9999px';
        dot.style.background = accent;
        
        var t = document.createElement('span');
        t.textContent = msg;
        
        n.appendChild(dot);
        n.appendChild(t);
        
        if (variant === 'error-reload') {
            var btn = document.createElement('button');
            btn.textContent = 'Reload';
            btn.style.background = '#991b1b';
            btn.style.color = '#fff';
            btn.style.border = 'none';
            btn.style.padding = '6px 10px';
            btn.style.borderRadius = '8px';
            btn.style.cursor = 'pointer';
            btn.style.fontWeight = '700';
            btn.onclick = function() { try { window.location.reload(); } catch(_){} };
            n.appendChild(btn);
        }
        
        box.appendChild(n);
        setTimeout(function() { try { if (n && n.parentNode) { n.parentNode.removeChild(n); } } catch(_){} }, timeout);
    }
`)

// Error UI helper injected into every page

var __error = Trim(`
    function __error(message) {
        (function(){
            try {
                var box = document.getElementById('__messages__');
                if (box == null) { box = document.createElement('div'); box.id='__messages__'; box.style.position='fixed'; box.style.top='0'; box.style.right='0'; box.style.padding='8px'; box.style.zIndex='9999'; box.style.pointerEvents='none'; document.body.appendChild(box); }
                var n = document.getElementById('__error_toast__');
                if (!n) {
                    n = document.createElement('div'); n.id='__error_toast__';
                    n.style.display='flex'; n.style.alignItems='center'; n.style.gap='10px'; n.style.padding='12px 16px'; n.style.margin='8px'; n.style.borderRadius='12px'; n.style.minHeight='44px'; n.style.minWidth='340px'; n.style.maxWidth='340px';
                    n.style.background='#fee2e2'; n.style.color='#991b1b'; n.style.border='1px solid #fecaca'; n.style.borderLeft='4px solid #dc2626'; n.style.boxShadow='0 6px 18px rgba(0,0,0,0.08)'; n.style.fontWeight='600'; n.style.pointerEvents='auto';
                    var dot=document.createElement('span'); dot.style.width='10px'; dot.style.height='10px'; dot.style.borderRadius='9999px'; dot.style.background='#dc2626'; n.appendChild(dot);
                    var span=document.createElement('span'); span.id='__error_text__'; n.appendChild(span);
                    var btn=document.createElement('button'); btn.textContent='Reload'; btn.style.background='#991b1b'; btn.style.color='#fff'; btn.style.border='none'; btn.style.padding='6px 10px'; btn.style.borderRadius='8px'; btn.style.cursor='pointer'; btn.style.fontWeight='700'; btn.onclick=function(){ try { window.location.reload(); } catch(_){} }; n.appendChild(btn);
                    box.appendChild(n);
                }
                var spanText = document.getElementById('__error_text__'); if (spanText) { spanText.textContent = message || 'Something went wrong ...'; }
            } catch(_) { try { alert(message || 'Something went wrong ...'); } catch(__){} }
        })();
    }
`)

var __cfmt = Trim(`
	var __cfmt = (function(){
		var loc = (window.__locale || "sk").toLowerCase();

		function date(val) {
			if (!val) return "";
			var d;
			if (val instanceof Date) { d = val; }
			else { d = new Date(val); }
			if (isNaN(d.getTime())) return String(val);
			var dd = ("0"+d.getDate()).slice(-2);
			var mm = ("0"+(d.getMonth()+1)).slice(-2);
			return dd+"."+mm+"."+d.getFullYear();
		}

		function amount(val) {
			if (val === null || val === undefined || val === "") return "";
			var n = parseFloat(val);
			if (isNaN(n)) return String(val);
			var parts = n.toFixed(2).split(".");
			parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, " ");
			return parts.join(",");
		}

		function number(val) {
			if (val === null || val === undefined || val === "") return "";
			var n = parseFloat(val);
			if (isNaN(n)) return String(val);
			if (n === Math.floor(n)) {
				return String(n).replace(/\B(?=(\d{3})+(?!\d))/g, " ");
			}
			var parts = n.toString().split(".");
			parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, " ");
			return parts.join(",");
		}

		function bool(val) {
			return val ? "check_circle" : "cancel";
		}

		function truncate(val, max) {
			if (!val) return "";
			var s = String(val);
			if (!max || s.length <= max) return s;
			return s.substring(0, max) + "...";
		}

		function escape(val) {
			if (!val) return "";
			var s = String(val);
			return s.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;");
		}

		function relativeTime(val) {
			if (!val) return "";
			var d;
			if (val instanceof Date) { d = val; }
			else { d = new Date(val); }
			if (isNaN(d.getTime())) return String(val);
			var now = new Date();
			var diff = Math.floor((now.getTime() - d.getTime()) / 1000);
			if (diff < 0) {
				var absDiff = Math.abs(diff);
				if (absDiff < 60) return "in " + absDiff + "s";
				if (absDiff < 3600) return "in " + Math.floor(absDiff / 60) + " min";
				if (absDiff < 86400) return "in " + Math.floor(absDiff / 3600) + "h";
				return "in " + Math.floor(absDiff / 86400) + "d";
			}
			if (diff < 10) return "just now";
			if (diff < 60) return diff + "s ago";
			if (diff < 3600) return Math.floor(diff / 60) + " min ago";
			if (diff < 86400) return Math.floor(diff / 3600) + " hours ago";
			if (diff < 604800) return Math.floor(diff / 86400) + " days ago";
			return date(val);
		}

		function datePreset(name) {
			var now = new Date();
			var y = now.getFullYear();
			var m = now.getMonth();
			var d = now.getDate();
			var from, to;
			function fmt(dt) { return dt.toISOString().slice(0, 10); }
			to = fmt(now);
			if (name === "today") {
				from = to;
			} else if (name === "thisWeek") {
				var dow = now.getDay() || 7;
				from = fmt(new Date(y, m, d - dow + 1));
			} else if (name === "thisMonth") {
				from = fmt(new Date(y, m, 1));
			} else if (name === "thisQuarter") {
				var qm = m - (m % 3);
				from = fmt(new Date(y, qm, 1));
			} else if (name === "thisYear") {
				from = fmt(new Date(y, 0, 1));
			} else if (name === "lastMonth") {
				from = fmt(new Date(y, m - 1, 1));
				to = fmt(new Date(y, m, 0));
			} else if (name === "lastYear") {
				from = fmt(new Date(y - 1, 0, 1));
				to = fmt(new Date(y - 1, 11, 31));
			} else {
				from = to;
			}
			return { from: from, to: to };
		}

		return { date: date, amount: amount, number: number, bool: bool, truncate: truncate, escape: escape, relativeTime: relativeTime, datePreset: datePreset };
	})();
`)

var __debounce = Trim(`
	function __debounce(fn, delay) {
		var timer = null;
		delay = delay || 250;
		function debounced() {
			var ctx = this, args = arguments;
			if (timer) clearTimeout(timer);
			timer = setTimeout(function() {
				timer = null;
				fn.apply(ctx, args);
			}, delay);
		}
		debounced.cancel = function() {
			if (timer) { clearTimeout(timer); timer = null; }
		};
		return debounced;
	}
`)

var __clipboard = Trim(`
	function __clipboard(text) {
		function fallback(t) {
			var ta = document.createElement("textarea");
			ta.value = t;
			ta.style.position = "fixed";
			ta.style.left = "-9999px";
			document.body.appendChild(ta);
			ta.select();
			try { document.execCommand("copy"); } catch(e) { document.body.removeChild(ta); return Promise.reject(e); }
			document.body.removeChild(ta);
			return Promise.resolve();
		}
		var p = navigator.clipboard ? navigator.clipboard.writeText(text) : fallback(text);
		return p.then(function() {
			__notify("Copied to clipboard", "success");
		});
	}
`)

var __capi = Trim(`
	var __capi = (function(){
		var pending = {};
		var baseUrl = "";

		function setBase(url) { baseUrl = url ? url.replace(/\/+$/, "") : ""; }

		function resolveUrl(url) {
			if (!baseUrl || url.indexOf("http") === 0) return url;
			return baseUrl + (url.charAt(0) === "/" ? "" : "/") + url;
		}

		function creds() {
			return baseUrl ? "include" : "same-origin";
		}

		function fetchOpts() {
			var o = { credentials: creds() };
			if (baseUrl) o.mode = "cors";
			return o;
		}

		function buildUrl(url, params) {
			if (!params) return url;
			var qs = [];
			for (var k in params) {
				if (!params.hasOwnProperty(k)) continue;
				qs.push(encodeURIComponent(k) + "=" + encodeURIComponent(params[k]));
			}
			if (qs.length === 0) return url;
			return url + (url.indexOf("?") >= 0 ? "&" : "?") + qs.join("&");
		}

		function get(url, params) {
			var fullUrl = buildUrl(resolveUrl(url), params);
			if (pending[fullUrl]) return pending[fullUrl];
			var opts = fetchOpts();
			opts.method = "GET";
			opts.headers = { "Accept": "application/json" };
			var p = fetch(fullUrl, opts).then(function(resp) {
				delete pending[fullUrl];
				if (!resp.ok) throw new Error("HTTP " + resp.status);
				return resp.json();
			}).catch(function(err) {
				delete pending[fullUrl];
				throw err;
			});
			pending[fullUrl] = p;
			return p;
		}

		function post(url, body) {
			var opts = fetchOpts();
			opts.method = "POST";
			opts.headers = { "Content-Type": "application/json", "Accept": "application/json" };
			opts.body = JSON.stringify(body);
			return fetch(resolveUrl(url), opts).then(function(resp) {
				if (!resp.ok) throw new Error("HTTP " + resp.status);
				return resp.json();
			});
		}

		function upload(url, formData, onProgress) {
			var resolved = resolveUrl(url);
			if (onProgress) {
				return new Promise(function(resolve, reject) {
					var xhr = new XMLHttpRequest();
					xhr.open("POST", resolved, true);
					xhr.withCredentials = !!baseUrl;
					xhr.upload.addEventListener("progress", function(e) {
						if (e.lengthComputable) onProgress(e.loaded / e.total);
					});
					xhr.onload = function() {
						if (xhr.status >= 200 && xhr.status < 300) {
							try { resolve(JSON.parse(xhr.responseText)); } catch(e) { reject(e); }
						} else {
							reject(new Error("HTTP " + xhr.status));
						}
					};
					xhr.onerror = function() { reject(new Error("Upload failed")); };
					xhr.send(formData);
				});
			}
			var opts = fetchOpts();
			opts.method = "POST";
			opts.body = formData;
			return fetch(resolved, opts).then(function(resp) {
				if (!resp.ok) throw new Error("HTTP " + resp.status);
				return resp.json();
			});
		}

		function download(url, filename) {
			var opts = fetchOpts();
			opts.method = "GET";
			return fetch(resolveUrl(url), opts).then(function(resp) {
				if (!resp.ok) throw new Error("HTTP " + resp.status);
				return resp.blob();
			}).then(function(blob) {
				var objUrl = URL.createObjectURL(blob);
				var a = document.createElement("a");
				a.href = objUrl;
				a.download = filename || "download";
				a.style.display = "none";
				document.body.appendChild(a);
				a.click();
				setTimeout(function() { URL.revokeObjectURL(objUrl); document.body.removeChild(a); }, 100);
			});
		}

		function healthCheck(url, opts) {
			opts = opts || {};
			var timeout = opts.timeout || 5000;
			var retries = opts.retries || 3;
			var interval = opts.interval || 2000;
			var onChange = opts.onChange;
			var lastState = null;

			function attempt(remaining) {
				return new Promise(function(resolve) {
					var ctrl = typeof AbortController !== "undefined" ? new AbortController() : null;
					var timer = setTimeout(function() {
						if (ctrl) ctrl.abort();
						resolve(false);
					}, timeout);
					var fOpts = { method: "GET", credentials: creds() };
					if (ctrl) fOpts.signal = ctrl.signal;
					fetch(resolveUrl(url), fOpts).then(function(resp) {
						clearTimeout(timer);
						resolve(resp.ok);
					}).catch(function() {
						clearTimeout(timer);
						resolve(false);
					});
				}).then(function(ok) {
					if (ok) {
						if (onChange && lastState !== true) { lastState = true; onChange(true); }
						return true;
					}
					if (remaining > 1) {
						return new Promise(function(r) { setTimeout(r, interval); }).then(function() {
							return attempt(remaining - 1);
						});
					}
					if (onChange && lastState !== false) { lastState = false; onChange(false); }
					return false;
				});
			}

			return attempt(retries);
		}

		return { get: get, post: post, upload: upload, download: download, healthCheck: healthCheck, buildUrl: buildUrl, setBase: setBase };
	})();
`)

var __cel = Trim(`
	function __cel(tag, attrs, children, events) {
		var el = { t: tag };
		if (attrs) {
			var a = {};
			for (var k in attrs) {
				if (attrs.hasOwnProperty(k) && attrs[k] !== null && attrs[k] !== undefined) {
					a[k] = String(attrs[k]);
				}
			}
			if (Object.keys(a).length > 0) el.a = a;
		}
		if (children) {
			var c = [];
			for (var i = 0; i < children.length; i++) {
				if (children[i] !== null && children[i] !== undefined && children[i] !== false) {
					c.push(children[i]);
				}
			}
			if (c.length > 0) el.c = c;
		}
		if (events) {
			el.e = events;
		}
		return el;
	}
	__cel.div = function(cls, children) {
		return __cel("div", cls ? { class: cls } : null, Array.isArray(children) ? children : null);
	};
	__cel.span = function(cls, children) {
		return __cel("span", cls ? { class: cls } : null, Array.isArray(children) ? children : null);
	};
	__cel.text = function(str) {
		return (str === null || str === undefined) ? "" : String(str);
	};
	__cel.on = function(event, jsFn) {
		var e = {};
		e[event] = { act: "raw", js: jsFn };
		return e;
	};
	__cel.if = function(cond, thenFn, elseFn) {
		return cond ? thenFn() : (elseFn ? elseFn() : null);
	};
	__cel.map = function(arr, fn) {
		if (!arr) return [];
		var result = [];
		for (var i = 0; i < arr.length; i++) {
			var item = fn(arr[i], i);
			if (item !== null && item !== undefined) result.push(item);
		}
		return result;
	};
	__cel.html = function(str) {
		return { t: "__html", v: str || "" };
	};
	__cel.icon = function(name, cls) {
		return __cel("span", { class: "material-icons" + (cls ? " " + cls : "") }, [name]);
	};
`)

var __cregister = Trim(`
	window.__cregistry = {};
	function __cregister(name, fn) {
		if (typeof fn !== "function") throw new Error("__cregister: fn must be a function for '" + name + "'");
		window.__cregistry[name] = fn;
	}
	function __cget(name) {
		var fn = window.__cregistry[name];
		if (!fn) throw new Error("__cget: component '" + name + "' not registered");
		return fn;
	}
`)

var __clientScript = Trim(`
	window.__clients = {};
	function __client(config) {
		var el = document.getElementById(config.id);
		if (!el) return;

		var data = null;
		var state = {
			zoneId: config.id,
			page: 0,
			sort: { col: "", dir: "" },
			filters: {},
			search: ""
		};
		var pollTimer = null;
		var destroyed = false;

		// C1: Restore state from URL params or localStorage
		(function restoreState() {
			var params = new URLSearchParams(window.location.search);
			var hasUrlState = false;
			if (params.has("search")) { state.search = params.get("search"); hasUrlState = true; }
			if (params.has("sort")) {
				var sp = params.get("sort").split(":");
				state.sort = { col: sp[0] || "", dir: sp[1] || "asc" };
				hasUrlState = true;
			}
			if (params.has("page")) { state.page = parseInt(params.get("page"), 10) || 0; hasUrlState = true; }
			params.forEach(function(val, key) {
				if (key.indexOf("f.") === 0) {
					var col = key.substring(2);
					var ci = val.indexOf(":");
					if (ci > 0) {
						state.filters[col] = { op: val.substring(0, ci), value: val.substring(ci + 1) };
					}
					hasUrlState = true;
				}
			});
			if (!hasUrlState) {
				try {
					var saved = localStorage.getItem("__gsui_filter_" + window.location.pathname);
					if (saved) {
						var parsed = JSON.parse(saved);
						if (parsed.search) state.search = parsed.search;
						if (parsed.sort) state.sort = parsed.sort;
						if (parsed.page) state.page = parsed.page;
						if (parsed.filters) state.filters = parsed.filters;
					}
				} catch(e) {}
			}
		})();

		function syncStateToUrl() {
			try {
				var params = new URLSearchParams();
				if (state.search) params.set("search", state.search);
				if (state.sort && state.sort.col) params.set("sort", state.sort.col + ":" + (state.sort.dir || "asc"));
				if (state.page > 0) params.set("page", String(state.page));
				if (state.filters) {
					for (var col in state.filters) {
						if (!state.filters.hasOwnProperty(col)) continue;
						var f = state.filters[col];
						params.set("f." + col, f.op + ":" + (f.value || ""));
					}
				}
				var qs = params.toString();
				var newUrl = window.location.pathname + (qs ? "?" + qs : "");
				history.replaceState(null, "", newUrl);
				localStorage.setItem("__gsui_filter_" + window.location.pathname, JSON.stringify({
					search: state.search, sort: state.sort, page: state.page, filters: state.filters
				}));
			} catch(e) {}
		}

		function showLoading() {
			el.innerHTML = "";
			var type = config.loading || "component";
			var skeleton = __cel.div("animate-pulse", [
				__cel.div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow", [
					__cel.div("bg-gray-200 dark:bg-gray-700 h-5 rounded w-5/6 mb-2", []),
					__cel.div("bg-gray-200 dark:bg-gray-700 h-5 rounded w-2/3 mb-2", []),
					__cel.div("bg-gray-200 dark:bg-gray-700 h-5 rounded w-4/6", [])
				])
			]);
			if (type === "table") {
				var headerCells = [];
				var rows = [];
				for (var h = 0; h < 4; h++) {
					headerCells.push(__cel("th", { class: "p-3" }, [
						__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-20", [])
					]));
				}
				for (var r = 0; r < 5; r++) {
					var cells = [];
					for (var c = 0; c < 4; c++) {
						var w = c === 0 ? "w-32" : c === 2 ? "w-16" : "w-24";
						cells.push(__cel("td", { class: "p-3" }, [
							__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded " + w, [])
						]));
					}
					rows.push(__cel("tr", { class: "border-t border-gray-100 dark:border-gray-800" }, cells));
				}
				skeleton = __cel.div("animate-pulse", [
					__cel.div("bg-white dark:bg-gray-900 rounded-lg shadow overflow-hidden", [
						__cel("table", { class: "w-full" }, [
							__cel("thead", null, [
								__cel("tr", { class: "border-b border-gray-200 dark:border-gray-700" }, headerCells)
							]),
							__cel("tbody", null, rows)
						])
					])
				]);
			} else if (type === "cards") {
				var cards = [];
				for (var ci = 0; ci < 6; ci++) {
					cards.push(__cel.div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow", [
						__cel.div("bg-gray-200 dark:bg-gray-700 h-5 rounded w-3/4 mb-3", []),
						__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-1/2 mb-2", []),
						__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-2/3", [])
					]));
				}
				skeleton = __cel.div("animate-pulse", [
					__cel.div("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4", cards)
				]);
			}
			var node = __engine.create(skeleton);
			if (node) el.appendChild(node);
		}

		function showError(msg) {
			if (!config.error && config.error !== undefined) return;
			el.innerHTML = "";
			var errorEl = __cel.div("bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6 text-center", [
				__cel.icon("error_outline", "text-red-400 text-3xl mb-2"),
				__cel.div("text-red-600 dark:text-red-400 font-medium mb-1", [__cel.text("Failed to load data")]),
				__cel.div("text-red-500 dark:text-red-500 text-sm", [__cel.text(msg || "Unknown error")])
			]);
			var node = __engine.create(errorEl);
			if (node) el.appendChild(node);
		}

		function showEmpty() {
			el.innerHTML = "";
			var icon = (config.empty && config.empty.icon) || "inbox";
			var message = (config.empty && config.empty.message) || "No data";
			var emptyEl = __cel.div("bg-white dark:bg-gray-900 rounded-lg p-12 text-center shadow", [
				__cel.icon(icon, "text-gray-300 dark:text-gray-600 text-5xl mb-3"),
				__cel.div("text-gray-400 dark:text-gray-500 text-lg", [__cel.text(message)])
			]);
			var node = __engine.create(emptyEl);
			if (node) el.appendChild(node);
		}

		function isEmpty(d) {
			if (d === null || d === undefined) return true;
			if (Array.isArray(d) && d.length === 0) return true;
			if (typeof d === "object" && !Array.isArray(d) && Object.keys(d).length === 0) return true;
			return false;
		}

		function render() {
			if (destroyed) return;
			if (isEmpty(data) && !config.component) { showEmpty(); return; }
			try {
				var fn = __cget(config.component);
				var jsEl = fn(data, state, config.opts || {});
				if (!jsEl) { el.innerHTML = ""; return; }
				el.innerHTML = "";
				var node = __engine.create(jsEl);
				if (node) el.appendChild(node);
				requestAnimationFrame(function(){
					var dd = el.querySelector("[data-filter-position]");
					if (!dd) return;
					var key = dd.getAttribute("data-filter-position");
					var th = el.querySelector("th[data-col-key='" + key + "']");
					if (th) {
						var r = th.getBoundingClientRect();
						var left = r.left;
						var ddW = dd.offsetWidth || 220;
						if (left + ddW > window.innerWidth - 8) left = window.innerWidth - ddW - 8;
						if (left < 8) left = 8;
						dd.style.top = (r.bottom + 4) + "px";
						dd.style.left = left + "px";
					}
				});
			} catch (err) {
				showError(err.message || "Render error");
			}
		}

		function load() {
			if (destroyed) return;
			var params = {};
			if (config.params) {
				for (var k in config.params) {
					if (config.params.hasOwnProperty(k)) params[k] = config.params[k];
				}
			}
			if (state.search) params._search = state.search;
			if (state.sort && state.sort.col) {
				params._sort = state.sort.col;
				params._dir = state.sort.dir || "asc";
			}
			if (state.page > 0) params._page = String(state.page);

			__capi.get(config.source, Object.keys(params).length > 0 ? params : null)
				.then(function(d) {
					if (destroyed) return;
					data = d;
					render();
				})
				.catch(function(err) {
					if (destroyed) return;
					showError(err.message || "Failed to load");
				});
		}

		var instance = {
			reload: function() { load(); },
			refetch: function() { load(); },
			destroy: function() {
				destroyed = true;
				if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
				el.innerHTML = "";
			},
			setState: function(partial) {
				for (var k in partial) {
					if (partial.hasOwnProperty(k)) state[k] = partial[k];
				}
				syncStateToUrl();
				render();
			},
			getState: function() { return state; },
			getData: function() { return data; }
		};

		window.__clients[config.id] = instance;

		if (config.autoLoad !== false) {
			showLoading();
			load();
		} else if (config.component) {
			render();
		}

		if (config.poll && config.poll > 0) {
			pollTimer = setInterval(function() {
				if (destroyed) { clearInterval(pollTimer); pollTimer = null; return; }
				if (config.pollWhile) {
					try {
						var shouldPoll = (new Function("data", "state", "return " + config.pollWhile))(data, state);
						if (!shouldPoll) { clearInterval(pollTimer); pollTimer = null; return; }
					} catch(e) { clearInterval(pollTimer); pollTimer = null; return; }
				}
				load();
			}, config.poll);
		}
	}
`)

var __caction = Trim(`
	function __caction(path, payload) {
		var fakeEvent = { preventDefault: function(){}, target: document.body };
		var vals = [];
		if (payload) {
			for (var k in payload) {
				if (payload.hasOwnProperty(k)) {
					vals.push({ name: k, value: String(payload[k]), type: "text" });
				}
			}
		}
		__post(fakeEvent, "none", "", path, vals);
	}
`)

var __ctable = Trim(`
	(function(){
		function renderFilterDropdown(col, state, zoneId, data) {
			var f = (state.filters && state.filters[col.key]) || {};
			var children = [];
			var colType = col.type || "text";

			if (colType === "text" || colType === "custom") {
				var ops = ["contains", "startsWith", "equals"];
				var curOp = f.op || "contains";
				children.push(__cel("select", {
					class: "w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 mb-2",
					"data-role": "filter-op"
				}, ops.map(function(op) {
					return __cel("option", { value: op, selected: op === curOp ? "selected" : null }, [__cel.text(op)]);
				})));
				children.push(__cel("input", {
					type: "text",
					class: "w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800",
					placeholder: "Filter value...",
					value: f.value || "",
					"data-role": "filter-val"
				}, null));
			} else if (colType === "number") {
				var numOps = ["eq", "gt", "lt", "gte", "lte", "between"];
				var curNumOp = f.op || "eq";
				children.push(__cel("select", {
					class: "w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 mb-2",
					"data-role": "filter-op"
				}, numOps.map(function(op) {
					var labels = { eq: "=", gt: ">", lt: "<", gte: ">=", lte: "<=", between: "between" };
					return __cel("option", { value: op, selected: op === curNumOp ? "selected" : null }, [__cel.text(labels[op] || op)]);
				})));
				children.push(__cel("input", {
					type: "number",
					class: "w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 mb-1",
					placeholder: curNumOp === "between" ? "From..." : "Value...",
					value: f.value || f.from || "",
					"data-role": "filter-val"
				}, null));
				if (curNumOp === "between") {
					children.push(__cel("input", {
						type: "number",
						class: "w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800",
						placeholder: "To...",
						value: f.to || "",
						"data-role": "filter-to"
					}, null));
				}
			} else if (colType === "date") {
				children.push(__cel.div("flex items-center gap-2 mb-2", [
					__cel("span", { class: "text-xs text-gray-500 dark:text-gray-400 w-8 shrink-0" }, [__cel.text("From")]),
					__cel("input", {
						type: "date",
						class: "flex-1 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800",
						value: f.from || "",
						"data-role": "filter-from"
					}, null)
				]));
				children.push(__cel.div("flex items-center gap-2 mb-3", [
					__cel("span", { class: "text-xs text-gray-500 dark:text-gray-400 w-8 shrink-0" }, [__cel.text("To")]),
					__cel("input", {
						type: "date",
						class: "flex-1 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800",
						value: f.to || "",
						"data-role": "filter-to"
					}, null)
				]));
				// Date presets
				var presets = [
					{ key: "today", label: "Today" },
					{ key: "thisWeek", label: "This week" },
					{ key: "thisMonth", label: "This month" },
					{ key: "thisQuarter", label: "This quarter" },
					{ key: "thisYear", label: "This year" },
					{ key: "lastMonth", label: "Last month" },
					{ key: "lastYear", label: "Last year" }
				];
				var presetRow = presets.map(function(p) {
					return __cel("button", {
						type: "button",
						class: "text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300",
						"data-preset": p.key
					}, [__cel.text(p.label)], {
						click: { act: "raw", js: "(function(){ var pr='" + p.key + "'; var now=new Date(); var from,to; if(pr==='today'){from=to=now.toISOString().slice(0,10);} else if(pr==='thisWeek'){var d=now.getDay()||7; from=new Date(now.getFullYear(),now.getMonth(),now.getDate()-d+1).toISOString().slice(0,10); to=now.toISOString().slice(0,10);} else if(pr==='thisMonth'){from=new Date(now.getFullYear(),now.getMonth(),1).toISOString().slice(0,10); to=now.toISOString().slice(0,10);} else if(pr==='thisQuarter'){var qm=Math.floor(now.getMonth()/3)*3; from=new Date(now.getFullYear(),qm,1).toISOString().slice(0,10); to=now.toISOString().slice(0,10);} else if(pr==='thisYear'){from=new Date(now.getFullYear(),0,1).toISOString().slice(0,10); to=now.toISOString().slice(0,10);} else if(pr==='lastMonth'){from=new Date(now.getFullYear(),now.getMonth()-1,1).toISOString().slice(0,10); to=new Date(now.getFullYear(),now.getMonth(),0).toISOString().slice(0,10);} else if(pr==='lastYear'){from=new Date(now.getFullYear()-1,0,1).toISOString().slice(0,10); to=new Date(now.getFullYear()-1,11,31).toISOString().slice(0,10);} var dd=event.target.closest('[data-filter-dropdown]'); var fromEl=dd.querySelector('[data-role=filter-from]'); var toEl=dd.querySelector('[data-role=filter-to]'); if(fromEl)fromEl.value=from; if(toEl)toEl.value=to; })()" }
					});
				});
				children.push(__cel.div("flex flex-wrap gap-1 mb-2", presetRow));
			} else if (colType === "enum") {
				var enumOpts = col.enumOptions || [];
				var selected = (f.op === "in" && f.values) ? f.values : [];
				children.push(__cel.div("flex gap-2 mb-2", [
					__cel("button", { type: "button", class: "text-xs text-blue-600 dark:text-blue-400 underline" }, [__cel.text("All")], {
						click: { act: "raw", js: "(function(){ var el=event.target.closest('[data-filter-dropdown]'); var cbs=el.querySelectorAll('input[type=checkbox]'); for(var i=0;i<cbs.length;i++) cbs[i].checked=true; })()" }
					}),
					__cel("button", { type: "button", class: "text-xs text-blue-600 dark:text-blue-400 underline" }, [__cel.text("None")], {
						click: { act: "raw", js: "(function(){ var el=event.target.closest('[data-filter-dropdown]'); var cbs=el.querySelectorAll('input[type=checkbox]'); for(var i=0;i<cbs.length;i++) cbs[i].checked=false; })()" }
					})
				]));
				children.push(__cel.div("max-h-40 overflow-y-auto", enumOpts.map(function(opt) {
					var checked = selected.indexOf(opt.value) >= 0;
					return __cel("label", { class: "flex items-center gap-2 py-1 text-sm cursor-pointer" }, [
						__cel("input", { type: "checkbox", value: opt.value, checked: checked ? "checked" : null, "data-role": "filter-enum" }, null),
						__cel.text(opt.label || opt.value)
					]);
				})));
			} else if (colType === "bool") {
				var curBool = f.value;
				var boolOpts = [{ label: "Any", value: "" }, { label: "Yes", value: "true" }, { label: "No", value: "false" }];
				boolOpts.forEach(function(bo) {
					children.push(__cel("label", { class: "flex items-center gap-2 py-1 text-sm cursor-pointer" }, [
						__cel("input", { type: "radio", name: "filter_bool_" + col.key, value: bo.value, checked: String(curBool) === bo.value ? "checked" : null, "data-role": "filter-bool" }, null),
						__cel.text(bo.label)
					]));
				});
			}

			// Apply + Clear buttons
			var btnRow = __cel.div("flex gap-2 mt-3", [
				__cel("button", {
					type: "button",
					class: "flex-1 bg-blue-600 text-white text-sm px-3 py-1.5 rounded hover:bg-blue-700"
				}, [__cel.text("Apply")], {
					click: { act: "raw", js: "(function(){ var dd=event.target.closest('[data-filter-dropdown]'); var s=__clients['" + zoneId + "'].getState(); var ff=Object.assign({},s.filters); var colType='" + colType + "'; var colKey='" + col.key + "'; if(colType==='enum'){var cbs=dd.querySelectorAll('[data-role=filter-enum]'); var vals=[]; for(var i=0;i<cbs.length;i++){if(cbs[i].checked) vals.push(cbs[i].value);} if(vals.length>0){ff[colKey]={op:'in',values:vals};} else {delete ff[colKey];}} else if(colType==='bool'){var rb=dd.querySelector('[data-role=filter-bool]:checked'); if(rb&&rb.value){ff[colKey]={op:'bool',value:rb.value==='true'};} else {delete ff[colKey];}} else if(colType==='date'){var from=dd.querySelector('[data-role=filter-from]'); var to=dd.querySelector('[data-role=filter-to]'); if((from&&from.value)||(to&&to.value)){ff[colKey]={op:'dateRange',from:from?from.value:'',to:to?to.value:''};} else {delete ff[colKey];}} else if(colType==='number'){var op=dd.querySelector('[data-role=filter-op]'); var val=dd.querySelector('[data-role=filter-val]'); var toEl=dd.querySelector('[data-role=filter-to]'); var opVal=op?op.value:'eq'; if(opVal==='between'&&toEl){ff[colKey]={op:'between',from:val?val.value:'',to:toEl.value};} else if(val&&val.value){ff[colKey]={op:opVal,value:val.value};} else {delete ff[colKey];}} else {var op=dd.querySelector('[data-role=filter-op]'); var val=dd.querySelector('[data-role=filter-val]'); if(val&&val.value){ff[colKey]={op:op?op.value:'contains',value:val.value};} else {delete ff[colKey];}} __clients['" + zoneId + "'].setState({filters:ff,page:0,_filterOpen:null}); })()" }
				}),
				__cel("button", {
					type: "button",
					class: "flex-1 border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 text-sm px-3 py-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
				}, [__cel.text("Clear")], {
					click: { act: "raw", js: "(function(){ var s=__clients['" + zoneId + "'].getState(); var ff=Object.assign({},s.filters); delete ff['" + col.key + "']; __clients['" + zoneId + "'].setState({filters:ff,page:0,_filterOpen:null}); })()" }
				})
			]);
			children.push(btnRow);

			var ddWidth = "w-[220px]";
			if (colType === "date") ddWidth = "w-[260px]";
			return __cel("div", {
				class: "fixed bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg p-3 " + ddWidth + " z-[9999]",
				"data-filter-dropdown": col.key,
				"data-filter-position": col.key
			}, children);
		}

		function tableComponent(data, state, opts) {
			var columns = opts.columns || [];
			var pageSize = opts.pageSize || 0;
			var expandable = !!opts.expandable;
			var searchEnabled = !!opts.search;
			var items = Array.isArray(data) ? data.slice() : [];
			var expandedIdx = state._expandedIdx;

			// --- Search filtering ---
			if (searchEnabled && state.search) {
				var q = state.search.toLowerCase();
				items = items.filter(function(item) {
					for (var ci = 0; ci < columns.length; ci++) {
						var col = columns[ci];
						if (col.type === "bool") continue;
						var v = item[col.key];
						if (v !== null && v !== undefined && String(v).toLowerCase().indexOf(q) >= 0) return true;
					}
					return false;
				});
			}

			// --- Column filters ---
			if (state.filters) {
				for (var fk in state.filters) {
					if (!state.filters.hasOwnProperty(fk)) continue;
					(function(col, f) {
						items = items.filter(function(item) {
							var val = item[col];
							if (f.op === "contains") return val !== null && val !== undefined && String(val).toLowerCase().indexOf(String(f.value).toLowerCase()) >= 0;
							if (f.op === "startsWith") return val !== null && val !== undefined && String(val).toLowerCase().indexOf(String(f.value).toLowerCase()) === 0;
							if (f.op === "equals") return String(val) === String(f.value);
							if (f.op === "eq") return parseFloat(val) === parseFloat(f.value);
							if (f.op === "gt") return parseFloat(val) > parseFloat(f.value);
							if (f.op === "lt") return parseFloat(val) < parseFloat(f.value);
							if (f.op === "gte") return parseFloat(val) >= parseFloat(f.value);
							if (f.op === "lte") return parseFloat(val) <= parseFloat(f.value);
							if (f.op === "between") return parseFloat(val) >= parseFloat(f.from) && parseFloat(val) <= parseFloat(f.to);
							if (f.op === "in" && f.values) return f.values.indexOf(String(val)) >= 0;
							if (f.op === "bool") return Boolean(val) === (f.value === true || f.value === "true");
							if (f.op === "dateRange") {
								var d = new Date(val).getTime();
								var from = f.from ? new Date(f.from).getTime() : -Infinity;
								var to = f.to ? new Date(f.to + "T23:59:59").getTime() : Infinity;
								return d >= from && d <= to;
							}
							return true;
						});
					})(fk, state.filters[fk]);
				}
			}

			// --- Sorting ---
			if (state.sort && state.sort.col) {
				var sortCol = state.sort.col;
				var sortDir = state.sort.dir === "desc" ? -1 : 1;
				var colDef = null;
				for (var si = 0; si < columns.length; si++) {
					if (columns[si].key === sortCol) { colDef = columns[si]; break; }
				}
				items.sort(function(a, b) {
					var av = a[sortCol], bv = b[sortCol];
					if (av === null || av === undefined) av = "";
					if (bv === null || bv === undefined) bv = "";
					if (colDef && (colDef.type === "number" || colDef.format === "amount")) {
						av = parseFloat(av) || 0;
						bv = parseFloat(bv) || 0;
					} else if (colDef && colDef.type === "date") {
						av = new Date(av).getTime() || 0;
						bv = new Date(bv).getTime() || 0;
					} else {
						av = String(av).toLowerCase();
						bv = String(bv).toLowerCase();
					}
					if (av < bv) return -1 * sortDir;
					if (av > bv) return 1 * sortDir;
					return 0;
				});
			}

			// --- Pagination ---
			var totalItems = items.length;
			var totalPages = pageSize > 0 ? Math.ceil(totalItems / pageSize) : 1;
			var currentPage = state.page || 0;
			if (currentPage >= totalPages) currentPage = Math.max(0, totalPages - 1);
			var pagedItems = pageSize > 0 ? items.slice(currentPage * pageSize, (currentPage + 1) * pageSize) : items;

			// --- Format cell ---
			function formatCell(item, col, idx) {
				var val = item[col.key];
				if (col.render) {
					try { var r = (new Function("item", "i", col.render))(item, idx); return __cel.html(r); } catch(e) { return __cfmt.escape(String(val || "")); }
				}
				if (col.type === "date") return __cfmt.date(val);
				if (col.format === "amount") return __cfmt.amount(val);
				if (col.type === "number") return __cfmt.number(val);
				if (col.type === "bool") {
					var iconName = __cfmt.bool(val);
					var iconCls = val ? "text-green-500" : "text-red-400";
					return __cel.icon(iconName, "text-base " + iconCls);
				}
				if (val === null || val === undefined) return "";
				return __cfmt.escape(String(val));
			}

			// --- Header cells ---
			var activeFilterDropdown = null;
			var headerCells = __cel.map(columns, function(col) {
				var headerChildren = [__cel.text(col.label || col.key)];
				var headerEvents = null;
				if (col.sortable) {
					if (state.sort && state.sort.col === col.key) {
						var dir = state.sort.dir;
						var arrowIcon = dir === "asc" ? "arrow_upward" : "arrow_downward";
						headerChildren.push(__cel.icon(arrowIcon, "text-xs ml-1 align-middle"));
					}
					var colKey = col.key;
					var zoneId = state.zoneId;
					headerEvents = __cel.on("click", "(function(){ var st = __clients['" + zoneId + "'].getState(); var s = st.sort || {}; var nd = 'asc'; if (s.col === '" + colKey + "') { nd = s.dir === 'asc' ? 'desc' : s.dir === 'desc' ? '' : 'asc'; } __clients['" + zoneId + "'].setState({ sort: { col: nd ? '" + colKey + "' : '', dir: nd }, page: 0 }); })()");
				}
				// Filter icon for filterable columns
				if (col.filterable) {
					var fKey = col.key;
					var fZone = state.zoneId;
					var hasFilter = state.filters && state.filters[fKey];
					var filterIconCls = hasFilter ? "text-blue-500" : "text-gray-400 dark:text-gray-500";
					headerChildren.push(__cel("button", {
						type: "button",
						class: "ml-1 align-middle inline-flex items-center " + filterIconCls
					}, [__cel.icon("tune", "text-sm")], {
						click: { act: "raw", js: "(function(e){ e.stopPropagation(); var st=__clients['" + fZone + "'].getState(); var open=st._filterOpen==='" + fKey + "'?null:'" + fKey + "'; __clients['" + fZone + "'].setState({_filterOpen:open}); })(event)" }
					}));
					// Render dropdown outside the table to avoid overflow clipping
					if (state._filterOpen === fKey) {
						activeFilterDropdown = renderFilterDropdown(col, state, fZone, data);
					}
				}
				var thJustify = "justify-start";
				if (col.cellClass) {
					if (col.cellClass.indexOf("text-right") >= 0) thJustify = "justify-end";
					else if (col.cellClass.indexOf("text-center") >= 0) thJustify = "justify-center";
				}
				var thCls = "p-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider" + (col.sortable ? " cursor-pointer select-none" : "") + (col.class ? " " + col.class : "");
				var innerWrap = __cel("span", { class: "inline-flex items-center whitespace-nowrap " + thJustify + " w-full" }, headerChildren);
				return __cel("th", { class: thCls, "data-col-key": col.key }, [innerWrap], headerEvents);
			});

			// --- Body rows ---
			var bodyRows = [];
			for (var ri = 0; ri < pagedItems.length; ri++) {
				(function(item, ri) {
					var globalIdx = pageSize > 0 ? currentPage * pageSize + ri : ri;
					var cells = __cel.map(columns, function(col) {
						var cellVal = formatCell(item, col, globalIdx);
						var cellCls = "p-3 text-sm text-gray-700 dark:text-gray-300" + (col.cellClass ? " " + col.cellClass : "");
						var cellChildren = (typeof cellVal === "string") ? [__cel.text(cellVal)] : [cellVal];
						return __cel("td", { class: cellCls }, cellChildren);
					});

					var rowCls = "border-t border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50";
					var rowEvents = null;

					if (expandable) {
						var zid = state.zoneId;
						var gIdx = globalIdx;
						rowCls += " cursor-pointer";
						rowEvents = __cel.on("click", "(function(){ var st = __clients['" + zid + "'].getState(); var ex = st._expandedIdx === " + gIdx + " ? -1 : " + gIdx + "; __clients['" + zid + "'].setState({ _expandedIdx: ex }); })()");
					} else if (opts.onRowClick) {
						rowEvents = __cel.on("click", "(function(){ (" + opts.onRowClick + ")(this); }).call(" + JSON.stringify(item) + ")");
					}

					bodyRows.push(__cel("tr", { class: rowCls }, cells, rowEvents));

					// Expanded detail row
					if (expandable && expandedIdx === globalIdx) {
						var detailContent = null;
						var detailCache = state._detailCache || {};
						var cacheKey = item.id || item.key || String(globalIdx);
						var activeTab = state._activeTab || 0;

						if (opts.detailTabs && opts.detailTabs.length > 0) {
							// Tabbed detail view
							var tabs = opts.detailTabs;
							var tabHeaders = tabs.map(function(tab, ti) {
								var isActive = ti === activeTab;
								return __cel("button", {
									type: "button",
									class: "px-3 py-1.5 text-sm rounded-t " + (isActive ? "bg-white dark:bg-gray-900 border border-b-0 border-gray-200 dark:border-gray-700 font-medium text-gray-800 dark:text-gray-200" : "text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300")
								}, [__cel.text(tab.label)], {
									click: { act: "raw", js: "(function(e){ e.stopPropagation(); __clients['" + state.zoneId + "'].setState({_activeTab:" + ti + "}); })(event)" }
								});
							});

							var tabKey = cacheKey + "_tab" + activeTab;
							var tabData = detailCache[tabKey];
							var tabContent = null;

							if (tabData === undefined && tabs[activeTab].source) {
								// Fetch tab data
								var tabSrc = tabs[activeTab].source.replace("{id}", encodeURIComponent(cacheKey)).replace("{key}", encodeURIComponent(cacheKey));
								tabContent = __cel.div("animate-pulse p-4", [
									__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-3/4 mb-2", []),
									__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-1/2", [])
								]);
								(function(tk, zid) {
									__capi.get(tabSrc).then(function(d) {
										var st = __clients[zid].getState();
										var dc = Object.assign({}, st._detailCache || {});
										dc[tk] = d;
										__clients[zid].setState({_detailCache: dc});
									}).catch(function() {
										var st = __clients[zid].getState();
										var dc = Object.assign({}, st._detailCache || {});
										dc[tk] = {_error: true};
										__clients[zid].setState({_detailCache: dc});
									});
								})(tabKey, state.zoneId);
							} else if (tabData && tabData._error) {
								tabContent = __cel.div("text-red-500 text-sm p-2", [__cel.text("Failed to load")]);
							} else if (tabData !== undefined) {
								if (opts.renderDetail) {
									try { tabContent = (new Function("item", "detail", opts.renderDetail))(item, tabData); } catch(e) { tabContent = __cel.text("Error: " + e.message); }
								} else {
									var pairs = [];
									for (var tk in tabData) {
										if (tabData.hasOwnProperty(tk) && tk !== "_error") {
											pairs.push(__cel.div("mb-1", [
												__cel("span", { class: "font-medium text-gray-500 dark:text-gray-400 mr-2" }, [__cel.text(tk + ":")]),
												__cel("span", null, [__cel.text(__cfmt.escape(String(tabData[tk] !== null && tabData[tk] !== undefined ? tabData[tk] : "")))])
											]));
										}
									}
									tabContent = __cel.div("text-sm text-gray-600 dark:text-gray-400", pairs);
								}
							} else {
								tabContent = __cel.div("text-sm text-gray-400 p-2", [__cel.text("No data")]);
							}

							detailContent = __cel.div("", [
								__cel.div("flex gap-1 border-b border-gray-200 dark:border-gray-700 mb-3", tabHeaders),
								tabContent
							]);
						} else if (opts.detailSource) {
							// Async detail loading
							var detail = detailCache[cacheKey];
							if (detail === undefined) {
								detailContent = __cel.div("animate-pulse p-2", [
									__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-3/4 mb-2", []),
									__cel.div("bg-gray-200 dark:bg-gray-700 h-4 rounded w-1/2", [])
								]);
								var src = opts.detailSource.replace("{id}", encodeURIComponent(cacheKey)).replace("{key}", encodeURIComponent(cacheKey));
								(function(ck, zid) {
									__capi.get(src).then(function(d) {
										var st = __clients[zid].getState();
										var dc = Object.assign({}, st._detailCache || {});
										dc[ck] = d;
										__clients[zid].setState({_detailCache: dc});
									}).catch(function() {
										var st = __clients[zid].getState();
										var dc = Object.assign({}, st._detailCache || {});
										dc[ck] = {_error: true};
										__clients[zid].setState({_detailCache: dc});
									});
								})(cacheKey, state.zoneId);
							} else if (detail._error) {
								detailContent = __cel.div("text-red-500 text-sm p-2", [__cel.text("Failed to load detail")]);
							} else {
								if (opts.renderDetail) {
									try { detailContent = (new Function("item", "detail", opts.renderDetail))(item, detail); } catch(e) { detailContent = __cel.text("Error: " + e.message); }
								} else {
									var pairs = [];
									for (var dk in detail) {
										if (detail.hasOwnProperty(dk) && dk !== "_error") {
											pairs.push(__cel.div("mb-1", [
												__cel("span", { class: "font-medium text-gray-500 dark:text-gray-400 mr-2" }, [__cel.text(dk + ":")]),
												__cel("span", null, [__cel.text(__cfmt.escape(String(detail[dk] !== null && detail[dk] !== undefined ? detail[dk] : "")))])
											]));
										}
									}
									detailContent = __cel.div("text-sm text-gray-600 dark:text-gray-400", pairs);
								}
							}
						} else if (opts.renderDetail) {
							try { detailContent = (new Function("item", opts.renderDetail))(item); } catch(e) { detailContent = __cel.text("Error: " + e.message); }
						} else {
							var pairs = [];
							for (var dk in item) {
								if (item.hasOwnProperty(dk)) {
									pairs.push(__cel.div("mb-1", [
										__cel("span", { class: "font-medium text-gray-500 dark:text-gray-400 mr-2" }, [__cel.text(dk + ":")]),
										__cel("span", null, [__cel.text(__cfmt.escape(String(item[dk] !== null && item[dk] !== undefined ? item[dk] : "")))])
									]));
								}
							}
							detailContent = __cel.div("text-sm text-gray-600 dark:text-gray-400", pairs);
						}
						bodyRows.push(__cel("tr", { class: "border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/30" }, [
							__cel("td", { class: "p-4", colspan: String(columns.length) }, [detailContent])
						]));
					}
				})(pagedItems[ri], ri);
			}

			// --- Search bar ---
			var searchBar = null;
			if (searchEnabled) {
				var zSearch = state.zoneId;
				searchBar = __cel("input", {
					class: "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-64 mb-4",
					type: "text",
					placeholder: "Search...",
					value: state.search || ""
				}, null, __cel.on("input", "__clients['" + zSearch + "'].setState({ search: this.value, page: 0 })"));
			}

			// --- Table element ---
			var table = __cel.div("bg-white dark:bg-gray-900 rounded-lg shadow overflow-hidden", [
				__cel("table", { class: "w-full" }, [
					__cel("thead", null, [
						__cel("tr", { class: "bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700" }, headerCells)
					]),
					__cel("tbody", null, bodyRows.length > 0 ? bodyRows : [
						__cel("tr", null, [
							__cel("td", { class: "p-8 text-center text-gray-400 dark:text-gray-500", colspan: String(columns.length) }, [
								__cel.text("No results")
							])
						])
					])
				])
			]);

			// --- Pagination controls ---
			var pagination = null;
			if (pageSize > 0 && totalPages > 1) {
				var zPage = state.zoneId;
				var pageButtons = [];

				// Previous
				pageButtons.push(__cel("button", {
					class: "px-3 py-1 text-sm rounded border border-gray-300 dark:border-gray-600" + (currentPage === 0 ? " opacity-40 cursor-not-allowed" : " hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer"),
					disabled: currentPage === 0 ? "true" : null
				}, [__cel.text("Prev")], currentPage > 0 ? __cel.on("click", "__clients['" + zPage + "'].setState({ page: " + (currentPage - 1) + " })") : null));

				// Page numbers
				var startP = Math.max(0, currentPage - 2);
				var endP = Math.min(totalPages, startP + 5);
				if (endP - startP < 5) startP = Math.max(0, endP - 5);
				for (var pi = startP; pi < endP; pi++) {
					(function(p) {
						var active = p === currentPage;
						pageButtons.push(__cel("button", {
							class: "px-3 py-1 text-sm rounded" + (active ? " bg-blue-600 text-white" : " border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer")
						}, [__cel.text(String(p + 1))], active ? null : __cel.on("click", "__clients['" + zPage + "'].setState({ page: " + p + " })")));
					})(pi);
				}

				// Next
				pageButtons.push(__cel("button", {
					class: "px-3 py-1 text-sm rounded border border-gray-300 dark:border-gray-600" + (currentPage >= totalPages - 1 ? " opacity-40 cursor-not-allowed" : " hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer"),
					disabled: currentPage >= totalPages - 1 ? "true" : null
				}, [__cel.text("Next")], currentPage < totalPages - 1 ? __cel.on("click", "__clients['" + zPage + "'].setState({ page: " + (currentPage + 1) + " })") : null));

				// Info text
				var fromItem = currentPage * pageSize + 1;
				var toItem = Math.min((currentPage + 1) * pageSize, totalItems);
				var infoText = __cel.span("text-sm text-gray-500 dark:text-gray-400", [
					__cel.text(fromItem + "-" + toItem + " of " + totalItems)
				]);

				pagination = __cel.div("flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700", [
					infoText,
					__cel.div("flex items-center gap-1", pageButtons)
				]);
			}

			// --- Root ---
			var rootChildren = [];
			if (searchBar) rootChildren.push(searchBar);
			rootChildren.push(table);
			if (activeFilterDropdown) rootChildren.push(activeFilterDropdown);
			if (pagination) rootChildren.push(pagination);

			return __cel.div("", rootChildren);
		}

		__cregister("table", tableComponent);
	})();
`)

var __cfilter = Trim(`
	var __cfilter = (function(){
		// Apply all filters/sort/search/pagination to data array
		function applyToData(data, filterState, columns) {
			var result = data ? data.slice() : [];
			
			// Search
			if (filterState.search) {
				var q = filterState.search.toLowerCase();
				result = result.filter(function(item) {
					for (var i = 0; i < columns.length; i++) {
						var val = item[columns[i].key];
						if (val !== null && val !== undefined && String(val).toLowerCase().indexOf(q) >= 0) return true;
					}
					return false;
				});
			}
			
			// Column filters
			if (filterState.filters) {
				for (var col in filterState.filters) {
					if (!filterState.filters.hasOwnProperty(col)) continue;
					var f = filterState.filters[col];
					result = result.filter(function(item) {
						var val = item[col];
						if (f.op === "contains") return val && String(val).toLowerCase().indexOf(String(f.value).toLowerCase()) >= 0;
						if (f.op === "equals") return String(val) === String(f.value);
						if (f.op === "gt") return parseFloat(val) > parseFloat(f.value);
						if (f.op === "lt") return parseFloat(val) < parseFloat(f.value);
						if (f.op === "gte") return parseFloat(val) >= parseFloat(f.value);
						if (f.op === "lte") return parseFloat(val) <= parseFloat(f.value);
						if (f.op === "between") {
							var n = parseFloat(val);
							return n >= parseFloat(f.from) && n <= parseFloat(f.to);
						}
						if (f.op === "in" && f.values) {
							return f.values.indexOf(String(val)) >= 0;
						}
						if (f.op === "bool") return Boolean(val) === Boolean(f.value);
						if (f.op === "dateRange") {
							var d = new Date(val).getTime();
							var from = f.from ? new Date(f.from).getTime() : -Infinity;
							var to = f.to ? new Date(f.to).getTime() : Infinity;
							return d >= from && d <= to;
						}
						return true;
					});
				}
			}
			
			// Sort
			if (filterState.sort && filterState.sort.col) {
				var sortCol = filterState.sort.col;
				var sortDir = filterState.sort.dir === "desc" ? -1 : 1;
				// Find column definition to determine type
				var colDef = null;
				for (var ci = 0; ci < columns.length; ci++) {
					if (columns[ci].key === sortCol) { colDef = columns[ci]; break; }
				}
				result.sort(function(a, b) {
					var va = a[sortCol], vb = b[sortCol];
					if (va === null || va === undefined) va = "";
					if (vb === null || vb === undefined) vb = "";
					if (colDef && (colDef.type === "number" || colDef.format === "amount")) {
						return (parseFloat(va) - parseFloat(vb)) * sortDir;
					}
					if (colDef && colDef.type === "date") {
						return (new Date(va).getTime() - new Date(vb).getTime()) * sortDir;
					}
					return String(va).localeCompare(String(vb)) * sortDir;
				});
			}
			
			return result;
		}
		
		// Paginate data
		function paginate(data, page, pageSize) {
			if (!pageSize || pageSize <= 0) return { items: data, totalPages: 1, total: data.length };
			var total = data.length;
			var totalPages = Math.max(1, Math.ceil(total / pageSize));
			var p = Math.max(0, Math.min(page || 0, totalPages - 1));
			var start = p * pageSize;
			return { items: data.slice(start, start + pageSize), totalPages: totalPages, page: p, total: total };
		}
		
		// Convert filter state to URL query string
		function toQueryString(filterState) {
			var params = {};
			if (filterState.search) params._search = filterState.search;
			if (filterState.sort && filterState.sort.col) {
				params._sort = filterState.sort.col;
				params._dir = filterState.sort.dir || "asc";
			}
			if (filterState.page > 0) params._page = String(filterState.page);
			return params;
		}
		
		// Cycle sort direction: none -> asc -> desc -> none
		function cycleSort(current, col) {
			if (!current || current.col !== col) return { col: col, dir: "asc" };
			if (current.dir === "asc") return { col: col, dir: "desc" };
			return { col: "", dir: "" };
		}
		
		return {
			applyToData: applyToData,
			paginate: paginate,
			toQueryString: toQueryString,
			cycleSort: cycleSort
		};
	})();
`)

var __cfilterbar = Trim(`
	function __cfilterbar(zoneId, state, columns) {
		var parts = [];
		
		// Search input (if any column is filterable or search is enabled)
		var hasSearch = false;
		for (var i = 0; i < columns.length; i++) {
			if (columns[i].filterable || columns[i].searchable) { hasSearch = true; break; }
		}
		
		if (hasSearch) {
			parts.push(__cel("input", {
				type: "text",
				class: "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-64 placeholder-gray-400",
				placeholder: "Search...",
				value: state.search || ""
			}, [], { input: { act: "raw", js: "__clients['" + zoneId + "'].setState({search:this.value,page:0})" } }));
		}
		
		// Active filter chips
		if (state.filters) {
			for (var col in state.filters) {
				if (!state.filters.hasOwnProperty(col)) continue;
				var f = state.filters[col];
				// Find label
				var label = col;
				for (var j = 0; j < columns.length; j++) {
					if (columns[j].key === col) { label = columns[j].label || col; break; }
				}
				var chipText = label + ": ";
				if (f.op === "contains") chipText += "contains '" + f.value + "'";
				else if (f.op === "equals") chipText += "= " + f.value;
				else if (f.op === "between") chipText += f.from + " - " + f.to;
				else if (f.op === "in") chipText += f.values.join(", ");
				else if (f.op === "bool") chipText += f.value ? "Yes" : "No";
				else chipText += f.op + " " + (f.value || "");
				
				parts.push(__cel.div(
					"inline-flex items-center gap-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs px-2 py-1 rounded-full",
					[
						__cel.text(chipText),
						__cel("button", {
							class: "ml-1 hover:text-blue-900 dark:hover:text-blue-100",
							type: "button"
						}, [__cel.icon("close", "text-xs")], {
							click: { act: "raw", js: "(function(){var s=__clients['" + zoneId + "'].getState();var f=Object.assign({},s.filters);delete f['" + col + "'];__clients['" + zoneId + "'].setState({filters:f,page:0});})()" }
						})
					]
				));
			}
		}
		
		// Clear all button (if any filters active)
		var hasFilters = state.filters && Object.keys(state.filters).length > 0;
		if (hasFilters || (state.search && state.search.length > 0)) {
			parts.push(__cel("button", {
				type: "button",
				class: "text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 underline"
			}, [__cel.text("Clear all")], {
				click: { act: "raw", js: "__clients['" + zoneId + "'].setState({filters:{},search:'',page:0})" }
			}));
		}
		
		if (parts.length === 0) return null;
		return __cel.div("flex flex-wrap items-center gap-2 mb-4", parts);
	}
`)

var __cpagination = Trim(`
	function __cpagination(zoneId, page, totalPages, pageSize, total) {
		if (totalPages <= 1) return null;
		
		var pages = [];
		
		// Prev button
		pages.push(__cel("button", {
			type: "button",
			class: "px-3 py-1 text-sm rounded " + (page <= 0 ? "text-gray-300 dark:text-gray-600 cursor-not-allowed" : "text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800")
		}, [__cel.icon("chevron_left", "text-sm")], page > 0 ? {
			click: { act: "raw", js: "__clients['" + zoneId + "'].setState({page:" + (page - 1) + "})" }
		} : null));
		
		// Page numbers with ellipsis
		var start = Math.max(0, page - 2);
		var end = Math.min(totalPages, page + 3);
		
		if (start > 0) {
			pages.push(__cel("button", {
				type: "button",
				class: "px-3 py-1 text-sm rounded text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
			}, [__cel.text("1")], {
				click: { act: "raw", js: "__clients['" + zoneId + "'].setState({page:0})" }
			}));
			if (start > 1) {
				pages.push(__cel.span("px-1 text-gray-400", [__cel.text("...")]));
			}
		}
		
		for (var i = start; i < end; i++) {
			var active = i === page;
			pages.push(__cel("button", {
				type: "button",
				class: "px-3 py-1 text-sm rounded " + (active ? "bg-blue-600 text-white" : "text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800")
			}, [__cel.text(String(i + 1))], active ? null : {
				click: { act: "raw", js: "__clients['" + zoneId + "'].setState({page:" + i + "})" }
			}));
		}
		
		if (end < totalPages) {
			if (end < totalPages - 1) {
				pages.push(__cel.span("px-1 text-gray-400", [__cel.text("...")]));
			}
			pages.push(__cel("button", {
				type: "button",
				class: "px-3 py-1 text-sm rounded text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
			}, [__cel.text(String(totalPages))], {
				click: { act: "raw", js: "__clients['" + zoneId + "'].setState({page:" + (totalPages - 1) + "})" }
			}));
		}
		
		// Next button
		pages.push(__cel("button", {
			type: "button",
			class: "px-3 py-1 text-sm rounded " + (page >= totalPages - 1 ? "text-gray-300 dark:text-gray-600 cursor-not-allowed" : "text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800")
		}, [__cel.icon("chevron_right", "text-sm")], page < totalPages - 1 ? {
			click: { act: "raw", js: "__clients['" + zoneId + "'].setState({page:" + (page + 1) + "})" }
		} : null));
		
		// Record count
		var from = page * pageSize + 1;
		var to = Math.min((page + 1) * pageSize, total);
		
		return __cel.div("flex items-center justify-between mt-4", [
			__cel.div("text-xs text-gray-500 dark:text-gray-400", [
				__cel.text(from + "-" + to + " of " + total)
			]),
			__cel.div("flex items-center gap-1", pages)
		]);
	}
`)

var __cchart = Trim(`
	(function(){
		var defaultColors = ["#3b82f6","#8b5cf6","#06b6d4","#10b981","#f59e0b","#ef4444","#ec4899","#6366f1"];

		function normalize(data) {
			if (!data || !data.length) return [];
			var out = [];
			for (var i = 0; i < data.length; i++) {
				var d = data[i];
				if (typeof d === "object" && d !== null && d.hasOwnProperty("value")) {
					var item = { label: d.label !== undefined ? String(d.label) : String(i), value: parseFloat(d.value) || 0 };
					if (d.value2 !== undefined) item.value2 = parseFloat(d.value2) || 0;
					if (d.series && Array.isArray(d.series)) item.series = d.series;
					out.push(item);
				} else {
					out.push({ label: String(i), value: parseFloat(d) || 0 });
				}
			}
			return out;
		}

		function fmtVal(v, fmt) {
			if (fmt === "amount") return __cfmt.amount(v);
			if (fmt === "number") return __cfmt.number(v);
			return String(Math.round(v * 100) / 100);
		}

		function maxVal(data) {
			var m = 0;
			for (var i = 0; i < data.length; i++) {
				if (data[i].value > m) m = data[i].value;
				if (data[i].value2 !== undefined && data[i].value2 > m) m = data[i].value2;
				if (data[i].series) {
					for (var s = 0; s < data[i].series.length; s++) {
						var sv = parseFloat(data[i].series[s].value) || 0;
						if (sv > m) m = sv;
					}
				}
			}
			return m || 1;
		}

		function niceMax(v) {
			if (v <= 0) return 1;
			var mag = Math.pow(10, Math.floor(Math.log10(v)));
			var norm = v / mag;
			var nice;
			if (norm <= 1) nice = 1;
			else if (norm <= 2) nice = 2;
			else if (norm <= 5) nice = 5;
			else nice = 10;
			return nice * mag;
		}

		function gridLines(max, count) {
			var step = max / count;
			var lines = [];
			for (var i = 0; i <= count; i++) {
				lines.push(Math.round(step * i * 100) / 100);
			}
			return lines;
		}

		function color(colors, i) {
			return colors[i % colors.length];
		}

		function renderBar(data, opts) {
			var w = opts.width || 600;
			var h = opts.height || 300;
			var cols = opts.colors || defaultColors;
			var showValues = opts.showValues !== false;
			var showLabels = opts.showLabels !== false;
			var pt = 20, pr = 20, pb = 40, pl = 50;
			var cw = w - pl - pr;
			var ch = h - pt - pb;
			var n = data.length;
			if (n === 0) return __cel.div("", [__cel.text("No data")]);
			var gap = opts.gap || Math.max(4, Math.floor(cw / n * 0.2));
			var bw = opts.barWidth || Math.max(8, Math.floor((cw - gap * (n + 1)) / n));
			var mx = maxVal(data);
			var nm = niceMax(mx);
			var gcount = 4;
			var gl = gridLines(nm, gcount);
			var children = [];
			for (var g = 0; g < gl.length; g++) {
				var gy = pt + ch - (gl[g] / nm * ch);
				children.push(__cel("line", { x1: String(pl), y1: String(gy), x2: String(w - pr), y2: String(gy), stroke: "#e5e7eb", "stroke-width": "1" }, null));
				children.push(__cel("text", { x: String(pl - 6), y: String(gy + 4), "text-anchor": "end", fill: "#6b7280", "font-size": "11", "font-family": "Inter, sans-serif" }, [fmtVal(gl[g], opts.valueFormat)]));
			}
			// Detect multi-series
			var hasSeries = data.length > 0 && (data[0].value2 !== undefined || (data[0].series && data[0].series.length > 0));
			var seriesCount = 1;
			var seriesNames = [opts.seriesName || "Series 1"];
			if (hasSeries) {
				if (data[0].series) {
					seriesCount = data[0].series.length;
					seriesNames = data[0].series.map(function(s) { return s.name || "Series"; });
				} else {
					seriesCount = 2;
					seriesNames = [opts.seriesName || "Current", opts.series2Name || "Previous"];
				}
			}
			var groupW = bw * seriesCount + (seriesCount > 1 ? 2 * (seriesCount - 1) : 0);
			var totalBarSpace = groupW * n + gap * (n + 1);
			var offsetX = pl + (cw - totalBarSpace) / 2;
			for (var i = 0; i < n; i++) {
				var groupX = offsetX + gap * (i + 1) + groupW * i;
				for (var si = 0; si < seriesCount; si++) {
					var sv;
					if (data[i].series) {
						sv = parseFloat(data[i].series[si].value) || 0;
					} else if (si === 0) {
						sv = data[i].value;
					} else {
						sv = data[i].value2 || 0;
					}
					var sbx = groupX + si * (bw + 2);
					var sbh = sv / nm * ch;
					var sby = pt + ch - sbh;
					var barCol = seriesCount > 1 ? color(cols, si) : color(cols, i);
					children.push(__cel("rect", { x: String(sbx), y: String(sby), width: String(bw), height: String(Math.max(sbh, 1)), fill: barCol, rx: "3" }, [
						__cel("title", null, [data[i].label + ": " + fmtVal(sv, opts.valueFormat)])
					]));
					if (showValues && seriesCount <= 2) {
						children.push(__cel("text", { x: String(sbx + bw / 2), y: String(sby - 6), "text-anchor": "middle", fill: "#374151", "font-size": "11", "font-family": "Inter, sans-serif" }, [fmtVal(sv, opts.valueFormat)]));
					}
				}
				if (showLabels) {
					children.push(__cel("text", { x: String(groupX + groupW / 2), y: String(h - pb + 16), "text-anchor": "middle", fill: "#6b7280", "font-size": "11", "font-family": "Inter, sans-serif" }, [data[i].label]));
				}
			}
			children.push(__cel("line", { x1: String(pl), y1: String(pt + ch), x2: String(w - pr), y2: String(pt + ch), stroke: "#d1d5db", "stroke-width": "1" }, null));
			// Legend for multi-series
			var legendEl = null;
			if (seriesCount > 1) {
				var legendItems = [];
				for (var li = 0; li < seriesCount; li++) {
					legendItems.push(__cel.div("flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400", [
						__cel("span", { style: "display:inline-block;width:10px;height:10px;border-radius:2px;background:" + color(cols, li) }, null),
						__cel.text(seriesNames[li])
					]));
				}
				legendEl = __cel.div("flex gap-4 justify-center mt-2", legendItems);
			}
			var svgEl = __cel("svg", { viewBox: "0 0 " + w + " " + h, xmlns: "http://www.w3.org/2000/svg", preserveAspectRatio: "xMidYMid meet", style: "width:100%;height:" + h + "px;max-height:" + h + "px" }, children);
			return legendEl ? __cel.div("", [svgEl, legendEl]) : svgEl;
		}

		function renderArea(data, opts) {
			var w = opts.width || 600;
			var h = opts.height || 300;
			var cols = opts.colors || defaultColors;
			var showValues = opts.showValues !== false;
			var showLabels = opts.showLabels !== false;
			var pt = 20, pr = 20, pb = 40, pl = 50;
			var cw = w - pl - pr;
			var ch = h - pt - pb;
			var n = data.length;
			if (n === 0) return __cel.div("", [__cel.text("No data")]);
			var mx = maxVal(data);
			var nm = niceMax(mx);
			var gcount = 4;
			var gl = gridLines(nm, gcount);
			var gradId = "areaGrad_" + Math.random().toString(36).substring(2, 8);
			var children = [];
			children.push(__cel("defs", null, [
				__cel("linearGradient", { id: gradId, x1: "0", y1: "0", x2: "0", y2: "1" }, [
					__cel("stop", { offset: "0%", "stop-color": cols[0], "stop-opacity": "0.4" }, null),
					__cel("stop", { offset: "100%", "stop-color": cols[0], "stop-opacity": "0.05" }, null)
				])
			]));
			for (var g = 0; g < gl.length; g++) {
				var gy = pt + ch - (gl[g] / nm * ch);
				children.push(__cel("line", { x1: String(pl), y1: String(gy), x2: String(w - pr), y2: String(gy), stroke: "#e5e7eb", "stroke-width": "1" }, null));
				children.push(__cel("text", { x: String(pl - 6), y: String(gy + 4), "text-anchor": "end", fill: "#6b7280", "font-size": "11", "font-family": "Inter, sans-serif" }, [fmtVal(gl[g], opts.valueFormat)]));
			}
			var points = [];
			for (var i = 0; i < n; i++) {
				var step = n > 1 ? cw / (n - 1) : cw / 2;
				var px = pl + (n > 1 ? step * i : cw / 2);
				var py = pt + ch - (data[i].value / nm * ch);
				points.push({ x: px, y: py });
			}
			var areaPath = "M" + points[0].x + "," + points[0].y;
			for (var i = 1; i < points.length; i++) {
				areaPath += " L" + points[i].x + "," + points[i].y;
			}
			areaPath += " L" + points[points.length - 1].x + "," + (pt + ch);
			areaPath += " L" + points[0].x + "," + (pt + ch) + " Z";
			children.push(__cel("path", { d: areaPath, fill: "url(#" + gradId + ")" }, null));
			var linePath = "M" + points[0].x + "," + points[0].y;
			for (var i = 1; i < points.length; i++) {
				linePath += " L" + points[i].x + "," + points[i].y;
			}
			children.push(__cel("path", { d: linePath, fill: "none", stroke: cols[0], "stroke-width": "2.5", "stroke-linecap": "round", "stroke-linejoin": "round" }, null));
			for (var i = 0; i < points.length; i++) {
				children.push(__cel("circle", { cx: String(points[i].x), cy: String(points[i].y), r: "4", fill: "#fff", stroke: cols[0], "stroke-width": "2" }, [
					__cel("title", null, [data[i].label + ": " + fmtVal(data[i].value, opts.valueFormat)])
				]));
				if (showValues) {
					children.push(__cel("text", { x: String(points[i].x), y: String(points[i].y - 10), "text-anchor": "middle", fill: "#374151", "font-size": "11", "font-family": "Inter, sans-serif" }, [fmtVal(data[i].value, opts.valueFormat)]));
				}
				if (showLabels) {
					children.push(__cel("text", { x: String(points[i].x), y: String(h - pb + 16), "text-anchor": "middle", fill: "#6b7280", "font-size": "11", "font-family": "Inter, sans-serif" }, [data[i].label]));
				}
			}
			children.push(__cel("line", { x1: String(pl), y1: String(pt + ch), x2: String(w - pr), y2: String(pt + ch), stroke: "#d1d5db", "stroke-width": "1" }, null));
			return __cel("svg", { viewBox: "0 0 " + w + " " + h, xmlns: "http://www.w3.org/2000/svg", preserveAspectRatio: "xMidYMid meet", style: "width:100%;height:" + h + "px;max-height:" + h + "px" }, children);
		}

		function renderHbar(data, opts) {
			var cols = opts.colors || defaultColors;
			var showValues = opts.showValues !== false;
			var n = data.length;
			if (n === 0) return __cel.div("", [__cel.text("No data")]);
			var barH = opts.barWidth || 24;
			var gap = opts.gap || 8;
			var pt = 10, pr = 60, pb = 10, pl = 120;
			var h = opts.height || (pt + pb + n * (barH + gap) - gap);
			var w = opts.width || 600;
			var cw = w - pl - pr;
			var mx = maxVal(data);
			var nm = niceMax(mx);
			var children = [];
			children.push(__cel("line", { x1: String(pl), y1: String(pt), x2: String(pl), y2: String(h - pb), stroke: "#d1d5db", "stroke-width": "1" }, null));
			for (var i = 0; i < n; i++) {
				var by = pt + i * (barH + gap);
				var bw = Math.max(data[i].value / nm * cw, 1);
				children.push(__cel("rect", { x: String(pl), y: String(by), width: String(bw), height: String(barH), fill: color(cols, i), rx: "3" }, [
					__cel("title", null, [data[i].label + ": " + fmtVal(data[i].value, opts.valueFormat)])
				]));
				children.push(__cel("text", { x: String(pl - 8), y: String(by + barH / 2 + 4), "text-anchor": "end", fill: "#374151", "font-size": "12", "font-family": "Inter, sans-serif" }, [data[i].label]));
				if (showValues) {
					children.push(__cel("text", { x: String(pl + bw + 8), y: String(by + barH / 2 + 4), "text-anchor": "start", fill: "#6b7280", "font-size": "11", "font-family": "Inter, sans-serif" }, [fmtVal(data[i].value, opts.valueFormat)]));
				}
			}
			return __cel("svg", { viewBox: "0 0 " + w + " " + h, xmlns: "http://www.w3.org/2000/svg", preserveAspectRatio: "xMidYMid meet", style: "width:100%;height:" + h + "px;max-height:" + h + "px" }, children);
		}

		function renderDonut(data, opts) {
			var cols = opts.colors || defaultColors;
			var n = data.length;
			if (n === 0) return __cel.div("", [__cel.text("No data")]);
			var w = opts.width || 300;
			var h = opts.height || 300;
			var cx = w / 2;
			var cy = h / 2;
			var outerR = Math.min(cx, cy) - 10;
			var innerR = opts.innerRadius !== undefined ? opts.innerRadius : outerR * 0.6;
			var total = 0;
			for (var i = 0; i < n; i++) total += data[i].value;
			if (total === 0) return __cel.div("", [__cel.text("No data")]);
			var children = [];
			var angle = -Math.PI / 2;
			for (var i = 0; i < n; i++) {
				var slice = data[i].value / total * Math.PI * 2;
				if (slice < 0.001) { angle += slice; continue; }
				var x1o = cx + outerR * Math.cos(angle);
				var y1o = cy + outerR * Math.sin(angle);
				var x1i = cx + innerR * Math.cos(angle);
				var y1i = cy + innerR * Math.sin(angle);
				var x2o = cx + outerR * Math.cos(angle + slice);
				var y2o = cy + outerR * Math.sin(angle + slice);
				var x2i = cx + innerR * Math.cos(angle + slice);
				var y2i = cy + innerR * Math.sin(angle + slice);
				var large = slice > Math.PI ? 1 : 0;
				var d = "M" + x1o.toFixed(2) + "," + y1o.toFixed(2) +
					" A" + outerR + "," + outerR + " 0 " + large + " 1 " + x2o.toFixed(2) + "," + y2o.toFixed(2) +
					" L" + x2i.toFixed(2) + "," + y2i.toFixed(2) +
					" A" + innerR + "," + innerR + " 0 " + large + " 0 " + x1i.toFixed(2) + "," + y1i.toFixed(2) +
					" Z";
				var pctTip = Math.round(data[i].value / total * 100);
				children.push(__cel("path", { d: d, fill: color(cols, i) }, [
					__cel("title", null, [data[i].label + ": " + fmtVal(data[i].value, opts.valueFormat) + " (" + pctTip + "%)"])
				]));
				angle += slice;
			}
			children.push(__cel("text", { x: String(cx), y: String(cy - 6), "text-anchor": "middle", fill: "#374151", "font-size": "20", "font-weight": "600", "font-family": "Inter, sans-serif" }, [fmtVal(total, opts.valueFormat)]));
			children.push(__cel("text", { x: String(cx), y: String(cy + 14), "text-anchor": "middle", fill: "#9ca3af", "font-size": "12", "font-family": "Inter, sans-serif" }, ["Total"]));
			var svgEl = __cel("svg", { viewBox: "0 0 " + w + " " + h, xmlns: "http://www.w3.org/2000/svg", preserveAspectRatio: "xMidYMid meet", style: "width:100%;height:" + h + "px;max-height:" + h + "px" }, children);
			var legendItems = [];
			for (var i = 0; i < n; i++) {
				var pct = Math.round(data[i].value / total * 100);
				legendItems.push(__cel.div("flex items-center gap-2 text-sm", [
					__cel("span", { style: "display:inline-block;width:12px;height:12px;border-radius:2px;background:" + color(cols, i) + ";flex-shrink:0" }, null),
					__cel.span("text-gray-700", [data[i].label]),
					__cel.span("text-gray-400 ml-auto", [fmtVal(data[i].value, opts.valueFormat) + " (" + pct + "%)"])
				]));
			}
			var legend = __cel.div("flex flex-col gap-1 mt-2", legendItems);
			return __cel.div("flex flex-col items-center", [svgEl, legend]);
		}

		var renderers = {
			bar: renderBar,
			area: renderArea,
			hbar: renderHbar,
			donut: renderDonut
		};

		function chartComponent(data, state, opts) {
			if (!opts) opts = {};
			var type = opts.type || "bar";
			var renderer = renderers[type];
			if (!renderer) return __cel.div("text-red-500", [__cel.text("Unknown chart type: " + type)]);
			return renderer(normalize(data), opts);
		}

		__cregister("chart", chartComponent);
	})();
`)

var __ckpi = Trim(`
	(function(){
		function kpiComponent(data, state, opts) {
			var items = opts.items || [];
			if (!items.length) return __cel.div("text-gray-400", [__cel.text("No KPI items configured")]);
			var cards = [];
			for (var i = 0; i < items.length; i++) {
				var item = items[i];
				var val = data && data[item.key] !== undefined ? data[item.key] : 0;
				var formatted = val;
				if (item.format === "amount") formatted = __cfmt.amount(val);
				else if (item.format === "number") formatted = __cfmt.number(val);
				else formatted = String(val);
				var iconColor = item.color === "red" ? "text-red-400" : item.color === "green" ? "text-green-400" : item.color === "yellow" ? "text-yellow-400" : "text-blue-400";
				var valueColor = item.color === "red" ? "text-red-600 dark:text-red-400" : "text-gray-900 dark:text-gray-100";
				cards.push(__cel.div("bg-white dark:bg-gray-900 rounded-lg shadow p-4 flex items-center gap-3", [
					item.icon ? __cel.icon(item.icon, iconColor + " text-3xl") : null,
					__cel.div("flex flex-col", [
						__cel.div("text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider", [__cel.text(item.label || item.key)]),
						__cel.div("text-xl font-bold " + valueColor, [__cel.text(formatted)])
					])
				]));
			}
			return __cel.div("grid grid-cols-2 md:grid-cols-" + items.length + " gap-4", cards);
		}
		__cregister("kpi-bar", kpiComponent);
	})();
`)

var __cfileupload = Trim(`
	(function(){
		function fileUploadComponent(data, state, opts) {
			var zoneId = state.zoneId;
			var maxSize = opts.maxSize || 33554432;
			var accept = opts.accept || "";
			var multiple = opts.multiple !== false;
			var maxFiles = opts.maxFiles || 20;
			var files = state._files || [];
			var uploading = state._uploading || false;
			var progress = state._progress || 0;
			var results = state._results || [];

			function formatSize(bytes) {
				if (bytes < 1024) return bytes + " B";
				if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
				return (bytes / 1048576).toFixed(1) + " MB";
			}

			var inputId = zoneId + "_input";

			// Drop zone
			var dropZone = __cel("div", {
				class: "border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors " +
					(uploading ? "border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800/50 pointer-events-none" :
					"border-gray-300 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500 bg-white dark:bg-gray-900"),
				id: zoneId + "_dropzone"
			}, [
				__cel.icon("cloud_upload", "text-4xl text-gray-400 dark:text-gray-500 mb-2"),
				__cel.div("text-gray-600 dark:text-gray-400 mb-1", [__cel.text("Drag & drop files here")]),
				__cel.div("text-gray-400 dark:text-gray-500 text-sm", [__cel.text("or click to browse")]),
				accept ? __cel.div("text-gray-400 dark:text-gray-500 text-xs mt-2", [__cel.text("Accepted: " + accept)]) : null,
				__cel("input", {
					type: "file",
					id: inputId,
					class: "hidden",
					accept: accept || null,
					multiple: multiple ? "multiple" : null
				}, null, {
					change: { act: "raw", js: "(function(){ var inp=document.getElementById('" + inputId + "'); if(!inp||!inp.files) return; var cl=__clients['" + zoneId + "']; var st=cl.getState(); var cur=st._files||[]; var maxSz=" + maxSize + "; var maxF=" + maxFiles + "; var errs=[]; for(var i=0;i<inp.files.length;i++){ if(cur.length>=maxF){errs.push(inp.files[i].name+': max files reached'); continue;} if(inp.files[i].size>maxSz){errs.push(inp.files[i].name+': file too large'); continue;} cur.push({name:inp.files[i].name, size:inp.files[i].size, file:inp.files[i]}); } if(errs.length) __notify(errs.join(', '),'error'); cl.setState({_files:cur}); inp.value=''; })()" }
				})
			], {
				click: { act: "raw", js: "document.getElementById('" + inputId + "').click()" },
				dragover: { act: "raw", js: "(function(e){ e.preventDefault(); e.currentTarget.classList.remove('border-gray-300','dark:border-gray-600'); e.currentTarget.classList.add('border-blue-400','dark:border-blue-500','bg-blue-50','dark:bg-blue-900/20'); })(event)" },
				dragleave: { act: "raw", js: "(function(e){ e.currentTarget.classList.remove('border-blue-400','dark:border-blue-500','bg-blue-50','dark:bg-blue-900/20'); e.currentTarget.classList.add('border-gray-300','dark:border-gray-600'); })(event)" },
				drop: { act: "raw", js: "(function(e){ e.preventDefault(); e.currentTarget.classList.remove('border-blue-400','dark:border-blue-500','bg-blue-50','dark:bg-blue-900/20'); e.currentTarget.classList.add('border-gray-300','dark:border-gray-600'); var dt=e.dataTransfer; if(!dt||!dt.files) return; var cl=__clients['" + zoneId + "']; var st=cl.getState(); var cur=st._files||[]; var maxSz=" + maxSize + "; var maxF=" + maxFiles + "; var errs=[]; for(var i=0;i<dt.files.length;i++){ if(cur.length>=maxF){errs.push(dt.files[i].name+': max files reached'); continue;} if(dt.files[i].size>maxSz){errs.push(dt.files[i].name+': file too large'); continue;} cur.push({name:dt.files[i].name, size:dt.files[i].size, file:dt.files[i]}); } if(errs.length) __notify(errs.join(', '),'error'); cl.setState({_files:cur}); })(event)" }
			});

			// File list
			var fileList = null;
			if (files.length > 0) {
				var fileItems = files.map(function(f, idx) {
					return __cel.div("flex items-center justify-between py-2 px-3 bg-gray-50 dark:bg-gray-800 rounded mb-1", [
						__cel.div("flex items-center gap-2 min-w-0", [
							__cel.icon("description", "text-gray-400 text-lg"),
							__cel.div("min-w-0", [
								__cel.div("text-sm text-gray-700 dark:text-gray-300 truncate", [__cel.text(f.name)]),
								__cel.div("text-xs text-gray-400", [__cel.text(formatSize(f.size))])
							])
						]),
						uploading ? null : __cel("button", {
							type: "button",
							class: "text-gray-400 hover:text-red-500 ml-2"
						}, [__cel.icon("close", "text-sm")], {
							click: { act: "raw", js: "(function(e){ e.stopPropagation(); var cl=__clients['" + zoneId + "']; var st=cl.getState(); var ff=(st._files||[]).slice(); ff.splice(" + idx + ",1); cl.setState({_files:ff}); })(event)" }
						})
					]);
				});
				fileList = __cel.div("mt-3", fileItems);
			}

			// Progress bar
			var progressBar = null;
			if (uploading) {
				progressBar = __cel.div("mt-3", [
					__cel.div("flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1", [
						__cel.text("Uploading..."),
						__cel.text(Math.round(progress * 100) + "%")
					]),
					__cel.div("w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2", [
						__cel("div", {
							class: "bg-blue-600 h-2 rounded-full transition-all",
							style: "width:" + Math.round(progress * 100) + "%"
						}, [])
					])
				]);
			}

			// Upload button
			var uploadBtn = null;
			if (files.length > 0 && !uploading) {
				uploadBtn = __cel("button", {
					type: "button",
					class: "mt-3 bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 cursor-pointer"
				}, [
					__cel.icon("cloud_upload", "text-sm mr-1 align-middle"),
					__cel.text("Upload " + files.length + " file" + (files.length > 1 ? "s" : ""))
				], {
					click: { act: "raw", js: "(function(){ var cl=__clients['" + zoneId + "']; var st=cl.getState(); var ff=st._files||[]; if(!ff.length) return; cl.setState({_uploading:true,_progress:0}); var fd=new FormData(); for(var i=0;i<ff.length;i++) fd.append('files',ff[i].file,ff[i].name); __capi.upload('" + (opts.uploadUrl || "") + "',fd,function(p){ cl.setState({_progress:p}); }).then(function(res){ cl.setState({_uploading:false,_files:[],_progress:0,_results:res}); __notify('Upload complete','success'); }).catch(function(err){ cl.setState({_uploading:false,_progress:0}); __notify(err.message||'Upload failed','error'); }); })()" }
				});
			}

			// Results
			var resultsEl = null;
			if (results && ((Array.isArray(results) && results.length > 0) || (typeof results === "object" && Object.keys(results).length > 0))) {
				resultsEl = __cel.div("mt-3 p-3 bg-green-50 dark:bg-green-900/20 rounded text-sm text-green-700 dark:text-green-400", [
					__cel.icon("check_circle", "text-green-500 mr-1 align-middle"),
					__cel.text("Upload successful")
				]);
			}

			return __cel.div("", [dropZone, fileList, progressBar, uploadBtn, resultsEl]);
		}

		__cregister("file-upload", fileUploadComponent);
	})();
`)

var __confirm = Trim(`
	function __confirm(message, onConfirm, opts) {
		opts = opts || {};
		var title = opts.title || "Confirm";
		var confirmText = opts.confirmText || "Confirm";
		var cancelText = opts.cancelText || "Cancel";
		var variant = opts.variant || "default";

		return new Promise(function(resolve) {
			var overlay = document.createElement("div");
			overlay.style.cssText = "position:fixed;inset:0;z-index:10000;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.5);";
			var isDark = document.documentElement.classList.contains("dark");
			var dialog = document.createElement("div");
			dialog.style.cssText = "background:" + (isDark ? "#1f2937" : "#fff") + ";border-radius:12px;padding:24px;min-width:320px;max-width:420px;box-shadow:0 20px 60px rgba(0,0,0,0.2);color:" + (isDark ? "#e5e7eb" : "#111827") + ";";

			var h = document.createElement("h3");
			h.style.cssText = "margin:0 0 8px;font-size:16px;font-weight:600;";
			h.textContent = title;
			dialog.appendChild(h);

			var msg = document.createElement("p");
			msg.style.cssText = "margin:0 0 20px;font-size:14px;color:" + (isDark ? "#9ca3af" : "#6b7280") + ";line-height:1.5;";
			msg.textContent = message;
			dialog.appendChild(msg);

			var btnRow = document.createElement("div");
			btnRow.style.cssText = "display:flex;gap:8px;justify-content:flex-end;";

			var cancelBtn = document.createElement("button");
			cancelBtn.textContent = cancelText;
			cancelBtn.style.cssText = "padding:8px 16px;border-radius:8px;font-size:14px;cursor:pointer;border:1px solid " + (isDark ? "#4b5563" : "#d1d5db") + ";background:" + (isDark ? "#374151" : "#f9fafb") + ";color:" + (isDark ? "#e5e7eb" : "#374151") + ";";
			cancelBtn.onclick = function() { overlay.remove(); resolve(false); };

			var confirmBtn = document.createElement("button");
			confirmBtn.textContent = confirmText;
			var confirmBg = variant === "danger" ? "#dc2626" : "#3b82f6";
			confirmBtn.style.cssText = "padding:8px 16px;border-radius:8px;font-size:14px;cursor:pointer;border:none;background:" + confirmBg + ";color:#fff;font-weight:500;";
			confirmBtn.onclick = function() {
				overlay.remove();
				if (typeof onConfirm === "function") onConfirm();
				resolve(true);
			};

			btnRow.appendChild(cancelBtn);
			btnRow.appendChild(confirmBtn);
			dialog.appendChild(btnRow);
			overlay.appendChild(dialog);
			document.body.appendChild(overlay);

			overlay.addEventListener("click", function(e) {
				if (e.target === overlay) { overlay.remove(); resolve(false); }
			});
			document.addEventListener("keydown", function handler(e) {
				if (e.key === "Escape") { overlay.remove(); resolve(false); document.removeEventListener("keydown", handler); }
			});
			confirmBtn.focus();
		});
	}
`)

var __toast = Trim(`
	var __toast = {
		show: function(message, type, duration) {
			type = type || "info";
			var variant = type;
			if (type === "success") variant = "success";
			else if (type === "error") variant = "error";
			__notify(message, variant);
		},
		success: function(message) { __notify(message, "success"); },
		error: function(message) { __notify(message, "error"); },
		info: function(message) { __notify(message, "info"); }
	};
`)

var __cmodal = Trim(`
	var __cmodal = {
		open: function(content, opts) {
			opts = opts || {};
			var isDark = document.documentElement.classList.contains("dark");
			var overlay = document.createElement("div");
			overlay.id = "__cmodal_overlay__";
			overlay.style.cssText = "position:fixed;inset:0;z-index:10000;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.6);";

			var modal = document.createElement("div");
			modal.style.cssText = "background:" + (isDark ? "#1f2937" : "#fff") + ";border-radius:12px;max-width:" + (opts.maxWidth || "90vw") + ";max-height:85vh;overflow:auto;box-shadow:0 20px 60px rgba(0,0,0,0.3);position:relative;";

			// Close button
			var closeBtn = document.createElement("button");
			closeBtn.innerHTML = "&times;";
			closeBtn.style.cssText = "position:absolute;top:8px;right:12px;background:none;border:none;font-size:24px;cursor:pointer;color:" + (isDark ? "#9ca3af" : "#6b7280") + ";z-index:1;padding:4px 8px;line-height:1;";
			closeBtn.onclick = function() { __cmodal.close(); };
			modal.appendChild(closeBtn);

			if (typeof content === "string") {
				// Check if it looks like an image URL
				if (/\.(jpg|jpeg|png|gif|webp|svg|bmp)(\?|$)/i.test(content)) {
					var img = document.createElement("img");
					img.src = content;
					img.style.cssText = "display:block;max-width:100%;max-height:80vh;border-radius:12px;";
					img.alt = opts.alt || "";
					modal.appendChild(img);
				} else if (content.charAt(0) === "<") {
					var wrapper = document.createElement("div");
					wrapper.style.cssText = "padding:24px;padding-top:40px;";
					wrapper.innerHTML = content;
					modal.appendChild(wrapper);
				} else {
					var textWrapper = document.createElement("div");
					textWrapper.style.cssText = "padding:24px;padding-top:40px;white-space:pre-wrap;font-size:14px;color:" + (isDark ? "#e5e7eb" : "#374151") + ";max-height:70vh;overflow:auto;";
					textWrapper.textContent = content;
					modal.appendChild(textWrapper);
				}
			} else if (content instanceof HTMLElement) {
				modal.appendChild(content);
			}

			overlay.appendChild(modal);
			document.body.appendChild(overlay);

			overlay.addEventListener("click", function(e) {
				if (e.target === overlay) __cmodal.close();
			});
			document.addEventListener("keydown", function handler(e) {
				if (e.key === "Escape") { __cmodal.close(); document.removeEventListener("keydown", handler); }
			});
		},
		close: function() {
			var el = document.getElementById("__cmodal_overlay__");
			if (el) el.remove();
		},
		preview: function(url, opts) {
			__cmodal.open(url, opts);
		}
	};
`)

var __cbadge = Trim(`
	(function(){
		function badgeComponent(data, state, opts) {
			var count = 0;
			if (typeof data === "number") count = data;
			else if (data && data.count !== undefined) count = parseInt(data.count, 10) || 0;
			else if (data && data.value !== undefined) count = parseInt(data.value, 10) || 0;

			if (count <= 0 && opts.hideZero !== false) return null;

			var bg = opts.bg || "bg-red-500";
			var text = opts.text || "text-white";
			var size = opts.size || "min-w-[20px] h-5 text-xs";
			var display = count > 99 ? "99+" : String(count);

			return __cel("span", {
				class: bg + " " + text + " " + size + " inline-flex items-center justify-center rounded-full px-1.5 font-medium leading-none"
			}, [__cel.text(display)]);
		}

		__cregister("badge", badgeComponent);
	})();
`)

func MakeApp(defaultLanguage string) *App {
	contentID := Target()
	return &App{
		Lanugage:  defaultLanguage,
		ContentID: contentID,
		HTMLHead: []string{
			`<meta charset="UTF-8">`,
			`<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
			`<link rel="preconnect" href="https://fonts.googleapis.com">`,
			`<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>`,
			`<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">`,
			`<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">`,
			`<style>
                html {
                    scroll-behavior: smooth;
                }
                .invalid, select:invalid, textarea:invalid, input:invalid {
                    border-bottom-width: 2px;
                    border-bottom-color: red;
                    border-bottom-style: dashed;
                }
                /* Fix for Safari mobile date input width overflow */
                @media (max-width: 768px) {
                    input[type="date"] {
                        max-width: 100% !important;
                        width: 100% !important;
                        min-width: 0 !important;
                        box-sizing: border-box !important;
                        overflow: hidden !important;
                    }
                    
                    /* Ensure parent containers don't overflow */
                    input[type="date"]::-webkit-datetime-edit {
                        max-width: 100% !important;
                        overflow: hidden !important;
                    }
                }
            </style>`,
			`<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>`,
			// CSS overrides (after Tailwind so they take precedence)
			`<style id="gsui-overrides">
                /* Light mode background override */
                html.bg-gray-200 { background-color: rgba(235, 235, 235, 1); }
                /* Disabled buttons - ensure text is visible */
                .pointer-events-none.bg-gray-50 { color: #6b7280 !important; }
                /* Dark mode overrides — no !important so explicit dark: Tailwind variants can win */
                html.dark{ color-scheme: dark; }
                /* Global text color fallback */
                .dark body { color:#e5e7eb; }
                /* Backgrounds */
                html.dark.bg-white, html.dark.bg-gray-100, html.dark.bg-gray-200 { background-color:#111827; }
                .dark .bg-white, .dark .bg-gray-50, .dark .bg-gray-100, .dark .bg-gray-200 { background-color:#111827; }
                /* Text color overrides for common grays */
                .dark .text-black, .dark .text-gray-900, .dark .text-gray-800, .dark .text-gray-700, .dark .text-gray-600, .dark .text-gray-500 { color:#e5e7eb; }
                .dark .text-gray-400, .dark .text-gray-300 { color:#d1d5db; }
                /* Borders */
                .dark .border-gray-100, .dark .border-gray-200, .dark .border-gray-300 { border-color:#374151; }
                /* Inputs — keep !important since form elements have UA styles that are hard to override */
                .dark input, .dark select, .dark textarea { color:#e5e7eb !important; background-color:#1f2937 !important; }
                .dark input::placeholder, .dark textarea::placeholder { color:#9ca3af !important; }
                /* Hover helpers used in nav/examples */
                .dark .hover\:bg-gray-200:hover { background-color:#374151; }
            </style>`,
			Script(__stringify, __loader, __offline, __error, __notify, __e, __engine, __post, __submit, __load, __router, __theme, __cfmt, __debounce, __clipboard, __capi, __cel, __cregister, __cfilter, __cfilterbar, __cpagination, __caction, __ctable, __cchart, __ckpi, __cfileupload, __confirm, __toast, __cmodal, __cbadge, __clientScript, __ws),
		},
		HTMLBody: func(class string) string {
			if class == "" {
				class = "bg-gray-200"
			}

			return fmt.Sprintf(`
				<!DOCTYPE html>
				<html lang="__lang__" class="%s">
					<head>__head__</head>
					<body id="%s" class="relative">__body__</body>
				</html>
			`, class, contentID.ID)
		},
		DebugEnabled:    false,
		mux:             http.NewServeMux(),
		stored:          make(map[*Callable]string),
		sessions:        make(map[string]*sessRec),
		captchaSessions: make(map[string]*CaptchaSession),
		routes:          make(map[string]*Route),
	}
}

// isSecure returns true if the request is over TLS or forwarded as https
func isSecure(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}
	xf := r.Header.Get("X-Forwarded-Proto")
	if xf == "" {
		return false
	}
	// Use first value if comma-separated
	parts := strings.Split(xf, ",")
	if len(parts) > 0 && strings.TrimSpace(strings.ToLower(parts[0])) == "https" {
		return true
	}
	return false
}

// sweepSessions prunes sessions not seen for more than 60 seconds
func (app *App) sweepSessions() {
	cutoff := time.Now().Add(-60 * time.Second)
	app.sessMu.Lock()
	for k, rec := range app.sessions {
		if rec == nil || rec.lastSeen.Before(cutoff) {
			delete(app.sessions, k)
		}
	}
	app.sessMu.Unlock()
}

// StartSweeper launches a background goroutine to prune inactive sessions
func (app *App) StartSweeper() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			app.sweepSessions()
		}
	}()
}

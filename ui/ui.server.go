package ui

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	"golang.org/x/net/websocket"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Callable = func(*Context) string

var (
	eventPath      = "/"
	mu             sync.Mutex
	stored         = make(map[*Callable]string)
	reReplaceChars = regexp.MustCompile(`[./:-]`)
	reRemoveChars  = regexp.MustCompile(`[*()\[\]]`)
)

type BodyItem struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
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
	App       *App
	Request   *http.Request
	Response  http.ResponseWriter
	SessionID string
	append    []string
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
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '[' || r == ']' || r == '_') {
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

		val := reflect.ValueOf(item.Value)

		if structFieldValue.Type() != val.Type() {
			switch item.Type {
			case "Skeleton":
				val = reflect.ValueOf(Skeleton(item.Value))

			case "date":
				if len(item.Value) > 10 { // Basic length check for date format
					fmt.Printf("Warning: Date value too long at index %d\n", i)
					continue
				}
				t, err := time.Parse("2006-01-02", item.Value)
				if err != nil {
					fmt.Printf("Warning: Error parsing date at index %d: %v\n", i, err)
					continue
				}
				if structFieldValue.Type() == reflect.TypeOf(gorm.DeletedAt{}) {
					val = reflect.ValueOf(gorm.DeletedAt{Time: t, Valid: true})
				} else {
					val = reflect.ValueOf(t)
				}

			case "bool", "checkbox":
				if item.Value != "true" && item.Value != "false" {
					fmt.Printf("Warning: Invalid boolean value at index %d: %s\n", i, item.Value)
					continue
				}
				val = reflect.ValueOf(item.Value == "true")

			case "radio", "string":
				val = reflect.ValueOf(item.Value)

			case "time":
				if len(item.Value) > 5 { // Basic length check for time format HH:MM
					fmt.Printf("Warning: Time value too long at index %d\n", i)
					continue
				}
				t, err := time.Parse("15:04", item.Value)
				if err != nil {
					fmt.Printf("Warning: Error parsing time at index %d: %v\n", i, err)
					continue
				}
				val = reflect.ValueOf(t)

			case "Time":
				if len(item.Value) > 50 { // Extended length check for full timestamp
					fmt.Printf("Warning: Timestamp value too long at index %d\n", i)
					continue
				}
				t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", item.Value)
				if err != nil {
					fmt.Printf("Warning: Error parsing timestamp at index %d: %v\n", i, err)
					continue
				}
				val = reflect.ValueOf(t)

			case "uint", "int", "int64", "number", "decimal", "float64":
				// Validate numeric input with bounds checking and get parsed value
				parsedVal, err := validateNumericInput(item.Value, item.Type)
				if err != nil {
					fmt.Printf("Warning: Invalid numeric value at index %d: %v\n", i, err)
					continue
				}
				val = reflect.ValueOf(parsedVal)

			case "datetime-local":
				if len(item.Value) > 16 { // Basic length check for datetime-local format
					fmt.Printf("Warning: DateTime value too long at index %d\n", i)
					continue
				}
				t, err := time.Parse("2006-01-02T15:04", item.Value)
				if err != nil {
					fmt.Printf("Warning: Error parsing datetime-local at index %d: %v\n", i, err)
					continue
				}
				val = reflect.ValueOf(t)

			case "":
				continue

			case "Model": // gorm.Model
				continue

			default:
				fmt.Printf("Warning: Unknown field type at index %d: %s\n", i, item.Type)
				continue
			}
		}

		// Safe reflection assignment with error handling
		if val.IsValid() && val.Type().ConvertibleTo(structFieldValue.Type()) {
			structFieldValue.Set(val.Convert(structFieldValue.Type()))
		} else if val.IsValid() {
			fmt.Printf("Warning: Cannot convert %s to %s for field %s\n", val.Type(), structFieldValue.Type(), item.Name)
		}
	}

	return nil
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

func (ctx *Context) Post(as ActionType, swap Swap, action *Action) string {
	path, ok := stored[action.Method]

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
	return Attr{OnClick: Trim(fmt.Sprintf(`__load("%s")`, escapeJS(href)))}
}

func (ctx *Context) Reload() string {
	// return Normalize("<html><!DOCTYPE html><body><script>window.location.reload();</script></body></html>")
	return Normalize("<script>window.location.reload();</script>")
}

func (ctx *Context) Redirect(href string) string {
	// return Normalize(fmt.Sprintf("<html><!DOCTYPE html><body><script>window.location.href = '%s';</script></body></html>", href))
	return Normalize(fmt.Sprintf("<script>window.location.href = '%s';</script>", escapeJS(href)))
}

// Deferred fragments removed. The previous ctx.Defer(...) builder and helpers
// are no longer available.

func displayMessage(ctx *Context, message string, color string) {
	// Styled toast matching t-sui visuals
	script := Trim(fmt.Sprintf(`<script>(function(){
        var box=document.getElementById("__messages__");
        if(box==null){box=document.createElement("div");box.id="__messages__";box.style.position="fixed";box.style.top="0";box.style.right="0";box.style.padding="8px";box.style.zIndex="9999";box.style.pointerEvents="none";document.body.appendChild(box);} 
        var n=document.createElement("div");
        n.style.display="flex";n.style.alignItems="center";n.style.gap="10px";
        n.style.padding="12px 16px";n.style.margin="8px";n.style.borderRadius="12px";
        n.style.minHeight="44px";n.style.minWidth="340px";n.style.maxWidth="340px";
        n.style.boxShadow="0 6px 18px rgba(0,0,0,0.08)";n.style.border="1px solid";
        var C=%q;var isGreen=C.indexOf('green')>=0;var isRed=C.indexOf('red')>=0;
        var accent=isGreen?"#16a34a":(isRed?"#dc2626":"#4f46e5");
        if(isGreen){n.style.background="#dcfce7";n.style.color="#166534";n.style.borderColor="#bbf7d0";}
        else if(isRed){n.style.background="#fee2e2";n.style.color="#991b1b";n.style.borderColor="#fecaca";}
        else{n.style.background="#eef2ff";n.style.color="#3730a3";n.style.borderColor="#e0e7ff";}
        n.style.borderLeft="4px solid "+accent;
        var dot=document.createElement("span");dot.style.width="10px";dot.style.height="10px";dot.style.borderRadius="9999px";dot.style.background=accent;
        var t=document.createElement("span");t.textContent=%q; 
        n.appendChild(dot);n.appendChild(t);
        box.appendChild(n);
        setTimeout(function(){try{box.removeChild(n);}catch(_){}} ,5000);
    })();</script>`, color, message))
	ctx.append = append(ctx.append, script)
}

// displayError renders an error toast similar to displayMessage's red variant
// and includes a Reload button that refreshes the application.
func displayError(ctx *Context, message string) {
	// Fixed red styling with reload button
	script := Trim(fmt.Sprintf(`<script>(function(){
        var box=document.getElementById("__messages__");
        if(box==null){box=document.createElement("div");box.id="__messages__";box.style.position="fixed";box.style.top="0";box.style.right="0";box.style.padding="8px";box.style.zIndex="9999";box.style.pointerEvents="none";document.body.appendChild(box);} 
        var n=document.createElement("div");
        n.style.display='flex';n.style.alignItems='center';n.style.gap='10px';
        n.style.padding='12px 16px';n.style.margin='8px';n.style.borderRadius='12px';
        n.style.minHeight='44px';n.style.minWidth='340px';n.style.maxWidth='340px';
        n.style.background='#fee2e2';n.style.color='#991b1b';n.style.border='1px solid #fecaca';
        n.style.borderLeft='4px solid #dc2626';n.style.boxShadow='0 6px 18px rgba(0,0,0,0.08)';
        n.style.fontWeight='600';n.style.pointerEvents='auto';
        var dot=document.createElement('span');dot.style.width='10px';dot.style.height='10px';dot.style.borderRadius='9999px';dot.style.background='#dc2626';
        var t=document.createElement('span');t.textContent=%q;
        var btn=document.createElement('button');btn.textContent='Reload';btn.style.background='#991b1b';btn.style.color='#fff';btn.style.border='none';btn.style.padding='6px 10px';btn.style.borderRadius='8px';btn.style.cursor='pointer';btn.style.fontWeight='700';btn.onclick=function(){ try { window.location.reload(); } catch(_){} };
        n.appendChild(dot);n.appendChild(t);n.appendChild(btn);
        box.appendChild(n);
        setTimeout(function(){ try { if(n && n.parentNode) { n.parentNode.removeChild(n); } } catch(_){} }, 88000);
    })();</script>`, message))
	ctx.append = append(ctx.append, script)
}

func (ctx *Context) Success(message string) {
	displayMessage(ctx, message, "bg-green-700 text-white")
}

func (ctx *Context) Error(message string) {
	displayMessage(ctx, message, "bg-red-700 text-white")
}

// ErrorReload shows an error toast with a Reload button.
func (ctx *Context) ErrorReload(message string) { displayError(ctx, message) }

func (ctx *Context) Info(message string) {
	displayMessage(ctx, message, "bg-blue-700 text-white")
}

// Title updates the page title dynamically
func (ctx *Context) Title(title string) {
	script := fmt.Sprintf(`<script>document.title=%q;</script>`, title)
	ctx.append = append(ctx.append, script)
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

	ctx.append = append(ctx.append,
		Trim(fmt.Sprintf(`<script>
            (function () {
                const byteCharacters = atob("%s");
                const byteNumbers = new Array(byteCharacters.length);
                for (let i = 0; i < byteCharacters.length; i++) {
                    byteNumbers[i] = byteCharacters.charCodeAt(i);
                }
                const byteArray = new Uint8Array(byteNumbers);
                const blob = new Blob([byteArray], { type: "%s" });
                const url = URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = "%s";
                a.click();
                URL.revokeObjectURL(url);
            })();
        </script>`, fileBase64, contentType, name)),
	)

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

func cacheControlMiddleware(next http.Handler, maxAge time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Cache-Control header
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds())))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

type App struct {
	Lanugage     string
	HTMLBody     func(string) string
	HTMLHead     []string
	DebugEnabled bool
	SmoothNav    bool
	sessMu       sync.Mutex
	sessions     map[string]*sessRec
	wsMu         sync.RWMutex
	wsClients    map[*websocket.Conn]*wsState
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

	_, ok := stored[method]
	if ok {
		panic("Method already registered: " + funcName)
	}

	for _, value := range stored {
		if value == path {
			panic("Path already registered: " + path)
		}
	}

	mu.Lock()
	stored[method] = path
	mu.Unlock()

	// fmt.Println("Registering: ", httpMethod, path, " -> ", funcName)

	return path
}

// Page registers a route with optional middleware support.
// Usage: Page("/", middleware1, middleware2, component)
// All middleware functions are called in order before the component.
// If any middleware returns a non-empty string, that response is used and execution stops.
// All middleware and components must have the same signature: func(ctx *Context) string
func (app *App) Page(path string, component ...Callable) **Callable {
	if len(component) == 0 {
		panic("Page requires at least one component")
	}

	for key, value := range stored {
		if value == path {
			return &key
		}
	}

	// If only one component provided, use it directly (backward compatible)
	if len(component) == 1 {
		fn := component[0]
		found := &fn
		mu.Lock()
		stored[found] = path
		mu.Unlock()
		return &found
	}

	// Multiple components: last one is the handler, others are middleware
	// Create a wrapper function that chains middleware before the final component
	handler := component[len(component)-1]
	middleware := component[:len(component)-1]

	// Create a chained wrapper function
	wrapper := func(ctx *Context) string {
		// Call all middleware in order
		for _, mw := range middleware {
			result := mw(ctx)
			// If middleware returns non-empty string, use it as response (early return)
			if result != "" {
				return result
			}
		}
		// All middleware passed, call the final component
		return handler(ctx)
	}

	found := &wrapper

	mu.Lock()
	stored[found] = path
	mu.Unlock()

	return &found
}

// Debug enables or disables server debug logging.
// When enabled, debug logs are printed with the "gsui:" prefix.
func (app *App) Debug(enable bool) {
	app.DebugEnabled = enable
}

// SmoothNavigation enables or disables automatic link interception for smooth navigation.
// When enabled, all internal links will use background loading with delayed loader instead of full page reloads.
func (app *App) SmoothNavigation(enable bool) {
	app.SmoothNav = enable
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

	for key, value := range stored {
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

	for key, value := range stored {
		if value == uid {
			return &key
		}
	}

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
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
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
	http.Handle("/"+path, mimeTypeMiddleware(cacheControlMiddleware(handler, maxAge)))
}

// Favicon serves a favicon file from the embedded filesystem at /favicon.ico.
// The path parameter should be the path to the favicon file in the embed.FS
// (e.g., "assets/favicon.ico", "assets/favicon.svg").
// Defaults to "favicon.ico" if not provided.
func (app *App) Favicon(assets embed.FS, path string, maxAge time.Duration) {
	if path == "" {
		path = "favicon.ico"
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
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

func (app *App) Listen(port string) {
	log.Println("Listening on http://0.0.0.0" + port)

	// Start session sweeper in background
	app.StartSweeper()

	// Init WebSocket endpoint for patches
	app.initWS()

	// Wrap the main handler with security headers middleware
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains("GET POST", r.Method) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		value := r.URL.Path

		if strings.Contains(strings.Join(r.Header["Upgrade"], " "), "websocket") {
			// Let explicit WS handlers handle upgrades
			http.NotFound(w, r)
			return
		}

		for found, path := range stored {
			if value == path {
				ctx := makeContext(app, r, w)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")

				// Recover from panics inside handler calls to avoid broken fetches
				defer func() {
					if rec := recover(); rec != nil {
						log.Println("handler panic recovered:", rec)
						// Serve a minimal error page that auto-reloads once the dev WS reconnects
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(devErrorPage()))
					}
				}()

				// Normal call
				app.debugf("route %s -> %s", r.Method, path)
				w.Write([]byte((*found)(ctx)))
				if len(ctx.append) > 0 {
					w.Write([]byte(strings.Join(ctx.append, "")))
				}

				return
			}
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

	// Apply security headers middleware
	http.Handle("/", securityHeadersMiddleware(mainHandler))

	if err := http.ListenAndServe(port, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println("Error:", err)
	}
}

// initWS registers the WebSocket endpoint for server-initiated patches.
func (app *App) initWS() {
	app.wsMu.Lock()
	if app.wsClients == nil {
		app.wsClients = make(map[*websocket.Conn]*wsState)
	}
	app.wsMu.Unlock()

	http.Handle("/__ws", websocket.Handler(func(ws *websocket.Conn) {
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

		// Receive loop: handle ping/pong from client and invalid target notices
		for {
			var s string
			if err := websocket.Message.Receive(ws, &s); err != nil {
				close(done)
				return
			}
			var obj map[string]any
			if err := json.Unmarshal([]byte(s), &obj); err == nil {
				if t, _ := obj["type"].(string); t != "" {
					if t == "ping" {
						_ = websocket.Message.Send(ws, `{"type":"pong"}`)
					} else if t == "pong" {
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
					} else if t == "invalid" {
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
					}
				}
			}
		}
	}))
}

// sendPatch broadcasts a patch message to all connected WS clients.
func (app *App) sendPatch(id string, swap Swap, html string) {
	msg := map[string]string{
		"type": "patch",
		"id":   id,
		"swap": string(swap),
		"html": Trim(html),
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
	ctx.App.sendPatch(ts.ID, ts.Swap, html)
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

	// Conditionally add smooth navigation script if enabled
	if app.SmoothNav {
		head = append(head, Script(__smoothnav))
	}

	html := app.HTMLBody(class)
	html = strings.ReplaceAll(html, "__lang__", app.Lanugage)
	html = strings.ReplaceAll(html, "__head__", strings.Join(head, " "))
	html = strings.ReplaceAll(html, "__body__", strings.Join(body, " "))

	return Trim(html)
}

// devErrorPage returns a minimal standalone HTML page displayed on handler panics in dev.
// It tries to reconnect to the app WS at /__ws and reloads the page when the socket opens.
func devErrorPage() string {
	return Trim(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Something went wrong…</title>
  <style>
    html,body{height:100%}
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
            var ws=new WebSocket(p+location.host+'/__ws');
            ws.onopen=function(){ try{ location.reload(); } catch(_){} };
            ws.onclose=function(){ setTimeout(connect, 1000); };
            ws.onerror=function(){ try{ ws.close(); } catch(_){} };
          }
          connect();
        } catch(_){ /* noop */ }
      })();
    </script>
  </body>
</html>`)
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

		fetch(path, {method: "POST", body: JSON.stringify(body)})
			.then(function(resp){ if(!resp.ok){ throw new Error('HTTP '+resp.status); } return resp.text(); })
			.then(function (html) {
				const parser = new DOMParser();
				const doc = parser.parseFromString(html, 'text/html');
				const scripts = [...doc.body.querySelectorAll('script'), ...doc.head.querySelectorAll('script')];

				for (let i = 0; i < scripts.length; i++) {
					const newScript = document.createElement('script');
					newScript.textContent = scripts[i].textContent;
					document.body.appendChild(newScript);
				}

				const el = document.getElementById(target_id);
				if (el != null) {
					if (swap === "inline") {
						el.innerHTML = html;
					} else if (swap === "outline") {
						el.outerHTML = html;
					} else if (swap === "append") {
						el.insertAdjacentHTML('beforeend', html);
					} else if (swap === "prepend") {
						el.insertAdjacentHTML('afterbegin', html);
					}
				}
			})
			.catch(function(_){ try { __error('Something went wrong ...'); } catch(__){} })
			.finally(function(){ try { L.stop(); } catch(_){} });
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

		let found = Array.from(document.querySelectorAll('[form=' + id + '][name]'));

		if (found.length === 0) {
			found = Array.from(form.querySelectorAll('[name]'));
		};

		found.forEach((item) => {
			const name = item.getAttribute("name");
			const type = item.getAttribute("type");
			let value = item.value;
			
			if (type === 'checkbox') {
				value = String(item.checked)
			}

			if(name != null) {
				body = body.filter(element => element.name !== name);
				body.push({ name, type, value });
			}
		});

		var L = (function(){ try { return __loader.start(); } catch(_) { return { stop: function(){} }; } })();

		fetch(path, {method: "POST", body: JSON.stringify(body)})
			.then(function(resp){ if(!resp.ok){ throw new Error('HTTP '+resp.status); } return resp.text(); })
			.then(function (html) {
				const parser = new DOMParser();
				const doc = parser.parseFromString(html, 'text/html');
				const scripts = [...doc.body.querySelectorAll('script'), ...doc.head.querySelectorAll('script')];

				for (let i = 0; i < scripts.length; i++) {
					const newScript = document.createElement('script');
					newScript.textContent = scripts[i].textContent;
					document.body.appendChild(newScript);
				}

				const el = document.getElementById(target_id);
				if (el != null) {
					if (swap === "inline") {
						el.innerHTML = html;
					} else if (swap === "outline") {
						el.outerHTML = html;
					} else if (swap === "append") {
						el.insertAdjacentHTML('beforeend', html);
					} else if (swap === "prepend") {
						el.insertAdjacentHTML('afterbegin', html);
					}
				}
			})
            .catch(function(_){ try { __error('Something went wrong ...'); } catch(__){} })
            .finally(function(){ try { L.stop(); } catch(_){} });
    }
`)

// __smoothnav: automatically intercepts clicks on internal links for smooth navigation
var __smoothnav = Trim(`
    (function(){
        try {
            if (window.__gsuiSmoothNavInit) { return; }
            window.__gsuiSmoothNavInit = true;
            
            function isInternalLink(href) {
                if (!href) return false;
                // Skip hash-only links
                if (href.startsWith('#')) return false;
                // Skip javascript: links
                if (href.startsWith('javascript:')) return false;
                // Skip data: and mailto: links
                if (href.startsWith('data:') || href.startsWith('mailto:')) return false;
                // Check if external (starts with http/https but not same origin)
                if (href.startsWith('http://') || href.startsWith('https://')) {
                    try {
                        var linkUrl = new URL(href, window.location.href);
                        return linkUrl.origin === window.location.origin;
                    } catch(_) {
                        return false;
                    }
                }
                // Relative paths are internal
                return true;
            }
            
            document.addEventListener('click', function(e) {
                var link = e.target.closest('a');
                if (!link) return;
                
                var href = link.getAttribute('href');
                if (!href) return;
                
                // Skip links with target attribute (e.g., _blank)
                if (link.target && link.target !== '_self') return;
                
                // Skip links with download attribute
                if (link.download) return;
                
                // Skip if link already has onclick handler (avoid double-handling with ctx.Load())
                var onclickAttr = link.getAttribute('onclick');
                if (onclickAttr && onclickAttr.trim().length > 0) return;
                
                // Only intercept internal links
                if (!isInternalLink(href)) return;
                
                // Prevent default navigation
                e.preventDefault();
                
                // Use __load for smooth navigation
                try {
                    __load(href);
                } catch(_) {
                    // Fallback to normal navigation if __load fails
                    window.location.href = href;
                }
            }, true);
        } catch(_) {}
    })();
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

// __ws: minimal WebSocket client for receiving server-initiated patches
var __ws = Trim(`
    (function(){
        try {
            if (window.__gsuiWSInit) { return; }
            window.__gsuiWSInit = true;
            var appPing = 0;
            // Track whether the WS has ever been closed; used to trigger a full reload once it reconnects
            try { if (!(window).__gsuiHadClose) { (window).__gsuiHadClose = false; } } catch(_){ }
            // Track targets we've actually seen in the DOM at least once
            // to avoid reporting them as invalid during initial load.
            try { if (!(window).__gsuiKnownTargets) { (window).__gsuiKnownTargets = Object.create(null); } } catch(_){}
            function markSeen(id){ try { (window).__gsuiKnownTargets[id] = true; } catch(_){ } }
            function wasSeen(id){ try { return !!((window).__gsuiKnownTargets && (window).__gsuiKnownTargets[id]); } catch(_){ return false; } }
            function handlePatch(msg){
                try {
                    var html = String(msg.html||'');
                    // execute inline scripts inside the html by extracting and re-inserting
                    try {
                        var tpl = document.createElement('template'); tpl.innerHTML = html;
                        var scripts = tpl.content.querySelectorAll('script');
                        for (var i=0;i<scripts.length;i++){
                            var s=document.createElement('script'); s.textContent=scripts[i].textContent; document.body.appendChild(s);
                        }
                    } catch(_){ }
                    var id = String(msg.id||'');
                    var el = document.getElementById(id);
                    if (!el) {
                        // Only report invalid once the target was present before.
                        // This prevents calling clear() prematurely during initial render.
                        if (wasSeen(id)) {
                            try {
                                var ws2 = (window).__gsuiWS;
                                if (ws2 && ws2.readyState === 1) {
                                    ws2.send(JSON.stringify({ type: 'invalid', id: id }));
                                }
                            } catch(_){ }
                        }
                        return;
                    }
                    // Mark target as seen when it's present in DOM
                    try { markSeen(id); } catch(_){ }
                    if (msg.swap==='inline') { el.innerHTML = html; }
                    else if (msg.swap==='outline') { el.outerHTML = html; }
                    else if (msg.swap==='append') { el.insertAdjacentHTML('beforeend', html); }
                    else if (msg.swap==='prepend') { el.insertAdjacentHTML('afterbegin', html); }
                } catch(_){ }
            }
            function connect(){
                var p=(location.protocol==='https:')?'wss://':'ws://';
                var ws = new WebSocket(p+location.host+'/__ws');
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
`)

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

var ContentID = Target()

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

func MakeApp(defaultLanguage string) *App {
	return &App{
		Lanugage: defaultLanguage,
		HTMLHead: []string{
			`<meta charset="UTF-8">`,
			`<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
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
			`<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/tailwindcss/2.2.19/tailwind.min.css" integrity="sha512-wnea99uKIC3TJF7v4eKk4Y+lMz2Mklv18+r4na2Gn1abDRPPOeef95xTzdwGD9e6zXJBteMIhZ1+68QC5byJZw==" crossorigin="anonymous" referrerpolicy="no-referrer" />`,
			// Dark mode CSS overrides (after Tailwind so they take precedence)
			`<style id="gsui-dark-overrides">
                html.dark{ color-scheme: dark; }
                /* Global text color fallback */
                .dark body { color:#e5e7eb; }
                /* Backgrounds */
                html.dark.bg-white, html.dark.bg-gray-100, html.dark.bg-gray-200 { background-color:#111827 !important; }
                .dark .bg-white, .dark .bg-gray-50, .dark .bg-gray-100, .dark .bg-gray-200 { background-color:#111827 !important; }
                /* Text color overrides for common grays */
                .dark .text-black, .dark .text-gray-900, .dark .text-gray-800, .dark .text-gray-700, .dark .text-gray-600, .dark .text-gray-500 { color:#e5e7eb !important; }
                .dark .text-gray-400, .dark .text-gray-300 { color:#d1d5db !important; }
                /* Borders */
                .dark .border-gray-100, .dark .border-gray-200, .dark .border-gray-300 { border-color:#374151 !important; }
                /* Inputs */
                .dark input, .dark select, .dark textarea { color:#e5e7eb !important; background-color:#1f2937 !important; }
                .dark input::placeholder, .dark textarea::placeholder { color:#9ca3af !important; }
                /* Hover helpers used in nav/examples */
                .dark .hover\:bg-gray-200:hover { background-color:#374151 !important; }
            </style>`,
			Script(__stringify, __loader, __offline, __error, __post, __submit, __load, __theme, __ws),
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
			`, class, ContentID.ID)
		},
		DebugEnabled: false,
		sessions:     make(map[string]*sessRec),
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

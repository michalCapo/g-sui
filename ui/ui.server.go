package ui

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
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
		fmt.Println(err)
		return
	}

	temp := &TSession{
		SessionID: session.SessionID,
		Name:      session.Name,
		Data:      data,
	}

	session.DB.Where("session_id = ? and name = ?", session.SessionID, session.Name).Save(temp)
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

func (ctx *Context) Body(output any) error {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}

	var data []BodyItem
	if len(body) > 0 {
		err = json.Unmarshal(body, &data)
		if err != nil {
			return err
		}
	}

	for _, item := range data {
		structFieldValue, err := PathValue(output, item.Name)
		if err != nil {
			fmt.Println("Error getting field", item.Name, err)
			continue
		}

		if !structFieldValue.CanSet() {
			continue
		}

		val := reflect.ValueOf(item.Value)

		if structFieldValue.Type() != val.Type() {
			switch item.Type {
			case "date":
				t, err := time.Parse("2006-01-02", item.Value)
				if err != nil {
					fmt.Println("Error parsing date", err)
					continue
				}
				if structFieldValue.Type() == reflect.TypeOf(gorm.DeletedAt{}) {
					val = reflect.ValueOf(gorm.DeletedAt{Time: t, Valid: true})
				} else {
					val = reflect.ValueOf(t)
				}

			case "bool", "checkbox":
				val = reflect.ValueOf(item.Value == "true")

			case "radio", "string":
				val = reflect.ValueOf(item.Value)

			case "time":
				t, err := time.Parse("15:04", item.Value)
				if err != nil {
					fmt.Println("Error parsing time", err)
					continue
				}
				val = reflect.ValueOf(t)

			case "Time":
				t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", item.Value)
				if err != nil {
					fmt.Println("Error parsing time", err)
				}
				val = reflect.ValueOf(t)

			case "uint":
				cleanedValue := strings.ReplaceAll(item.Value, "_", "")
				n, err := strconv.ParseUint(cleanedValue, 10, 64)
				if err != nil {
					fmt.Println("Error parsing number", err)
					continue
				}
				val = reflect.ValueOf(uint(n))

			case "int":
				cleanedValue := strings.ReplaceAll(item.Value, "_", "")
				n, err := strconv.ParseInt(cleanedValue, 10, 64)
				if err != nil {
					fmt.Println("Error parsing number", err)
					continue
				}
				val = reflect.ValueOf(int(n))

			case "int64":
				cleanedValue := strings.ReplaceAll(item.Value, "_", "")
				n, err := strconv.ParseInt(cleanedValue, 10, 64)
				if err != nil {
					fmt.Println("Error parsing number", err)
					continue
				}
				val = reflect.ValueOf(int64(n))

			case "number":
				cleanedValue := strings.ReplaceAll(item.Value, "_", "")
				n, err := strconv.Atoi(cleanedValue)
				if err != nil {
					fmt.Println("Error parsing number", err)
					continue
				}
				val = reflect.ValueOf(n)

			case "decimal":
				cleanedValue := strings.ReplaceAll(item.Value, "_", "")
				f, err := strconv.ParseFloat(cleanedValue, 64)
				if err != nil {
					fmt.Println("Error parsing decimal", err)
					continue
				}
				val = reflect.ValueOf(f)

			case "float64":
				cleanedValue := strings.ReplaceAll(item.Value, "_", "")
				f, err := strconv.ParseFloat(cleanedValue, 64)
				if err != nil {
					fmt.Println("Error parsing float64", err)
					continue
				}
				val = reflect.ValueOf(f)

			case "datetime-local":
				t, err := time.Parse("2006-01-02T15:04", item.Value)
				if err != nil {
					fmt.Println("Error parsing datetime-local", err)
					continue
				}
				val = reflect.ValueOf(t)

			// case "text":
			// 	val = reflect.ValueOf(item.Value)

			case "":
				continue

			case "Model": // gorm.Model
				continue

			default:
				fmt.Println("Skipping (name;type;value):", item.Name, ";", item.Type, ";", item.Value)
				continue
			}
		}

		// fmt.Println("Setting", item.Name, "to", item.Value)
		structFieldValue.Set(val)
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
		return Normalize(fmt.Sprintf(`__submit(event, "%s", "%s", "%s", %s) `, swap, action.Target.ID, path, values))
	}

	return Normalize(fmt.Sprintf(`__post(event, "%s", "%s", "%s", %s) `, swap, action.Target.ID, path, values))
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
	return Attr{OnClick: Normalize(fmt.Sprintf(`__load("%s")`, href))}
}

func (ctx *Context) Reload() string {
	// return Normalize("<html><!DOCTYPE html><body><script>window.location.reload();</script></body></html>")
	return Normalize("<script>window.location.reload();</script>")
}

func (ctx *Context) Redirect(href string) string {
	// return Normalize(fmt.Sprintf("<html><!DOCTYPE html><body><script>window.location.href = '%s';</script></body></html>", href))
	return Normalize(fmt.Sprintf("<script>window.location.href = '%s';</script>", href))
}

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

func RandomString(n ...int) string {
	if len(n) == 0 {
		return RandomString(20)
	}

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n[0])
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
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
    Lanugage string
    HTMLBody func(string) string
    HTMLHead []string
    DebugEnabled bool
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

func (app *App) Page(path string, component Callable) **Callable {
	for key, value := range stored {
		if value == path {
			return &key
		}
	}

	found := &component

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

func (app *App) Assets(assets embed.FS, path string, maxAge time.Duration) {
	path = strings.TrimPrefix(path, "/")
	http.Handle("/"+path, cacheControlMiddleware(http.FileServer(http.FS(assets)), maxAge))
}

func (app *App) Favicon(assets embed.FS, path string, maxAge time.Duration) {
	path = strings.TrimPrefix(path, "/")
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		file, err := assets.ReadFile(path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds())))
		w.Write(file)
	})
}

func makeContext(app *App, r *http.Request, w http.ResponseWriter) *Context {
	var sessionID string

	cookie, err := r.Cookie("session_id")
	if err != nil {
		sessionID = RandomString(30)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		sessionID = cookie.Value
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

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains("GET POST", r.Method) {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        value := r.URL.Path

        if strings.Contains(strings.Join(r.Header["Upgrade"], " "), "websocket") {
            app.debugf("upgrade request (ignored): %s %s", r.Method, value)
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
                        // Enqueue an error toast to the client
                        displayError(ctx, "Something went wrong ...")
                        // Send minimal body with any queued scripts
                        if len(ctx.append) > 0 {
                            w.Write([]byte(strings.Join(ctx.append, "")))
                        }
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

	if err := http.ListenAndServe(port, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println("Error:", err)
	}
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

	go watchAndRestart(mainDir)
}

// watchAndRestart watches the provided directory recursively for file changes
// and rebuilds + execs the binary in-place when a change is detected.
func watchAndRestart(root string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("[autorestart] watcher error:", err)
		return
	}
	defer watcher.Close()

	// Add directories recursively
	addDirs := func() error {
		return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // ignore traversal errors
			}
			if d.IsDir() {
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
					// Add new directory and its children
					_ = filepath.WalkDir(ev.Name, func(p string, d os.DirEntry, err error) error {
						if err != nil {
							return nil
						}
						if d.IsDir() {
							if shouldSkipDir(d.Name()) {
								return filepath.SkipDir
							}
							_ = watcher.Add(p)
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

// rebuildAndExec builds the main package in root and re-execs into the new binary.
func rebuildAndExec(root string) error {
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("g-sui-%d", time.Now().UnixNano()))
	cmd := exec.Command("go", "build", "-o", tmp)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Replace current process with the new binary
	args := append([]string{tmp}, os.Args[1:]...)
	env := os.Environ()

	// Best effort: exec on Unix, spawn+exit on Windows
	if runtime.GOOS == "windows" {
		c := exec.Command(tmp, os.Args[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Start(); err != nil {
			return err
		}
		// Exit current process to let the new one take over
		os.Exit(0)
		return nil
	}

	return syscall.Exec(tmp, args, env)
}

func (app *App) Autoreload(enable bool) {
	if enable {
		app.HTMLHead = append(app.HTMLHead, `
        <script>
            (function(){
                if (window.__srui_live__) return;
                window.__srui_live__ = true;
                const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
                const socket = new WebSocket(protocol + window.location.host + '/live');
                socket.addEventListener('close', function () {
                    document.body.innerHTML += '<div class="fixed inset-0 z-40 opacity-75 bg-gray-800"></div>';
                    document.body.innerHTML += '<div class="fixed z-50 top-6 left-6 p-6 text-white bg-red-700 rounded border border-gray-500 uppercase font-bold">Offline</div>';
                    setInterval(() => {
                        fetch('/').then(() => window.location.reload()).catch(() => {});
                    }, 2000);
                });
            })();
        </script>
    `)

		http.Handle("/live", websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			for {
				time.Sleep(10 * time.Second)
				ws.Write([]byte("ok"))
			}
		}))
	}
}

func (app *App) Description(description string) {
	app.HTMLHead = append(app.HTMLHead, `<meta name="description" content="`+description+`">`)
}

func (app *App) HTML(title string, class string, body ...string) string {
	head := []string{
		`<title>` + title + `</title>`,
	}

	head = append(head, app.HTMLHead...)

	html := app.HTMLBody(class)
	html = strings.ReplaceAll(html, "__lang__", app.Lanugage)
	html = strings.ReplaceAll(html, "__head__", strings.Join(head, " "))
	html = strings.ReplaceAll(html, "__body__", strings.Join(body, " "))

	return Trim(html)
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

		let loader;
		let loading = setTimeout(() => {
			loader = document.createElement("div");
			loader.classList = "fixed inset-0 flex gap-4 items-center justify-center z-50 bg-white opacity-75 font-bold text-3xl";
			loader.innerHTML = "Loading ...";
			document.body.appendChild(loader);
		}, 100);

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
			.finally(function(){ clearTimeout(loading); if(loader){ try { document.body.removeChild(loader); } catch(_){} } });
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

		let loader;
		let loading = setTimeout(() => {
			loader = document.createElement("div");
			loader.classList = "fixed inset-0 flex gap-4 items-center justify-center z-50 bg-white opacity-75 font-bold text-3xl";
			loader.innerHTML = "Loading ...";
			document.body.appendChild(loader);
		}, 100);

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
            .finally(function(){ clearTimeout(loading); if(loader){ try { document.body.removeChild(loader); } catch(_){} } });
    }
`)

var __load = Trim(`
    function __load(href) {
		event.preventDefault(); 

		let loader;
		let loading = setTimeout(() => {
			loader = document.createElement("div");
			loader.classList = "fixed inset-0 flex gap-4 items-center justify-center z-50 bg-white opacity-75 font-bold text-3xl";
			loader.innerHTML = "Loading ...";
			document.body.appendChild(loader);
		}, 100);

		fetch(href, {method: "GET"})
			.then(function(resp){ if(!resp.ok){ throw new Error('HTTP '+resp.status); } return resp.text(); })
			.then(function (html) {
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
			.catch(function(_){ try { __error('Something went wrong ...'); } catch(__){} })
			.finally(function(){ clearTimeout(loading); if(loader){ try { document.body.removeChild(loader); } catch(_){} } });
    }
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
			Script(__stringify, __error, __post, __submit, __load),
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
    }
}

package ui

import (
	"encoding/json"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func pwaGet(t *testing.T, app *App, url string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", url, nil))
	if rr.Code != 200 {
		t.Fatalf("GET %s: status %d", url, rr.Code)
	}
	return rr
}

// The manifest is generated from App.PWA, with defaults for empty fields.
func TestPWAManifest(t *testing.T) {
	app := NewApp()
	app.Lang = "sk"
	app.PWA(PWA{
		Name:       "Mail",
		Icons:      []PWAIcon{{Src: "/assets/icon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "maskable"}},
		Shortcuts:  []PWAShortcut{{Name: "Compose", URL: "/compose"}},
		Categories: []string{"email"},
	})
	rr := pwaGet(t, app, "/manifest.webmanifest")
	if rr.Header().Get("Content-Type") != "application/manifest+json" || rr.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("headers %v", rr.Header())
	}
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{
		"name": "Mail", "short_name": "Mail", "id": "/", "start_url": "/", "scope": "/", "lang": "sk",
		"display": "standalone", "theme_color": "#ffffff", "background_color": "#ffffff",
	} {
		if m[key] != want {
			t.Errorf("%s = %v, want %v", key, m[key], want)
		}
	}
	if _, ok := m["description"]; ok {
		t.Error("empty description must be left out")
	}
	icons, _ := json.Marshal(m["icons"])
	shortcuts, _ := json.Marshal(m["shortcuts"])
	if string(icons) != `[{"purpose":"maskable","sizes":"512x512","src":"/assets/icon-512.png","type":"image/png"}]` ||
		string(shortcuts) != `[{"name":"Compose","url":"/compose"}]` {
		t.Fatalf("icons %s, shortcuts %s", icons, shortcuts)
	}
	// The browser revalidates it and gets 304 while it is unchanged.
	req := httptest.NewRequest("GET", "/manifest.webmanifest", nil)
	req.Header.Set("If-None-Match", rr.Header().Get("ETag"))
	rr = httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	if rr.Code != 304 {
		t.Fatalf("revalidation status %d", rr.Code)
	}
}

// StartURL defaults to Scope and ID to StartURL.
func TestPWAManifestURLDefaults(t *testing.T) {
	app := NewApp()
	app.Title = "Trenum"
	app.PWA(PWA{Scope: "/app/"})
	var m map[string]any
	if err := json.Unmarshal(pwaGet(t, app, "/manifest.webmanifest").Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "Trenum" || m["scope"] != "/app/" || m["start_url"] != "/app/" || m["id"] != "/app/" {
		t.Fatalf("manifest %v", m)
	}
}

// The worker serves the offline page only for failed page loads and ends
// with the app's own code.
func TestPWAServiceWorker(t *testing.T) {
	app := NewApp()
	app.PWA(PWA{Name: "Mail", Worker: "self.addEventListener('push', function () {});"})
	rr := pwaGet(t, app, "/sw.js")
	h := rr.Header()
	if h.Get("Content-Type") != "application/javascript; charset=utf-8" || h.Get("Cache-Control") != "no-cache" || h.Get("Service-Worker-Allowed") != "/" {
		t.Fatalf("headers %v", h)
	}
	sw := rr.Body.String()
	m := regexp.MustCompile(`var OFFLINE = "(/__gsui/asset/([0-9a-f]{12})/offline\.html)", CACHE = "gsui-offline-([0-9a-f]{12})";`).FindStringSubmatch(sw)
	if m == nil || m[2] != m[3] {
		t.Fatalf("offline page and cache name missing:\n%s", sw)
	}
	for _, want := range []string{
		"request.mode !== 'navigate' || request.method !== 'GET') return;",
		"self.skipWaiting()", "self.clients.claim()", "key.indexOf('gsui-offline-') === 0 && key !== CACHE",
	} {
		if !strings.Contains(sw, want) {
			t.Errorf("worker missing %q", want)
		}
	}
	if strings.Count(sw, "respondWith") != 1 || strings.Contains(sw, "cache.put") {
		t.Error("worker must answer navigations only and cache nothing but the offline page")
	}
	if !strings.HasSuffix(sw, "\nself.addEventListener('push', function () {});\n") {
		t.Fatalf("Worker not appended:\n%s", sw)
	}
	page := pwaGet(t, app, m[1])
	if !strings.HasPrefix(page.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("offline page type %q", page.Header().Get("Content-Type"))
	}
}

// The offline page is static: no g-sui runtime and nothing loaded from
// elsewhere.
func TestPWAOfflinePage(t *testing.T) {
	app := NewApp()
	app.Lang = "sk"
	app.Texts = Texts{Reload: "Načítať znova"}
	app.PWA(PWA{Name: "Pošta", BackgroundColor: "#fafafa", DarkThemeColor: "#09090b", OfflineText: "Ste offline."})
	sw := pwaGet(t, app, "/sw.js").Body.String()
	page := pwaGet(t, app, regexp.MustCompile(`/__gsui/asset/[0-9a-f]{12}/offline\.html`).FindString(sw)).Body.String()
	for _, want := range []string{`<html lang="sk">`, "<h1>Pošta</h1>", "<p>Ste offline.</p>", "<span>Načítať znova</span>",
		"background:#fafafa;color:#111827", "background:#09090b;color:#f3f4f6", "addEventListener('online'"} {
		if !strings.Contains(page, want) {
			t.Errorf("offline page missing %q", want)
		}
	}
	for _, unwanted := range []string{"__ws", "src=", "href=", "http"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("offline page must be self-contained, has %q", unwanted)
		}
	}
}

// Without App.PWA the shell has no PWA parts; with it the head has the tags
// phones read and the worker registration.
func TestPWAShellHead(t *testing.T) {
	plain := shellBody(t, NewApp())
	for _, unwanted := range []string{"manifest", "theme-color", "serviceWorker", "data-gsui-install", "viewport-fit"} {
		if strings.Contains(plain, unwanted) {
			t.Errorf("shell without PWA has %q", unwanted)
		}
	}
	app := NewApp()
	app.PWA(PWA{Name: "Mail Client", ShortName: "Mail", ThemeColor: "#fafafa", DarkThemeColor: "#09090b", AppleTouchIcon: "/icons/apple.png", Scope: "/app/"})
	body := shellBodyAt(t, app, "/app/")
	for _, want := range []string{
		`<meta name="viewport" content="width=device-width,initial-scale=1">`,
		`<link rel="manifest" href="/manifest.webmanifest">`,
		`<meta name="theme-color" content="#09090b" media="(prefers-color-scheme: dark)">` + "\n" + `<meta name="theme-color" content="#fafafa">`,
		`<meta name="application-name" content="Mail Client">`,
		`<meta name="mobile-web-app-capable" content="yes">`,
		`<meta name="apple-mobile-web-app-capable" content="yes">`,
		`<meta name="apple-mobile-web-app-title" content="Mail">`,
		`<meta name="apple-mobile-web-app-status-bar-style" content="default">`,
		`<link rel="apple-touch-icon" sizes="180x180" href="/icons/apple.png">`,
		`html:not([data-gsui-install]) [data-gsui-installable]{display:none!important}`,
		`navigator.serviceWorker.register('/sw.js', {scope: scope, updateViaCache: 'none'})`,
		`)("/app/");`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell missing %q", want)
		}
	}
	// The client script runs after local.js, which it extends.
	if strings.Index(body, "window.__gsui = {") > strings.Index(body, "g.install = function") {
		t.Error("install support must follow local.js")
	}
	app = NewApp()
	app.PWA(PWA{SafeAreas: true})
	body = shellBody(t, app)
	if !strings.Contains(body, `<meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">`) ||
		strings.Count(body, `media="(prefers-color-scheme: dark)"`) != 0 {
		t.Fatal("SafeAreas viewport or single theme color missing")
	}
}

// shellBodyAt is the page shell of path.
func shellBodyAt(t *testing.T, app *App, path string) string {
	t.Helper()
	app.Page("/{rest...}", func(_ *Context) *Node { return Div() })
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
	return rr.Body.String()
}

// Every configured value is escaped where it lands.
func TestPWAEscapesValues(t *testing.T) {
	evil := `"><script>alert(1)</script>`
	app := NewApp()
	app.PWA(PWA{Name: evil, ShortName: evil, ThemeColor: `red}</style><script>alert(1)</script>`, AppleTouchIcon: evil,
		Scope: `/"</script><script>alert(1)`, OfflineText: evil})
	body := shellBodyAt(t, app, `/%22%3C/script%3E%3Cscript%3Ealert(1)`)
	sw := pwaGet(t, app, "/sw.js").Body.String()
	page := pwaGet(t, app, regexp.MustCompile(`/__gsui/asset/[0-9a-f]{12}/offline\.html`).FindString(sw)).Body.String()
	for name, text := range map[string]string{"shell": body, "offline page": page} {
		if strings.Contains(text, evil) || strings.Contains(text, "<script>alert") {
			t.Errorf("%s does not escape a value:\n%s", name, text)
		}
	}
	if !strings.Contains(page, "background:#ffffff") {
		t.Error("an invalid color must fall back")
	}
	if !strings.Contains(body, `)("/\"\u003c/script\u003e\u003cscript\u003ealert(1)");`) {
		t.Error("scope is not JSON escaped")
	}
}

func TestInstallAppAndShowIfInstallable(t *testing.T) {
	if got := InstallApp("install-help").code(); got != `if(__gsui.install){__gsui.install(el,"install-help");}` {
		t.Fatalf("InstallApp = %s", got)
	}
	js := Button().Text("Install").ShowIfInstallable().OnClick(InstallApp("")).toJS()
	if !strings.Contains(js, `setAttribute('data-gsui-installable','')`) || !strings.Contains(js, `if(__gsui.install){__gsui.install(el,"");}`) {
		t.Fatalf("node js %s", js)
	}
}

// pwaClientShim runs the PWA client script with a fake browser. The case
// sets UA, TOUCH, STANDALONE and SECURE first, then STEPS, and prints the
// install attribute, the gsui:install states, the opened dialogs and the
// service worker registration.
const pwaClientShim = `
var attrs={},events=[],opened=[],listeners={},regs=[],mq={matches:STANDALONE,addEventListener:function(){}};
function fire(type,event){(listeners[type]||[]).forEach(function(f){f(event||{})})}
function promptEvent(outcome,fail){return{preventDefault:function(){this.prevented=true},prompt:function(){if(fail)throw new Error('used');return Promise.resolve()},userChoice:Promise.resolve({outcome:outcome})}}
var win={isSecureContext:SECURE,__gsui:{open:function(el,id){opened.push(id)}},
 addEventListener:function(type,f){(listeners[type]=listeners[type]||[]).push(f)},
 dispatchEvent:function(e){events.push(e.detail.state||'-')},matchMedia:function(){return mq}};
var nav={userAgent:UA,maxTouchPoints:TOUCH,serviceWorker:{register:function(url,opts){regs.push(url+' '+opts.scope+' '+opts.updateViaCache);return Promise.resolve()}}};
function CE(type,init){this.detail=init.detail}
function run(script){new Function('window','document','navigator','matchMedia','CustomEvent',script)(win,{documentElement:{setAttribute:function(k,v){attrs[k]=v},removeAttribute:function(k){delete attrs[k]}}},nav,win.matchMedia,CE)}
`

func TestPWAInstallState(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	// Only the install script: notifications have their own test.
	script, _ := json.Marshal("(" + pwaJS + `)("/app/");`)
	const (
		chrome    = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
		iphone    = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1"
		macSafari = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15"
		macOld    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15"
		macChrome = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
	)
	for _, tc := range []struct {
		name, ua   string
		touch      int
		standalone bool
		insecure   bool
		steps      string
		want       string // attribute | events | opened dialogs | registrations
	}{
		{"no prompt yet", chrome, 0, false, false, "", "||| /sw.js /app/ none"},
		{"prompt offered", chrome, 0, false, false, "var e=promptEvent('accepted');fire('beforeinstallprompt',e);if(!e.prevented)attrs.x='not prevented'", "prompt|prompt|| /sw.js /app/ none"},
		{"prompt accepted", chrome, 0, false, false, "fire('beforeinstallprompt',promptEvent('accepted'));win.__gsui.install(null,'help')", "|prompt,-|| /sw.js /app/ none"},
		{"prompt offered again after dismiss", chrome, 0, false, false, "fire('beforeinstallprompt',promptEvent('dismissed'));win.__gsui.install(null,'help');setTimeout(function(){fire('beforeinstallprompt',promptEvent('accepted'))},0)", "prompt|prompt,-,prompt|| /sw.js /app/ none"},
		{"failed prompt opens help", chrome, 0, false, false, "fire('beforeinstallprompt',promptEvent('accepted',true));win.__gsui.install(null,'help')", "|prompt,-|help| /sw.js /app/ none"},
		{"no prompt opens help", chrome, 0, false, false, "win.__gsui.install(null,'help');win.__gsui.install(null,'')", "||help| /sw.js /app/ none"},
		{"app installed", chrome, 0, false, false, "fire('beforeinstallprompt',promptEvent('accepted'));fire('appinstalled')", "|prompt,-|| /sw.js /app/ none"},
		{"iPhone", iphone, 5, false, false, "", "manual|manual|| /sw.js /app/ none"},
		{"iPad with a Mac user agent", macSafari, 5, false, false, "", "manual|manual|| /sw.js /app/ none"},
		{"macOS Safari 17", macSafari, 0, false, false, "", "manual|manual|| /sw.js /app/ none"},
		{"macOS Safari 16", macOld, 0, false, false, "", "||| /sw.js /app/ none"},
		{"macOS Chrome", macChrome, 0, false, false, "", "||| /sw.js /app/ none"},
		{"installed app", iphone, 5, true, false, "fire('beforeinstallprompt',promptEvent('accepted'))", "||| /sw.js /app/ none"},
		{"insecure page", iphone, 5, false, true, "", "|||"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ua, _ := json.Marshal(tc.ua)
			js := "var UA=" + string(ua) + ",TOUCH=" + strconv.Itoa(tc.touch) + ",STANDALONE=" + map[bool]string{true: "true", false: "false"}[tc.standalone] +
				",SECURE=" + map[bool]string{true: "false", false: "true"}[tc.insecure] + ";" + pwaClientShim +
				"run(" + string(script) + ");" + tc.steps + ";setTimeout(function(){fire('load');setTimeout(function(){console.log([attrs['data-gsui-install']||'',events.join(','),opened.join(','),regs.map(function(r){return ' '+r}).join('')].join('|'))},5)},5)"
			out, err := exec.Command(node, "-e", js).CombinedOutput()
			if err != nil {
				t.Fatalf("node: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// pwaWorkerShim runs the service worker with fake caches and network. ONLINE
// says whether fetch succeeds.
const pwaWorkerShim = `
var handlers={},store={},fetched=[],claimed=false,skipped=false,preload=false;
var self={addEventListener:function(t,f){handlers[t]=f},skipWaiting:function(){skipped=true;return Promise.resolve()},
 clients:{claim:function(){claimed=true;return Promise.resolve()}},
 registration:{navigationPreload:{enable:function(){preload=true;return Promise.resolve()}}}};
function cacheOf(name){var c=store[name]=store[name]||{};return{add:function(req){fetched.push(req.url);c[req.url]='offline page';return Promise.resolve()},match:function(url){return Promise.resolve(c[url])}}}
var caches={open:function(n){return Promise.resolve(cacheOf(n))},keys:function(){return Promise.resolve(Object.keys(store))},delete:function(n){delete store[n];return Promise.resolve(true)}};
function Request(url){this.url=url}
var Response={error:function(){return 'error'}};
function fetch(req){fetched.push(req.url);return ONLINE?Promise.resolve('network '+req.url):Promise.reject(new TypeError('offline'))}
function wait(p){return Promise.resolve(p)}
async function event(type,request){var p=null,answered=false,e={request:request,waitUntil:function(x){p=x},respondWith:function(x){answered=true;p=x}};handlers[type](e);return {answered:answered,value:await wait(p)}}
`

func TestPWAWorkerAnswersOnlyFailedPageLoads(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	app := NewApp()
	app.PWA(PWA{})
	worker := pwaGet(t, app, "/sw.js").Body.String()
	for _, tc := range []struct {
		name   string
		online bool
		want   string
	}{
		{"online", true, "network /inbox|-|-|-"},
		{"offline", false, "offline page|-|-|-"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			js := "var ONLINE=" + strconv.FormatBool(tc.online) + ";" + pwaWorkerShim + worker + `;(async function(){
store['gsui-offline-old']={};store['other']={};
await event('install');await event('activate');
if(!skipped||!claimed||!preload||store['gsui-offline-old']||!store['other']||Object.keys(store).length!==2)throw new Error('lifecycle '+JSON.stringify(store));
var out=[];
for(var r of [{mode:'navigate',method:'GET',url:'/inbox'},{mode:'navigate',method:'POST',url:'/form'},{mode:'cors',method:'GET',url:'/api'},{mode:'no-cors',method:'GET',url:'/__ws.js'}]){
 var res=await event('fetch',r);out.push(res.answered?res.value:'-');
}
console.log(out.join('|'));
})().catch(function(e){console.log(e.message);process.exit(1)})`
			out, err := exec.Command(node, "-e", js).CombinedOutput()
			if err != nil {
				t.Fatalf("node: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

package ui

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

func TestNotificationHelpers(t *testing.T) {
	two := 2
	for got, want := range map[string]string{
		EnableNotifications().code(): `if(__gsui.notifyEnable){__gsui.notifyEnable(el);}`,
		ShowNotification(Notification{Title: "New", URL: "/", Count: &two}).code(): `if(__gsui.notification){__gsui.notification(el,{"title":"New","url":"/","count":2});}`,
	} {
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
	js := Button().ShowIfNotifications(NotificationsOff, NotificationsBlocked).NotificationSwitch().toJS()
	for _, want := range []string{
		`setAttribute('data-gsui-notify-show','off blocked')`, `setAttribute('role','switch')`,
		`setAttribute('data-gsui-notify-switch','')`, `if(__gsui.notifyToggle){__gsui.notifyToggle(el);}`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("node js missing %q:\n%s", want, js)
		}
	}
	if js := Span().AppBadge(-3).toJS(); !strings.Contains(js, `setAttribute('data-gsui-badge','0')`) {
		t.Errorf("AppBadge js %s", js)
	}
}

// The shell gets the state CSS and the client config; the notification icon
// defaults to the first 192x192 icon that is not maskable.
func TestNotificationShell(t *testing.T) {
	app := NewApp()
	app.PWA(PWA{
		Name: "Mail", StartURL: "/inbox",
		Icons: []PWAIcon{
			{Src: "/icon.svg", Sizes: "any"},
			{Src: "/mask-192.png", Sizes: "192x192", Purpose: "maskable"},
			{Src: "/icon-192.png", Sizes: "192x192"},
		},
		NotificationBadge: "/badge.png",
		Push:              &PWAPush{PublicKey: "AQID", Save: func(*Context, PushSubscription, bool) error { return nil }},
	})
	body := shellBody(t, app)
	for _, want := range []string{
		notifyStyle,
		`)({"badge":"/badge.png","icon":"/icon-192.png","key":"AQID","scope":"/","title":"Mail","url":"/inbox"});`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell missing %q", want)
		}
	}
	sw := pwaGet(t, app, "/sw.js").Body.String()
	if !strings.Contains(sw, `var NOTIFY = {"badge":"/badge.png","icon":"/icon-192.png","title":"Mail","url":"/inbox"};`) ||
		!strings.Contains(sw, "addEventListener('push'") {
		t.Errorf("worker config or push handler missing:\n%s", sw)
	}

	// Without Save there is no push: no key, no push handler, no endpoint.
	app = NewApp()
	app.PWA(PWA{Icons: []PWAIcon{{Src: "/icon.svg", Sizes: "any"}}, Push: &PWAPush{PublicKey: "AQID"}})
	if body := shellBody(t, app); !strings.Contains(body, `)({"badge":"","icon":"/icon.svg","scope":"/","title":"App","url":"/"});`) {
		t.Error("Push without Save must not offer push")
	}
	if sw := pwaGet(t, app, "/sw.js").Body.String(); strings.Contains(sw, "addEventListener('push'") || !strings.Contains(sw, "addEventListener('notificationclick'") {
		t.Error("only PWA.Push adds the push handler; clicks are always handled")
	}
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("POST", "/__gsui/push", strings.NewReader("{}")))
	if rr.Code != http.StatusNotFound && rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("push endpoint without Save: %d", rr.Code)
	}
}

// Pages outside PWA.Scope, such as a public site next to /app, get no PWA
// tags or scripts.
func TestPWAOnlyInScope(t *testing.T) {
	app := NewApp()
	app.PWA(PWA{Scope: "/app"})
	app.Page("/", func(*Context) *Node { return Div() })
	app.Page("/app", func(*Context) *Node { return Div() })
	app.Page("/app/team", func(*Context) *Node { return Div() })
	for path, want := range map[string]bool{"/": false, "/app": true, "/app/team": true} {
		rr := httptest.NewRecorder()
		app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		body := rr.Body.String()
		if strings.Contains(body, `rel="manifest"`) != want || strings.Contains(body, "data-gsui-notify") != want {
			t.Errorf("%s: PWA files %v, want %v", path, !want, want)
		}
	}
}

// Live navigation across PWA.Scope loads the page, so the PWA scripts come
// and go with it.
func TestPWAScopeCrossingLoadsPage(t *testing.T) {
	app := NewApp()
	app.PWA(PWA{Scope: "/app"})
	for _, path := range []string{"/", "/about", "/app", "/app/team"} {
		app.Page(path, func(*Context) *Node { return Div().Text("page " + path) })
	}
	for _, tc := range []struct {
		from, to string
		load     bool
	}{
		{"/", "/about", false}, {"/about", "/app", true}, {"/app", "/app/team", false}, {"/app/team", "/", true},
	} {
		// A page load opens a new connection.
		ws := runtimeSocket(t, app)
		js := runtimeCall(t, ws, "__nav", tc.from, 1, map[string]any{"url": tc.to, "history": "push"})
		if strings.Contains(js, "__gsui.redirect(") != tc.load {
			t.Errorf("%s to %s: page load %v, want %v\n%s", tc.from, tc.to, !tc.load, tc.load, js)
		}
	}
}

// pushKeys are valid subscription keys: a P-256 point and a 16-byte secret.
func pushKeys(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	auth := make([]byte, 16)
	_, _ = rand.Read(auth)
	return base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()), base64.RawURLEncoding.EncodeToString(auth)
}

// /__gsui/push takes JSON from the same origin, checks the subscription and
// App.Authorize, then calls Save.
func TestPushEndpoint(t *testing.T) {
	type saved struct {
		sub    PushSubscription
		remove bool
		cookie string
	}
	var got []saved
	var fail, refuse error
	var actions []string
	app := NewApp()
	app.Authorize = func(_ *Context, action string) error { actions = append(actions, action); return refuse }
	app.PWA(PWA{Push: &PWAPush{PublicKey: "AQID", Save: func(ctx *Context, sub PushSubscription, remove bool) error {
		c, _ := ctx.Request.Cookie("session")
		got = append(got, saved{sub, remove, c.Value})
		return fail
	}}})
	h := app.Handler()
	p256dh, auth := pushKeys(t)
	other, _ := pushKeys(t)
	post := func(body, origin, contentType string) int {
		req := httptest.NewRequest("POST", "/__gsui/push", strings.NewReader(body))
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		req.Header.Set("Content-Type", contentType)
		req.AddCookie(&http.Cookie{Name: "session", Value: "s1"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		return rr.Code
	}
	sub := func(endpoint, p256dh, auth string) string {
		b, _ := json.Marshal(map[string]any{"endpoint": endpoint, "keys": map[string]string{"p256dh": p256dh, "auth": auth}})
		return string(b)
	}
	const fcm, origin, jsonType = "https://fcm.googleapis.com/fcm/send/x", "http://example.com", "application/json"
	for _, tc := range []struct {
		name, body, origin, contentType string
		want                            int
	}{
		{"save", sub(fcm, p256dh, auth), origin, jsonType, 204},
		{"remove without keys", `{"endpoint":"https://web.push.apple.com/x","remove":true}`, origin, "application/json; charset=utf-8", 204},
		{"no origin", sub(fcm, p256dh, auth), "", jsonType, 403},
		{"other origin", sub(fcm, p256dh, auth), "https://evil.example", jsonType, 403},
		{"form post", sub(fcm, p256dh, auth), origin, "text/plain", 403},
		{"save without keys", `{"endpoint":"https://fcm.googleapis.com/x"}`, origin, jsonType, 400},
		{"short keys", sub(fcm, "p", "a"), origin, jsonType, 400},
		{"not a point", sub(fcm, base64.RawURLEncoding.EncodeToString(append([]byte{4}, make([]byte, 64)...)), auth), origin, jsonType, 400},
		{"short secret", sub(fcm, other, "AAAA"), origin, jsonType, 400},
		{"other host", sub("https://evil.example/x", p256dh, auth), origin, jsonType, 400},
		{"look-alike host", sub("https://evilpush.apple.com/x", p256dh, auth), origin, jsonType, 400},
		{"plain http", sub("http://fcm.googleapis.com/x", p256dh, auth), origin, jsonType, 400},
		{"port", sub("https://fcm.googleapis.com:8443/x", p256dh, auth), origin, jsonType, 400},
		{"not JSON", "endpoint=x", origin, jsonType, 400},
	} {
		if code := post(tc.body, tc.origin, tc.contentType); code != tc.want {
			t.Errorf("%s: status %d, want %d", tc.name, code, tc.want)
		}
	}
	if len(got) != 2 || got[0].remove || got[0].sub.Keys.P256dh != p256dh || got[0].cookie != "s1" ||
		!got[1].remove || got[1].sub.Endpoint != "https://web.push.apple.com/x" {
		t.Fatalf("saved %+v", got)
	}
	if len(actions) != 2 || actions[0] != "__push" {
		t.Fatalf("Authorize saw %v", actions)
	}
	refuse = errors.New("signed out")
	if code := post(sub(fcm, p256dh, auth), origin, jsonType); code != 403 || len(got) != 2 {
		t.Fatalf("refused by Authorize: %d, saved %d", code, len(got))
	}
	refuse, fail = nil, errors.New("demo")
	if code := post(sub(fcm, p256dh, auth), origin, jsonType); code != 403 {
		t.Fatalf("failed Save: %d", code)
	}
}

// notifyClientShim runs the notifications script with a fake browser. The
// case sets UA, HAS (Notification API), PERM, ASK (the permission answer),
// SUB (an existing subscription's key or null), OLD (one in another
// registration, or null) and SAVE (the push endpoint's answer).
const notifyClientShim = `
var attrs={},events=[],calls=[],toasts=[],shown=[],badges=[],perm=PERM,ready=[];
function Sub(bytes){this.endpoint='https://fcm.googleapis.com/x';this.options={applicationServerKey:new Uint8Array(bytes).buffer}}
Sub.prototype.toJSON=function(){return{endpoint:this.endpoint,keys:{p256dh:'p',auth:'a'}}};
var UNSUB='';
Sub.prototype.unsubscribe=function(){calls.push('unsubscribe');
 if(UNSUB==='fail')return Promise.reject(new Error('unsubscribe'));
 if(UNSUB==='slow')return new Promise(function(r){setTimeout(function(){sub=null;calls.push('unsubscribed');r(true)},150)});
 sub=null;return Promise.resolve(true)};
var sub=SUB?new Sub(SUB):null,old=OLD?new Sub(OLD):null;
if(old)old.unsubscribe=function(){old=null;calls.push('unsubscribe old');return Promise.resolve(true)};
var reg={scope:'https://app.test/',active:{scriptURL:'https://app.test/sw.js'},pushManager:{getSubscription:function(){return Promise.resolve(sub)},
 subscribe:function(o){calls.push('subscribe '+Array.from(o.applicationServerKey));if(typeof STALL!=='undefined'&&STALL){STALL=false;return new Promise(function(){})}
  sub=new Sub(Array.from(o.applicationServerKey));return Promise.resolve(sub)}},
 showNotification:function(t,o){shown.push([t,o.body,o.data.url,o.icon,o.badge,o.tag].join(','));return Promise.resolve()}};
var oldReg={scope:'https://app.test/old/',active:{scriptURL:'https://app.test/sw.js'},pushManager:{getSubscription:function(){return Promise.resolve(old)}}};
var other=new Sub([1,2,3]);other.unsubscribe=function(){calls.push('unsubscribe other');return Promise.resolve(true)};
var otherReg={scope:'https://app.test/other/',active:{scriptURL:'https://app.test/other/worker.js'},pushManager:{getSubscription:function(){return Promise.resolve(other)}}};
// ready would give the other app's worker; g-sui takes its own registration.
var registered=[];
var sw={ready:Promise.resolve(otherReg),register:function(u,o){registered.push(u+' '+o.scope+' '+o.updateViaCache);return Promise.resolve(reg)},
 getRegistrations:function(){return Promise.resolve([otherReg,oldReg,reg])}};
var nav={userAgent:UA,maxTouchPoints:5,serviceWorker:sw,setAppBadge:function(n){badges.push(n);return Promise.resolve()},clearAppBadge:function(){badges.push(0);return Promise.resolve()}};
var listeners={};
var win={isSecureContext:true,PushManager:function(){},__gsui:{notify:function(el,v,m){toasts.push(v+' '+m)}},
 matchMedia:function(){return{matches:false}},dispatchEvent:function(e){events.push(e.detail.state||'-')},
 addEventListener:function(t,f){listeners[t]=f}};
if(HAS)win.Notification={get permission(){return perm},requestPermission:function(){perm=ASK;return Promise.resolve(perm)}};
var slow=typeof SLOW!=='undefined'&&SLOW;
function fetch(url,o){var data=JSON.parse(o.body);
 if(url!=='/__gsui/push'||o.method!=='POST'||o.headers['Content-Type']!=='application/json')throw new Error('bad request');
 calls.push('__push '+(data.remove?'remove ':'')+data.endpoint+' '+(data.keys&&data.keys.auth||''));
 if(!slow)return Promise.resolve({ok:SAVE});
 if(slow!=='always')slow=false;return new Promise(function(r){setTimeout(function(){calls.push('answered');r({ok:SAVE})},40)})}
var observed=null;win.MutationObserver=function(f){observed=f};win.MutationObserver.prototype.observe=function(){};
function el(a){return{attrs:a,disabled:false,getAttribute:function(k){return k in this.attrs?this.attrs[k]:null},setAttribute:function(k,v){this.attrs[k]=v}}}
var toggle=el({}),badge=null,loaded=[];
var doc={documentElement:{setAttribute:function(k,v){attrs[k]=v},removeAttribute:function(k){delete attrs[k]}},
 querySelectorAll:function(s){return s==='[data-gsui-notify-switch]'?[toggle]:[]},querySelector:function(){return badge},
 addEventListener:function(t,f){loaded.push(f)}};
function CE(type,init){this.detail=init.detail}
function run(script){new Function('window','document','navigator','matchMedia','CustomEvent','Notification','fetch','location',script)(win,doc,nav,win.matchMedia,CE,win.Notification,fetch,{href:'https://app.test/',origin:'https://app.test'})}
function settle(){return new Promise(function(r){setTimeout(r,20)})}
`

func TestNotificationClient(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	const (
		chrome = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
		iphone = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1"
	)
	for _, tc := range []struct {
		name, ua, key, perm, ask, sub string
		has, save                     bool
		steps                         string
		want                          string // state | events | calls | toasts | switch checked, disabled
	}{
		{"iPhone browser", iphone, "", "default", "", "null", false, true, "", "install|install|||false,true"},
		{"no notifications", chrome, "", "default", "", "null", false, true, "", "|-|||false,true"},
		{"blocked", chrome, "", "denied", "", "null", true, true, "", "blocked|blocked|||false,true"},
		{"page: enable", chrome, "", "default", "granted", "null", true, true, "win.__gsui.notifyEnable()", "on|off,on|||true,false"},
		{"page: refused", chrome, "", "default", "denied", "null", true, true, "win.__gsui.notifyToggle()", "blocked|off,blocked|||false,true"},
		{"page: no switching off", chrome, "", "granted", "", "null", true, true, "win.__gsui.notifyToggle()", "on|on|||true,false"},
		{"push: load saves again", chrome, "AQID", "granted", "", "[1,2,3]", true, false, "", "on|on|__push https://fcm.googleapis.com/x a||true,false"},
		{"push: other key is off", chrome, "AQID", "granted", "", "[9]", true, true, "", "off|off|||false,false"},
		{"push: enable", chrome, "AQID", "default", "granted", "null", true, true, "win.__gsui.notifyToggle()",
			"on|off,on|subscribe 1,2,3,__push https://fcm.googleapis.com/x a||true,false"},
		{"push: enable replaces other key", chrome, "AQID", "granted", "", "[9]", true, true, "win.__gsui.notifyToggle()",
			"on|off,on|unsubscribe,subscribe 1,2,3,__push https://fcm.googleapis.com/x a||true,false"},
		{"push: failed save undoes the subscription", chrome, "AQID", "default", "granted", "null", true, false, "win.__gsui.notifyToggle()",
			"off|off|subscribe 1,2,3,__push https://fcm.googleapis.com/x a,unsubscribe|error Notifications could not be changed. Try again.|false,false"},
		{"push: disable", chrome, "AQID", "granted", "", "[1,2,3]", true, true, "calls=[];win.__gsui.notifyToggle()",
			"off|on,off|__push remove https://fcm.googleapis.com/x a,unsubscribe||false,false"},
		{"push: failed disable stays on", chrome, "AQID", "granted", "", "[1,2,3]", true, false, "calls=[];win.__gsui.notifyToggle()",
			"on|on|__push remove https://fcm.googleapis.com/x a,__push https://fcm.googleapis.com/x a|error Notifications could not be changed. Try again.|true,false"},
		{"push: failed load save is retried after navigation", chrome, "AQID", "granted", "", "[1,2,3]", true, false,
			"SAVE=true;listeners['gsui:navigated']();await settle();listeners['gsui:navigated']()",
			"on|on|__push https://fcm.googleapis.com/x a,__push https://fcm.googleapis.com/x a||true,false"},
		{"push: old registration keeps its subscription", chrome, "AQID", "granted", "", "[1,2,3]", true, true, "calls=[];win.__gsui.notifyToggle()",
			"off|on,off|__push remove https://fcm.googleapis.com/x a,unsubscribe old||false,false"},
		{"push: other workers are not touched", chrome, "AQID", "granted", "", "null", true, true, "win.__gsui.notifyToggle()",
			"on|off,on|subscribe 1,2,3,__push https://fcm.googleapis.com/x a||true,false"},
		{"push: a bad key turns push off", chrome, "%%%", "granted", "", "[1,2,3]", true, true, "win.__gsui.notifyToggle()",
			"[g-sui] PWA.Push.PublicKey is not base64url\non|on|||true,false"},
		{"push: saves run in turn", chrome, "AQID", "granted", "", "[1,2,3]", true, true, "SLOW=true",
			"off|on,off|__push https://fcm.googleapis.com/x a,answered,__push remove https://fcm.googleapis.com/x a,unsubscribe||false,false"},
		{"push: a stalled step does not block a retry", chrome, "AQID", "granted", "", "null", true, true,
			"STALL=true;win.__gsui.notifyToggle();await new Promise(function(r){setTimeout(r,150)});win.__gsui.notifyToggle()",
			"on|off,on|subscribe 1,2,3,subscribe 1,2,3,__push https://fcm.googleapis.com/x a|error Notifications could not be changed. Try again.|true,false"},
		{"push: a removal whose caller gave up is not sent", chrome, "AQID", "granted", "", "[1,2,3]", true, true, "SLOW=true",
			"on|on|__push https://fcm.googleapis.com/x a,answered,__push https://fcm.googleapis.com/x a|error Notifications could not be changed. Try again.|true,false"},
		{"push: a timed out enable does not subscribe later", chrome, "AQID", "granted", "", "[9]", true, true,
			"UNSUB='slow';win.__gsui.notifyToggle();await new Promise(function(r){setTimeout(r,250)})",
			"off|off|unsubscribe,unsubscribed|error Notifications could not be changed. Try again.|false,false"},
		{"push: a late unsubscribe is read again", chrome, "AQID", "granted", "", "[1,2,3]", true, true,
			"calls=[];UNSUB='slow';win.__gsui.notifyToggle();await new Promise(function(r){setTimeout(r,250)})",
			"off|on,off|__push remove https://fcm.googleapis.com/x a,unsubscribe,unsubscribed|error Notifications could not be changed. Try again.|false,false"},
		{"push: a failed unsubscribe saves the device again", chrome, "AQID", "granted", "", "[1,2,3]", true, true,
			"calls=[];UNSUB='fail';win.__gsui.notifyToggle()",
			"on|on|__push remove https://fcm.googleapis.com/x a,unsubscribe,__push https://fcm.googleapis.com/x a|error Notifications could not be changed. Try again.|true,false"},
		{"push: slow saves are not repeated", chrome, "AQID", "granted", "", "[1,2,3]", true, true,
			"await new Promise(function(r){setTimeout(r,250)})",
			"on|on|__push https://fcm.googleapis.com/x a,answered||true,false"},
		{"push: a subscription removed elsewhere reads as off", chrome, "AQID", "granted", "", "[1,2,3]", true, true,
			"calls=[];sub=null;listeners['gsui:navigated']()", "off|on,off|||false,false"},
		{"push: a renewed subscription is saved", chrome, "AQID", "granted", "", "[1,2,3]", true, true,
			"calls=[];sub=new Sub([1,2,3]);sub.endpoint='https://fcm.googleapis.com/y';listeners['gsui:navigated']();await settle();listeners['gsui:navigated']()",
			"on|on|__push https://fcm.googleapis.com/y a||true,false"},
		{"switch after a morph", chrome, "", "granted", "", "null", true, true, "delete toggle.attrs['aria-checked'];toggle.disabled=true;observed()",
			"on|on|||true,false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ua, _ := json.Marshal(tc.ua)
			config, _ := json.Marshal(map[string]any{"scope": "/", "title": "Mail", "url": "/", "icon": "/i.png", "badge": "", "key": tc.key})
			// These cases run with 1 ms for every second of the time limits.
			if strings.Contains(tc.name, "stalled") || strings.Contains(tc.name, "gave up") || strings.Contains(tc.name, "late") ||
				strings.Contains(tc.name, "unsubscribe") || strings.Contains(tc.name, "slow saves") {
				config, _ = json.Marshal(map[string]any{"scope": "/", "title": "Mail", "url": "/", "icon": "/i.png", "badge": "", "key": tc.key, "unit": 1})
			}
			script, _ := json.Marshal("(" + notifyJS + ")(" + string(config) + ");")
			old := "null"
			if strings.HasPrefix(tc.name, "push: old registration") {
				old, tc.sub = tc.sub, "null"
			}
			slow := tc.steps == "SLOW=true"
			if slow {
				tc.steps = "await new Promise(function(r){setTimeout(r,100)})"
				script, _ = json.Marshal("(" + notifyJS + ")(" + string(config) + ");setTimeout(function(){win.__gsui.notifyToggle()},5);")
			}
			slowJS := boolJS(slow)
			if tc.name == "push: slow saves are not repeated" {
				slowJS = "'always'"
			}
			js := "var STALL=false,SLOW=" + slowJS + ",UA=" + string(ua) + ",HAS=" + boolJS(tc.has) + ",PERM='" + tc.perm + "',ASK='" + tc.ask + "',SUB=" + tc.sub + ",OLD=" + old + ",SAVE=" + boolJS(tc.save) + ";" +
				notifyClientShim + "(async function(){run(" + string(script) + ");await settle();" + tc.steps + ";await settle();" +
				"console.log([attrs['data-gsui-notify']||'',events.join(','),calls.join(','),toasts.join(','),[toggle.attrs['aria-checked'],toggle.disabled].join(',')].join('|'));process.exit(0)})()"
			out, err := exec.Command(node, "-e", js).CombinedOutput()
			if err != nil {
				t.Fatalf("node: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != tc.want {
				t.Fatalf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

// ShowNotification goes through the service worker; counts go on the icon,
// from the notification and from AppBadge nodes.
func TestNotificationShowAndBadge(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	config, _ := json.Marshal(map[string]string{"scope": "/", "title": "Mail", "url": "/inbox", "icon": "/i.png", "badge": "/b.png"})
	script, _ := json.Marshal("(" + notifyJS + ")(" + string(config) + ");")
	js := `var UA='Chrome',HAS=true,PERM='granted',ASK='',SUB=null,OLD=null,SAVE=true;` + notifyClientShim + `(async function(){
run(` + string(script) + `);await settle();
win.__gsui.notification(null,{body:'2 new',tag:'mail',count:3});
win.__gsui.notification(null,{title:'Planned',url:'/p',icon:'/x.png'});
badge=el({'data-gsui-badge':'2'});loaded.forEach(function(f){f()});loaded.forEach(function(f){f()});
badge=el({'data-gsui-badge':'2'});loaded.forEach(function(f){f()});
badge=null;loaded.forEach(function(f){f()});
perm='default';win.__gsui.notification(null,{title:'Hidden'});
await settle();console.log(shown.join('|')+' # '+badges.join(',')+' # '+registered[0]);process.exit(0)})()`
	out, err := exec.Command(node, "-e", js).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	want := "Mail,2 new,/inbox,/i.png,/b.png,mail|Planned,,/p,/x.png,/b.png, # 3,2,2 # /sw.js / none"
	if got := strings.TrimSpace(string(out)); got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// ShowNotification waits for /sw.js to replace a previous worker of the
// scope, so the notification belongs to g-sui's worker.
func TestNotificationWaitsForOwnWorker(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	config, _ := json.Marshal(map[string]string{"scope": "/", "title": "Mail", "url": "/", "icon": "/i.png", "badge": ""})
	script, _ := json.Marshal("(" + notifyJS + ")(" + string(config) + ");")
	js := `var UA='Chrome',HAS=true,PERM='granted',ASK='',SUB=null,OLD=null,SAVE=true;` + notifyClientShim + `(async function(){
var change=null;reg.active={scriptURL:'https://app.test/previous.js'};
reg.installing={state:'installing',addEventListener:function(t,f){change=f}};reg.addEventListener=function(){};
run(` + string(script) + `);await settle();
win.__gsui.notification(null,{body:'hi'});await settle();
var before=shown.length;
reg.active={scriptURL:'https://app.test/sw.js'};reg.installing.state='activated';change();await settle();
console.log(before+' '+shown.join('|'));process.exit(0)})()`
	out, err := exec.Command(node, "-e", js).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "0 Mail,hi,/,/i.png,," {
		t.Fatalf("got %q", got)
	}
}

// notifyWorkerShim runs the service worker's notification handlers in the
// scope https://app.test/app/.
const notifyWorkerShim = `
var handlers={},shown=[],badges=[],opened=[],focused=[],navigated=[],windows=WINDOWS;
var self={addEventListener:function(t,f){handlers[t]=f},location:{origin:'https://app.test'},
 navigator:{setAppBadge:function(n){badges.push(n);return Promise.resolve()},clearAppBadge:function(){badges.push(0);return Promise.resolve()}},
 registration:{scope:'https://app.test/app/',showNotification:function(t,o){shown.push([t,o.body,o.data.url,o.data.gsui,o.icon,o.badge,o.tag].join(','));return Promise.resolve()}},
 clients:{matchAll:function(){return Promise.resolve(windows.map(function(u){return{url:u,focus:function(){focused.push(u);var c=this;return Promise.resolve(c)},navigate:function(x){navigated.push(x);return Promise.resolve()}}}))},
  openWindow:function(u){opened.push(u);return Promise.resolve()}}};
async function fire(type,e){var p,closed=false;e.waitUntil=function(x){p=x};if(e.notification)e.notification.close=function(){closed=true};handlers[type](e);await p;return closed}
function push(data){return fire('push',{data:data===undefined?null:{json:function(){if(data==='bad')throw new Error('bad');return data}}})}
function click(url,mine){return fire('notificationclick',{notification:{data:mine===false?{url:url}:{url:url,gsui:true}}})}
`

func TestNotificationWorker(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	app := NewApp()
	app.PWA(PWA{Name: "Mail", Scope: "/app/", StartURL: "/app/inbox", NotificationIcon: "/i.png", NotificationBadge: "/b.png",
		Push: &PWAPush{PublicKey: "AQID", Save: func(*Context, PushSubscription, bool) error { return nil }}})
	worker := pwaGet(t, app, "/sw.js").Body.String()
	for _, tc := range []struct{ name, windows, steps, want string }{
		{"push", "[]", "await push({title:'New',body:'Hi',url:'/app/m/1',tag:'t',count:4});await push({count:0});await push('bad');await push();await push(null);await push({title:{toString:null},body:5,url:['/x']})",
			"New,Hi,/app/m/1,true,/i.png,/b.png,t|Mail,,/app/inbox,true,/i.png,/b.png,|Mail,,/app/inbox,true,/i.png,/b.png,|Mail,,/app/inbox,true,/i.png,/b.png,|Mail,,/app/inbox,true,/i.png,/b.png,|Mail,,/app/inbox,true,/i.png,/b.png, # 4,0"},
		{"click opens the app", "[]", "await click('/app/m/1');await click('https://evil.example/');await click('https://[');await click()",
			"opened https://app.test/app/m/1,https://app.test/app/inbox,https://app.test/app/inbox,https://app.test/app/inbox"},
		{"click uses a window of the app", "['https://app.test/other/unsaved','https://other.test/app/','https://app.test/app/inbox']", "await click('/app/m/1')",
			"focused https://app.test/app/inbox navigated https://app.test/app/m/1"},
		{"click prefers the same page", "['https://app.test/app/inbox','https://app.test/app/m/1']", "await click('/app/m/1')",
			"focused https://app.test/app/m/1"},
		{"the app's own notifications are its own", "[]", "if(await click('/app/m/1',false))throw new Error('closed')", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			js := "var WINDOWS=" + tc.windows + ";" + notifyWorkerShim + worker + `;(async function(){` + tc.steps + `;
var out=[];if(shown.length)out.push(shown.join('|')+' # '+badges.join(','));if(opened.length)out.push('opened '+opened.join(','));
if(focused.length)out.push('focused '+focused.join(','));if(navigated.length)out.push('navigated '+navigated.join(','));console.log(out.join(' '))})().catch(function(e){console.log(e.message);process.exit(1)})`
			out, err := exec.Command(node, "-e", js).CombinedOutput()
			if err != nil {
				t.Fatalf("node: %v\n%s", err, out)
			}
			if got := strings.TrimSpace(string(out)); got != tc.want {
				t.Fatalf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

func boolJS(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

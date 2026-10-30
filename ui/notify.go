package ui

import (
	"crypto/ecdh"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//go:embed notify.js
var notifyJS string

// Notification is a system notification: in the status bar on phones, in
// the notification center on computers. ShowNotification shows one from an
// open page. As a Web Push payload (json.Marshal it) the g-sui service worker
// shows it while the app is closed. Empty fields get the PWA's defaults.
type Notification struct {
	Title string `json:"title,omitempty"` // default PWA.Name
	Body  string `json:"body,omitempty"`
	URL   string `json:"url,omitempty"`  // local page a click opens; default PWA.StartURL
	Tag   string `json:"tag,omitempty"`  // a newer notification with the same tag replaces the older one
	Icon  string `json:"icon,omitempty"` // default PWA.NotificationIcon
	// Count is the number on the app icon: 0 clears it, nil leaves it.
	Count *int `json:"count,omitempty"`
}

// PWAPush turns on Web Push: notifications that arrive while the app is
// closed. g-sui subscribes the device and shows the pushes; the app stores
// the subscriptions and sends the pushes with a Web Push library, signed
// with the VAPID key pair. Send them with an HTTP client that does not
// follow redirects.
type PWAPush struct {
	// PublicKey is the VAPID public key, base64url encoded.
	PublicKey string
	// Save stores this device's subscription for the user of ctx, or removes
	// it when remove is true. The page posts it to /__gsui/push: same origin
	// only, after App.Identity and App.Authorize (action "__push"), with the
	// request's cookies in ctx.Request. A page saves again on every load
	// while notifications are on, so Save also moves a device to the user
	// who signed in last. g-sui checks the endpoint and keys before calling
	// it.
	Save func(ctx *Context, sub PushSubscription, remove bool) error
}

// PushSubscription is one device's Web Push subscription.
type PushSubscription struct {
	Endpoint string   `json:"endpoint"`
	Keys     PushKeys `json:"keys"`
}

// PushKeys are a subscription's encryption keys.
type PushKeys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

// Notification states for ShowIfNotifications.
const (
	NotificationsOff     = "off"     // can be turned on
	NotificationsOn      = "on"      // allowed, and with PWA.Push subscribed
	NotificationsBlocked = "blocked" // the user blocked them in the browser
	NotificationsInstall = "install" // iPhone and iPad allow them only in the installed app
)

// EnableNotifications asks for permission to show notifications and, with
// PWA.Push, subscribes this device. Call it from a click: browsers ignore
// the request otherwise. Needs App.PWA.
func EnableNotifications() *Action {
	return wrap(op("notifyEnable"), "if(__gsui.notifyEnable){", "}")
}

// ShowNotification shows a system notification now, if the user allowed
// them. A click opens n.URL in the app. Needs App.PWA.
func ShowNotification(n Notification) *Action {
	return wrap(op("notification", n), "if(__gsui.notification){", "}")
}

// ShowIfNotifications hides the node unless the notification state is one
// of states (NotificationsOff, NotificationsOn, ...). Where the browser has
// no notifications it stays hidden. Without App.PWA it always shows.
func (n *Node) ShowIfNotifications(states ...string) *Node {
	return n.Attr("data-gsui-notify-show", strings.Join(states, " "))
}

// NotificationSwitch makes the node an on/off switch for notifications on
// this device: a click turns them on (EnableNotifications) or, with
// PWA.Push, off. g-sui keeps aria-checked and disabled in step with the
// state. Needs App.PWA.
func (n *Node) NotificationSwitch() *Node {
	return n.Attr("role", "switch").Attr("data-gsui-notify-switch", "").
		OnClick(wrap(op("notifyToggle"), "if(__gsui.notifyToggle){", "}"))
}

// AppBadge puts count on the app icon while the node is on the page; 0
// clears it. Needs App.PWA, and the browser must support app badges.
func (n *Node) AppBadge(count int) *Node {
	return n.Attr("data-gsui-badge", strconv.Itoa(max(count, 0)))
}

// notifyStyle hides ShowIfNotifications nodes outside their states.
const notifyStyle = "html:not([data-gsui-notify]) [data-gsui-notify-show]{display:none!important}" +
	"html[data-gsui-notify=off] [data-gsui-notify-show]:not([data-gsui-notify-show~=off])," +
	"html[data-gsui-notify=on] [data-gsui-notify-show]:not([data-gsui-notify-show~=on])," +
	"html[data-gsui-notify=blocked] [data-gsui-notify-show]:not([data-gsui-notify-show~=blocked])," +
	"html[data-gsui-notify=install] [data-gsui-notify-show]:not([data-gsui-notify-show~=install]){display:none!important}"

// notifyScript is the client script of notifications, run after pwa.js.
func notifyScript(p PWA) string {
	config := map[string]string{"scope": p.Scope, "title": p.Name, "url": p.StartURL, "icon": p.NotificationIcon, "badge": p.NotificationBadge}
	if p.Push != nil && p.Push.Save != nil {
		config["key"] = p.Push.PublicKey
	}
	b, _ := json.Marshal(config)
	return "\n(" + notifyJS + ")(" + string(b) + ");"
}

// notifyWorker handles clicks on g-sui's notifications and, with
// PWA.Push, shows pushes. Notifications of the app's own PWA.Worker code
// lack data.gsui, so their clicks stay the app's.
func notifyWorker(p PWA) string {
	b, _ := json.Marshal(map[string]string{"title": p.Name, "url": p.StartURL, "icon": p.NotificationIcon, "badge": p.NotificationBadge})
	worker := `var NOTIFY = ` + string(b) + `;
// A click shows the notification's page, same origin only, in a window of
// the app: one already there, else any in the worker's scope, else a new one.
self.addEventListener('notificationclick', function (event) {
  var data = event.notification.data;
  if (!data || !data.gsui) return;
  event.notification.close();
  var url;
  try { url = new URL(data.url || NOTIFY.url, self.location.origin); } catch (_) {}
  if (!url || url.origin !== self.location.origin) url = new URL(NOTIFY.url, self.location.origin);
  var scope = self.registration.scope;
  event.waitUntil(self.clients.matchAll({type: 'window', includeUncontrolled: true}).then(function (list) {
    list = list.filter(function (c) { return c.url.indexOf(scope) === 0 && 'focus' in c; });
    var same = list.filter(function (c) { return c.url === url.href; })[0];
    if (same) return same.focus();
    if (list[0]) return list[0].focus().then(function (w) { return w && w.navigate ? w.navigate(url.href) : w; });
    return self.clients.openWindow(url.href);
  }));
});
`
	if p.Push != nil && p.Push.Save != nil {
		worker += `// A push carries a Notification as JSON; its count goes on the app icon.
self.addEventListener('push', function (event) {
  var data = null;
  try { data = event.data ? event.data.json() : null; } catch (_) {}
  if (!data || typeof data !== 'object') data = {};
  // Only text fields count; anything else gets the default, so every push
  // shows a notification.
  function text(v, d) { return typeof v === 'string' && v ? v : d; }
  var tasks = [self.registration.showNotification(text(data.title, NOTIFY.title), {
    body: text(data.body, ''), tag: text(data.tag, undefined), data: {url: text(data.url, NOTIFY.url), gsui: true},
    icon: text(data.icon, NOTIFY.icon) || undefined, badge: NOTIFY.badge || undefined
  })];
  if (typeof data.count === 'number' && self.navigator.setAppBadge) {
    tasks.push((data.count > 0 ? self.navigator.setAppBadge(data.count) : self.navigator.clearAppBadge()).catch(function () {}));
  }
  event.waitUntil(Promise.all(tasks));
});
`
	}
	return worker
}

// pushHosts are the browsers' push services. Only these are saved, so a
// subscription cannot point the app's push requests at another host.
var pushHosts = []string{"fcm.googleapis.com", ".push.apple.com", ".push.services.mozilla.com", ".notify.windows.com"}

// pushEndpointAllowed accepts HTTPS endpoints of the browsers' push services.
func pushEndpointAllowed(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || len(endpoint) > 1024 {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, allowed := range pushHosts {
		if host == allowed || (strings.HasPrefix(allowed, ".") && strings.HasSuffix(host, allowed)) {
			return true
		}
	}
	return false
}

// pushKeysValid accepts a P-256 public key and a 16-byte auth secret,
// base64url encoded, as RFC 8291 needs them.
func pushKeysValid(keys PushKeys) bool {
	decode := func(s string) []byte {
		b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
		if err != nil {
			return nil
		}
		return b
	}
	public := decode(keys.P256dh)
	if _, err := ecdh.P256().NewPublicKey(public); err != nil || len(public) != 65 {
		return false
	}
	return len(decode(keys.Auth)) == 16
}

// pushHandler serves POST /__gsui/push: it saves or removes the posting
// device's subscription through PWA.Push.Save. A JSON body from the same
// origin only, so another site cannot post it.
func (app *App) pushHandler(save func(*Context, PushSubscription, bool) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if r.Header.Get("Origin") == "" || app.wsHandshake(r) != nil || media != "application/json" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var in struct {
			PushSubscription
			Remove bool `json:"remove"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&in); err != nil ||
			!pushEndpointAllowed(in.Endpoint) || !in.Remove && !pushKeysValid(in.Keys) {
			http.Error(w, "Invalid subscription", http.StatusBadRequest)
			return
		}
		ctx := &Context{Request: r, Session: map[string]any{}, app: app}
		ctx.setRequest(r)
		if err := app.authorize(ctx, "__push"); err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if err := save(ctx, in.PushSubscription, in.Remove); err != nil {
			log.Printf("gsui: push subscription: %v", err)
			http.Error(w, "Not saved", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

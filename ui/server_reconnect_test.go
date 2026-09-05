package ui

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/websocket"
)

func shellBody(t *testing.T, app *App) string {
	t.Helper()
	app.Page("/", func(_ *Context) *Node { return Div() })
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	return rr.Body.String()
}

func TestReconnectConfigDefaults(t *testing.T) {
	body := shellBody(t, NewApp())
	if !strings.Contains(body, `window.__gsuiCfg={grace:3000,reloadAfter:15000,keepAlive:25000,inst:'`+serverInstanceID+`'};`) {
		t.Fatalf("default reconnect config missing from shell:\n%s", body)
	}
}

func TestReconnectConfigOverrides(t *testing.T) {
	app := NewApp()
	app.OfflineGraceMs = 5000
	app.ReconnectReloadAfterMs = -1
	app.KeepAliveMs = -1
	body := shellBody(t, app)
	if !strings.Contains(body, `window.__gsuiCfg={grace:5000,reloadAfter:-1,keepAlive:-1,inst:'`+serverInstanceID+`'};`) {
		t.Fatalf("reconnect config overrides missing from shell:\n%s", body)
	}
}

func TestClientKeepsPageUsableWhileOffline(t *testing.T) {
	// The offline indicator must never block input or reload eagerly.
	if strings.Contains(wsClientJS, "pointer-events-none") {
		t.Error("offline indicator must not disable pointer events on the body")
	}
	for _, want := range []string{
		"__offline.schedule()",
		"window.__gsuiHolds>0",
		"gsui:reconnected",
		"message('__ping'",
	} {
		if !strings.Contains(wsClientJS, want) {
			t.Errorf("client script missing %q", want)
		}
	}
}

// The stub runs before /__ws.js loads; every documented helper must exist on
// it, and hold() must share its counter with the real client.
func TestStubExposesFullClientAPI(t *testing.T) {
	for _, want := range []string{"call", "subscribe", "unsubscribe", "notfound", "hold", "holds", "connected", "offline", "reconnect"} {
		if !strings.Contains(wsStubJS, want+":function") {
			t.Errorf("stub missing %q", want)
		}
	}
	if !strings.Contains(wsStubJS, "window.__gsuiHolds") || !strings.Contains(wsClientJS, "window.__gsuiHolds") {
		t.Error("hold counter must be shared between stub and client via window.__gsuiHolds")
	}
	// Replaying a queued hold() would double-count it.
	if strings.Contains(wsStubJS, "__q.push(['hold'") {
		t.Error("hold() must not be queued for replay")
	}
}

func TestPingActionIsRegistered(t *testing.T) {
	app := NewApp()
	app.setup()
	app.mu.RLock()
	_, ok := app.actions["__ping"]
	app.mu.RUnlock()
	if !ok {
		t.Fatal("__ping action not registered")
	}
}

// TestClientJSBehavior runs the embedded client source in Node against a fake
// DOM and a scriptable WebSocket (ui/testdata/client_harness.js), covering the
// reconnect state machine, rejected offline calls, holds, subscriptions and the badge.
func TestClientJSBehavior(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	dir := t.TempDir()
	stub := filepath.Join(dir, "stub.js")
	client := filepath.Join(dir, "client.js")
	if err := os.WriteFile(stub, []byte(wsStubJS), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(client, []byte(wsClientJS), 0o600); err != nil {
		t.Fatal(err)
	}
	// Copy the harness through the test binary so it counts as a test input:
	// node reads it in a subprocess, which go test's cache would not notice.
	harness, err := os.ReadFile("testdata/client_harness.js")
	if err != nil {
		t.Fatal(err)
	}
	harnessCopy := filepath.Join(dir, "client_harness.js")
	if err := os.WriteFile(harnessCopy, harness, 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, harnessCopy, stub, client).CombinedOutput()
	if err != nil {
		t.Fatalf("client harness failed: %v\n%s", err, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}

// A restarted server hands out fresh ui.Target() ids, so a page rendered by an
// earlier process can no longer be patched. The instance handshake is what
// lets the client notice and resync.
func TestServerAnnouncesInstanceOnConnect(t *testing.T) {
	app := NewApp()
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	config, err := websocket.NewConfig("ws"+strings.TrimPrefix(server.URL, "http")+"/__ws", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var raw string
	if err := websocket.Message.Receive(conn, &raw); err != nil {
		t.Fatalf("receive hello frame: %v", err)
	}
	var hello struct {
		Hello string `json:"__hello"`
	}
	if err := json.Unmarshal([]byte(raw), &hello); err != nil {
		t.Fatalf("decode hello frame %q: %v", raw, err)
	}
	if hello.Hello == "" || hello.Hello != serverInstanceID {
		t.Fatalf("hello frame %q does not carry the instance id %q", raw, serverInstanceID)
	}
	// The id in the page shell must match, otherwise every client would
	// reload itself on its first connection.
	if !strings.Contains(shellBody(t, NewApp()), "inst:'"+serverInstanceID+"'") {
		t.Fatal("page shell does not carry the instance id announced over the socket")
	}
}

// The client must reload a DOM the current server cannot patch, and must not
// destroy in-progress work while doing so.
func TestClientResyncsAfterServerRestart(t *testing.T) {
	for _, want := range []string{
		"m.__hello",
		"gsui:serverchanged",
		"gsui:reloadpending",
		"function requestReload(",
	} {
		if !strings.Contains(wsClientJS, want) {
			t.Errorf("client script missing %q", want)
		}
	}
}

// Package proxy provides a reverse proxy server with WebSocket support and URL rewriting.
// It forwards HTTP and WebSocket requests to a target server while rewriting port references
// in responses to maintain transparent proxying.
package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Config holds the proxy configuration
type Config struct {
	ProxyPort  string // Port to listen on (e.g., "8640")
	TargetPort string // Port to forward to (e.g., "8642")
	TargetHost string // Host to forward to (e.g., "localhost")
	Logger     *log.Logger
}

// Proxy represents a reverse proxy server
type Proxy struct {
	config    Config
	proxyAddr string // ProxyPort with colon prefix for listening
	targetURL string
	proxy     *httputil.ReverseProxy
	mux       *http.ServeMux
	server    *http.Server
	logger    *log.Logger
}

// New creates a new Proxy instance with the given configuration
func New(config Config) (*Proxy, error) {
	// Strip colon prefix if present for consistent internal storage
	if strings.HasPrefix(config.ProxyPort, ":") {
		config.ProxyPort = config.ProxyPort[1:]
	}

	// Set default target host if not provided
	if config.TargetHost == "" {
		config.TargetHost = "localhost"
	}

	// Set default logger if not provided
	if config.Logger == nil {
		config.Logger = log.Default()
	}

	targetURL := fmt.Sprintf("http://%s:%s", config.TargetHost, config.TargetPort)
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	p := &Proxy{
		config:    config,
		proxyAddr: ":" + config.ProxyPort,
		targetURL: targetURL,
		logger:    config.Logger,
	}

	p.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
		},
		ModifyResponse: p.modifyResponse,
	}

	p.mux = http.NewServeMux()
	p.setupHandlers()

	return p, nil
}

// setupHandlers configures the HTTP handlers
func (p *Proxy) setupHandlers() {
	// All requests go through the same handler
	p.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a WebSocket upgrade request
		if p.isWebSocketRequest(r) {
			p.handleWebSocket(w, r)
			return
		}
		p.proxy.ServeHTTP(w, r)
	})
}

// Start starts the proxy server
func (p *Proxy) Start() error {
	p.server = &http.Server{
		Addr:    p.proxyAddr,
		Handler: p.mux,
	}

	p.logger.Printf("Proxy server starting on %s -> %s", p.proxyAddr, p.targetURL)
	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}
	return nil
}

// Stop gracefully stops the proxy server
func (p *Proxy) Stop() error {
	if p.server != nil {
		p.logger.Println("Stopping proxy server...")
		return p.server.Close()
	}
	return nil
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func (p *Proxy) isWebSocketRequest(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

// handleWebSocket proxies WebSocket connections by hijacking the client
// connection and establishing a raw TCP tunnel to the target. The upgrade
// request is manually reconstructed to ensure correct Host and Origin headers
// so that the target's websocket.Handler (golang.org/x/net/websocket) accepts
// the handshake.
func (p *Proxy) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// p.logger.Printf("WebSocket: %s %s", r.RemoteAddr, r.URL.Path)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		p.logger.Printf("WebSocket hijack failed: %v", err)
		return
	}

	defer func() {
		if err := clientConn.Close(); err != nil {
			p.logger.Printf("WebSocket: client connection close failed: %v", err)
		}
	}()

	// Use net.JoinHostPort to properly handle both IPv4 and IPv6 addresses
	targetAddr := net.JoinHostPort(p.config.TargetHost, p.config.TargetPort)
	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		p.logger.Printf("WebSocket: target connect failed: %v", err)
		return
	}

	defer targetConn.Close()

	// Manually construct the upgrade request so we control every header.
	// r.Write() uses r.Host for the Host header which still points at the
	// proxy (e.g. localhost:8640). The target's websocket handler checks
	// Origin against Host, so both must reference the target.
	targetHost := fmt.Sprintf("%s:%s", p.config.TargetHost, p.config.TargetPort)

	var reqBuf bytes.Buffer
	fmt.Fprintf(&reqBuf, "%s %s HTTP/1.1\r\n", r.Method, r.URL.RequestURI())
	fmt.Fprintf(&reqBuf, "Host: %s\r\n", targetHost)

	// Copy all original headers except Host (already written)
	// and rewrite Origin to match the target so origin check passes.
	for key, vals := range r.Header {
		if strings.EqualFold(key, "Host") {
			continue
		}
		for _, val := range vals {
			if strings.EqualFold(key, "Origin") {
				// Rewrite origin to target so x/net/websocket checkOrigin passes
				val = strings.Replace(val, r.Host, targetHost, 1)
			}
			fmt.Fprintf(&reqBuf, "%s: %s\r\n", key, val)
		}
	}
	reqBuf.WriteString("\r\n")

	if _, err := targetConn.Write(reqBuf.Bytes()); err != nil {
		p.logger.Printf("WebSocket: forward request failed: %v", err)
		return
	}

	// Read upgrade response via buffered reader so any extra bytes
	// (first WS frame from target) stay in the buffer for the copy phase.
	targetBuf := bufio.NewReader(targetConn)
	resp, err := http.ReadResponse(targetBuf, r)
	if err != nil {
		p.logger.Printf("WebSocket: target response failed: %v", err)
		return
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		p.logger.Printf("WebSocket: target rejected upgrade (status %d)", resp.StatusCode)
		if err := resp.Write(clientConn); err != nil {
			p.logger.Printf("WebSocket: failed to write rejection response: %v", err)
		}
		return
	}

	// Write the 101 response back to the client
	if err := resp.Write(clientConn); err != nil {
		p.logger.Printf("WebSocket: client response failed: %v", err)
		return
	}
	if err := clientBuf.Flush(); err != nil {
		p.logger.Printf("WebSocket: client flush failed: %v", err)
		return
	}

	// Bidirectional streaming.
	// targetBuf (bufio.Reader) may hold bytes beyond the HTTP response.
	// clientBuf (bufio.ReadWriter) reader may hold bytes the client sent
	// while we were setting up the target connection.
	errChan := make(chan error, 2)

	// Client -> Target
	go func() {
		_, err := io.Copy(targetConn, clientBuf)
		errChan <- err
	}()

	// Target -> Client
	go func() {
		_, err := io.Copy(clientConn, targetBuf)
		errChan <- err
	}()

	// When one direction ends, close both so the other goroutine unblocks.
	<-errChan
	targetConn.Close()
	clientConn.Close()
	<-errChan
}

// modifyResponse modifies the response to rewrite URLs
func (p *Proxy) modifyResponse(resp *http.Response) error {
	contentType := resp.Header.Get("Content-Type")

	// Only modify HTML, CSS, and JavaScript responses
	if !p.shouldModifyContent(contentType) {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Rewrite URLs in the content
	modifiedBody := p.rewriteContent(body, contentType)

	resp.Body = io.NopCloser(bytes.NewReader(modifiedBody))
	resp.ContentLength = int64(len(modifiedBody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

	// Remove content encoding since we've decoded it
	resp.Header.Del("Content-Encoding")

	return nil
}

// shouldModifyContent determines if the content type should be modified
func (p *Proxy) shouldModifyContent(contentType string) bool {
	modifiableTypes := []string{
		"text/html",
		"text/css",
		"application/javascript",
		"text/javascript",
		"application/json",
	}

	for _, t := range modifiableTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// rewriteContent rewrites URLs in the content
func (p *Proxy) rewriteContent(body []byte, contentType string) []byte {
	content := string(body)

	// Patterns to replace
	replacements := []struct {
		pattern string
		replace string
	}{
		// Absolute URLs with port
		{fmt.Sprintf("http://localhost:%s", p.config.TargetPort), "http://localhost:" + p.config.ProxyPort},
		{fmt.Sprintf("https://localhost:%s", p.config.TargetPort), "https://localhost:" + p.config.ProxyPort},
		{fmt.Sprintf("//localhost:%s", p.config.TargetPort), "//localhost:" + p.config.ProxyPort},
		{fmt.Sprintf("localhost:%s", p.config.TargetPort), "localhost:" + p.config.ProxyPort},

		// WebSocket URLs
		{fmt.Sprintf("ws://localhost:%s", p.config.TargetPort), "ws://localhost:" + p.config.ProxyPort},
		{fmt.Sprintf("wss://localhost:%s", p.config.TargetPort), "wss://localhost:" + p.config.ProxyPort},

		// Port references in attributes
		{fmt.Sprintf(":%s/", p.config.TargetPort), ":" + p.config.ProxyPort + "/"},
		{fmt.Sprintf(":%s\"", p.config.TargetPort), ":" + p.config.ProxyPort + "\""},
		{fmt.Sprintf(":%s'", p.config.TargetPort), ":" + p.config.ProxyPort + "'"},
	}

	for _, r := range replacements {
		content = strings.ReplaceAll(content, r.pattern, r.replace)
	}

	// Handle dynamic port references in JavaScript
	if strings.Contains(contentType, "javascript") || strings.Contains(contentType, "html") {
		// Replace port numbers in window.location style patterns
		portRegex := regexp.MustCompile(fmt.Sprintf(`(['"])%s(['"])`, p.config.TargetPort))
		proxyPortNum := strings.TrimPrefix(p.config.ProxyPort, ":")
		content = portRegex.ReplaceAllString(content, "${1}"+proxyPortNum+"${2}")
	}

	// Inject WebSocket debug logging into the g-sui WS client code
	if strings.Contains(contentType, "html") {
		// Add console.log to ws.onopen
		content = strings.Replace(content,
			`ws.onopen = function(){`,
			`ws.onopen = function(){ console.log('[PROXY-WS] onopen', ws.url, 'readyState:', ws.readyState);`,
			1)

		// Add console.log to ws.onmessage
		content = strings.Replace(content,
			`ws.onmessage = function(ev){`,
			`ws.onmessage = function(ev){ console.log('[PROXY-WS] onmessage', ev.data.substring(0, 200));`,
			1)

		// Add console.log to ws.onclose
		content = strings.Replace(content,
			`ws.onclose = function(){`,
			`ws.onclose = function(){ console.log('[PROXY-WS] onclose');`,
			1)

		// Add console.log to ws.onerror
		content = strings.Replace(content,
			`ws.onerror = function(){`,
			`ws.onerror = function(){ console.log('[PROXY-WS] onerror');`,
			1)

		// Add console.log to ws.send
		content = strings.Replace(content,
			`ws.send(JSON.stringify(msg));`,
			`console.log('[PROXY-WS] send call:', JSON.stringify(msg).substring(0, 200)); ws.send(JSON.stringify(msg));`,
			1)

		// Log ping sends
		content = strings.Replace(content,
			`ws.send(JSON.stringify({ type: 'ping', t: Date.now() }));`,
			`console.log('[PROXY-WS] send ping'); ws.send(JSON.stringify({ type: 'ping', t: Date.now() }));`,
			1)
	}

	return []byte(content)
}

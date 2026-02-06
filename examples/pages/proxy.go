package pages

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/michalCapo/g-sui/proxy"
	"github.com/michalCapo/g-sui/ui"
)

var (
	proxyTarget       = ui.Target()
	proxyConfigTarget = ui.Target()
	proxyStatus       = "stopped"
	proxyMutex        sync.Mutex
	proxyServer       *proxy.Proxy
)

type ProxyConfig struct {
	ProxyPort  string
	TargetPort string
	TargetHost string
}

func Proxy(ctx *ui.Context) string {
	config := ProxyConfig{
		ProxyPort:  "1423",
		TargetPort: "1422",
		TargetHost: "localhost",
	}

	// Restore config from request if available
	ctx.Body(&config)

	return ui.Div("flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Proxy Example"),
		ui.Div("text-gray-600 dark:text-gray-400")(
			"This example demonstrates how to use the reverse proxy server with WebSocket support. ",
			"The proxy forwards HTTP and WebSocket requests to a target server while rewriting port references in responses.",
		),

		ui.Card().Header("<h3 class='font-bold'>Proxy Configuration</h3>").Body(
			renderProxyConfig(ctx, &config),
		).Render(),

		ui.Card().Header("<h3 class='font-bold'>Proxy Status</h3>").Body(
			ui.Div("", proxyTarget)(renderProxyStatus(ctx)),
		).Render(),

		ui.Card().Header("<h3 class='font-bold'>Usage Instructions</h3>").Body(
			renderUsageInstructions(),
		).Render(),
	)
}

func renderProxyConfig(ctx *ui.Context, config *ProxyConfig) string {
	return ui.Form("flex flex-col gap-4", proxyConfigTarget, ctx.Submit(UpdateProxyConfig).Replace(proxyConfigTarget))(
		ui.Div("grid grid-cols-1 md:grid-cols-3 gap-4")(
			ui.IText("ProxyPort", config).Placeholder("1423").Render("Proxy Port"),
			ui.IText("TargetPort", config).Placeholder("1422").Render("Target Port"),
			ui.IText("TargetHost", config).Placeholder("localhost").Render("Target Host"),
		),
		ui.Div("flex gap-2")(
			ui.Button().Submit().Color(ui.Blue).Class("rounded").Render("Update Config"),
		),
	)
}

func UpdateProxyConfig(ctx *ui.Context) string {
	config := ProxyConfig{}
	if err := ctx.Body(&config); err != nil {
		ctx.Error("Failed to parse configuration: " + err.Error())
		return renderProxyConfig(ctx, &config)
	}

	ctx.Success(fmt.Sprintf("Configuration updated: Proxy %s -> %s:%s", config.ProxyPort, config.TargetHost, config.TargetPort))
	return renderProxyConfig(ctx, &config)
}

func renderProxyStatus(ctx *ui.Context) string {
	proxyMutex.Lock()
	status := proxyStatus
	proxyMutex.Unlock()

	var statusColor string
	var statusText string
	switch status {
	case "running":
		statusColor = "text-green-600 dark:text-green-400"
		statusText = "● Running"
	case "stopped":
		statusColor = "text-gray-600 dark:text-gray-400"
		statusText = "○ Stopped"
	case "starting":
		statusColor = "text-yellow-600 dark:text-yellow-400"
		statusText = "◐ Starting..."
	case "stopping":
		statusColor = "text-yellow-600 dark:text-yellow-400"
		statusText = "◐ Stopping..."
	default:
		statusColor = "text-red-600 dark:text-red-400"
		statusText = "✗ Error"
	}

	return ui.Div("flex flex-col gap-4")(
		ui.Div("flex items-center gap-2")(
			ui.Div("text-lg font-semibold "+statusColor)(statusText),
		),
		ui.Div("flex gap-2")(
			ui.Button().
				Color(ui.Green).
				Class("rounded").
				Click(ctx.Call(StartProxy).Replace(proxyTarget)).
				Render("Start Proxy"),
			ui.Button().
				Color(ui.Red).
				Class("rounded").
				Click(ctx.Call(StopProxy).Replace(proxyTarget)).
				Render("Stop Proxy"),
		),
	)
}

func renderProxyStatusWithTarget(ctx *ui.Context) string {
	return ui.Div("", proxyTarget)(renderProxyStatus(ctx))
}

func StartProxy(ctx *ui.Context) string {
	proxyMutex.Lock()
	status := proxyStatus
	proxyMutex.Unlock()

	if status == "running" {
		ctx.Info("Proxy is already running")
		return renderProxyStatusWithTarget(ctx)
	}

	config := ProxyConfig{
		ProxyPort:  "1423",
		TargetPort: "1422",
		TargetHost: "localhost",
	}
	ctx.Body(&config)

	proxyMutex.Lock()
	proxyStatus = "starting"
	proxyMutex.Unlock()

	ctx.Info("Starting proxy server...")

	// Start proxy in a goroutine
	go func() {
		proxyConfig := proxy.Config{
			ProxyPort:  config.ProxyPort,
			TargetPort: config.TargetPort,
			TargetHost: config.TargetHost,
			Logger:     log.New(os.Stdout, "[PROXY] ", log.LstdFlags),
		}

		var err error
		proxyServer, err = proxy.New(proxyConfig)
		if err != nil {
			proxyMutex.Lock()
			proxyStatus = "error"
			proxyMutex.Unlock()
			log.Printf("Failed to create proxy: %v", err)
			return
		}

		proxyMutex.Lock()
		proxyStatus = "running"
		proxyMutex.Unlock()

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			proxyMutex.Lock()
			proxyStatus = "stopping"
			proxyMutex.Unlock()
			if err := proxyServer.Stop(); err != nil {
				log.Printf("Error stopping proxy: %v", err)
			}
			proxyMutex.Lock()
			proxyStatus = "stopped"
			proxyMutex.Unlock()
		}()

		if err := proxyServer.Start(); err != nil {
			proxyMutex.Lock()
			proxyStatus = "error"
			proxyMutex.Unlock()
			log.Printf("Proxy server error: %v", err)
		} else {
			// Server stopped gracefully
			proxyMutex.Lock()
			if proxyStatus == "running" {
				proxyStatus = "stopped"
			}
			proxyMutex.Unlock()
		}
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	ctx.Success(fmt.Sprintf("Proxy server starting on port %s -> %s:%s", config.ProxyPort, config.TargetHost, config.TargetPort))
	return renderProxyStatusWithTarget(ctx)
}

func StopProxy(ctx *ui.Context) string {
	proxyMutex.Lock()
	status := proxyStatus
	server := proxyServer
	proxyMutex.Unlock()

	if status != "running" {
		ctx.Info("Proxy is not running (status: " + status + ")")
		return renderProxyStatusWithTarget(ctx)
	}

	if server == nil {
		ctx.Error("Proxy server reference is nil")
		proxyMutex.Lock()
		proxyStatus = "stopped"
		proxyMutex.Unlock()
		return renderProxyStatusWithTarget(ctx)
	}

	proxyMutex.Lock()
	proxyStatus = "stopping"
	proxyMutex.Unlock()

	if err := server.Stop(); err != nil {
		ctx.Error("Failed to stop proxy: " + err.Error())
		proxyMutex.Lock()
		proxyStatus = "error"
		proxyMutex.Unlock()
		return renderProxyStatusWithTarget(ctx)
	}

	proxyMutex.Lock()
	proxyStatus = "stopped"
	proxyServer = nil
	proxyMutex.Unlock()

	ctx.Success("Proxy server stopped")
	return renderProxyStatusWithTarget(ctx)
}

func renderUsageInstructions() string {
	return ui.Div("flex flex-col gap-4 text-sm")(
		ui.Div("")(
			ui.Div("font-bold mb-2")("1. Start the Example Application"),
			ui.Div("text-gray-600 dark:text-gray-400")(
				"The example application should be running on port 1422. ",
				"If not, start it with: ",
				`<code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">go run examples/main.go</code>`,
			),
		),
		ui.Div("")(
			ui.Div("font-bold mb-2")("2. Configure Proxy Settings"),
			ui.Div("text-gray-600 dark:text-gray-400")(
				"Set the proxy port (e.g., 1423) and target port (e.g., 1422). ",
				"The proxy will forward requests from the proxy port to the target port.",
			),
		),
		ui.Div("")(
			ui.Div("font-bold mb-2")("3. Start the Proxy"),
			ui.Div("text-gray-600 dark:text-gray-400")(
				"Click the 'Start Proxy' button to start the reverse proxy server. ",
				"The proxy will rewrite URLs in responses to maintain transparent proxying.",
			),
		),
		ui.Div("")(
			ui.Div("font-bold mb-2")("4. Access Through Proxy"),
			ui.Div("text-gray-600 dark:text-gray-400")(
				"Access the application through the proxy port: ",
				`<code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">http://localhost:1423</code>`,
				". The proxy will forward all requests (including WebSocket connections) to the target server.",
			),
		),
		ui.Div("")(
			ui.Div("font-bold mb-2")("Features"),
			ui.Div("text-gray-600 dark:text-gray-400")(
				`<ul class="list-disc list-inside space-y-1">`,
				`<li>HTTP request forwarding</li>`,
				`<li>WebSocket connection proxying</li>`,
				`<li>Automatic URL rewriting in HTML, CSS, and JavaScript</li>`,
				`<li>Port reference rewriting in responses</li>`,
				`<li>WebSocket debug logging (injected into HTML responses)</li>`,
				`</ul>`,
			),
		),
	)
}

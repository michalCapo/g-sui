package pages

import r "github.com/michalCapo/g-sui/ui"

func Others(ctx *r.Context) *r.Node {
	hello := r.Div("bg-white p-6 rounded-lg shadow w-full").Render(
		r.Div("text-lg font-bold mb-2").Text("Hello"),
		Hello(ctx),
	)

	counter := r.Div("bg-white p-6 rounded-lg shadow w-full").Render(
		r.Div("text-lg font-bold mb-2").Text("Counter"),
		Counter(ctx),
	)

	mdSample := `# Markdown Example

This is rendered using the built-in ` + "`r.Markdown`" + ` helper which converts **markdown** to HTML.

## Features

- Server-rendered Go UI
- Component-based architecture
- WebSocket-driven interactivity

> All HTML is generated server-side and delivered as pure JavaScript.

### Code example

Inline ` + "`code`" + ` and **bold** with *italic* text.

1. First ordered item
2. Second ordered item
3. Third ordered item
`

	markdown := r.Div("bg-white p-6 rounded-lg shadow flex flex-col gap-3 w-full").Render(
		r.Div("text-xl font-bold").Text("Markdown"),
		r.Div("text-sm text-gray-500 mb-2").Text("Demonstrates rendering markdown content using r.Markdown()."),
		r.Markdown("prose prose-sm max-w-none", mdSample),
	)

	return r.Div("max-w-6xl mx-auto flex flex-col gap-6 w-full").Render(
		r.Div("text-3xl font-bold").Text("Others"),
		r.Div("text-gray-600").Text("Miscellaneous demos: Hello, Counter, Login, and Markdown."),
		hello,
		counter,
		markdown,
	)
}

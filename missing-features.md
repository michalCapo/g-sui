# Missing Features Compared to t-sui

This document lists features present in `t-sui` (TypeScript implementation) that are not implemented in `g-sui` (Go implementation), based on the current code and READMEs of both projects.

- WebSocket patching: `t-sui` provides a WS server at `/__ws` to push server-initiated DOM patches; `g-sui` implements WS only for dev autoreload at `/live` and does not support server-driven patches.
- Deferred fragments: `t-sui` supports rendering skeletons and replacing them asynchronously via WS (`ctx.Patch(...)`, skeleton helpers); `g-sui` has no deferred/skeleton swap mechanism.
- Server-driven patches: `t-sui` has `ctx.Patch(target, html[, clear])` to push updates at arbitrary times; `g-sui` has no equivalent.
- Invalid-target handling: `t-sui` notifies the server when a patch target ID is missing in the DOM and invokes the optional `clear()` callback; `g-sui` has no such mechanism.
- Session model: `t-sui` manages per-client sessions via `tsui__sid` cookie, tracks last-seen, and prunes inactive sessions; `g-sui` sets a `session_id` cookie but does not implement WS session tracking or pruning.
- Heartbeats: `t-sui` implements WS ping/pong heartbeats and client pings to keep sessions alive; `g-sui` has no WS heartbeat system.
- Dev error page: On request errors, `t-sui` serves a minimal fallback page that auto-reloads on WS reconnect; `g-sui` recovers panics to a toast but has no WS-driven error-reload page.

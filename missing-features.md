# Missing Features Compared to t-sui

This document lists features present in `t-sui` (TypeScript implementation) that are not implemented in `g-sui` (Go implementation), based on the current code and READMEs of both projects.

- Deferred fragments: `t-sui` supports rendering skeletons and replacing them asynchronously via WS (`ctx.Patch(...)`, skeleton helpers); `g-sui` has no deferred/skeleton swap mechanism.
- Invalid-target handling: `t-sui` notifies the server when a patch target ID is missing in the DOM and invokes the optional `clear()` callback; `g-sui` has no such mechanism.

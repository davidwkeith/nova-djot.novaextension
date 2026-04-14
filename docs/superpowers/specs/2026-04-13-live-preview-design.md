# Djot Live Preview Design Spec

**Date**: 2026-04-13
**Status**: Approved

## Overview

Add a live preview to the Nova Djot extension that renders Djot to HTML in a split pane, updated in real-time via WebSocket. Includes editor→preview scroll sync.

## Architecture

The existing `djot-lsp` binary is extended with:

1. **HTTP server** — starts on a random localhost port at LSP initialization. Serves a self-contained HTML page at `/preview`.
2. **WebSocket endpoint** — at `/ws`. Pushes rendered HTML on every `didChange`. Receives scroll commands from the Nova client.

Port is communicated to Nova via a custom LSP notification `djot/previewReady` with `{ "port": N }`.

### Data Flow

```
User types → Nova didChange → djot-lsp parses + renders HTML → WebSocket push → preview tab updates
User moves cursor → Nova onDidChangeSelection → WebSocket {"type":"scroll","line":N} → preview scrolls
User runs "Preview Djot" → Nova opens http://localhost:{port}/preview
```

## Go Server Changes

### New File: `djot-lsp/server/preview.go`

Responsibilities:
- Start `net/http` server on `localhost:0` (random port)
- Serve HTML shell page at `GET /preview`
- WebSocket endpoint at `/ws` using `gorilla/websocket` or `nhooyr.io/websocket`
- Track connected WebSocket clients
- `renderAndBroadcast(doc *Document)` — renders Djot→HTML via `djot_html.ConvertDjot()`, injects `data-line` attributes, sends to all clients
- `handleScrollMessage(line int)` — forwarded from client (no-op on server, just relayed to other potential clients)

### HTML Rendering

Use godjot's `djot_html.New().ConvertDjot()` to render the document AST to HTML.

Wrap the default renderer with custom conversion rules that inject `data-line="N"` attributes on block-level elements. The line number is derived from finding the element's source text in the document and converting the byte offset to a line number using the document's `OffsetToPosition()`.

Block elements that get `data-line`:
- Paragraphs, headings, code blocks, raw blocks, block quotes, lists, list items, tables, divs, thematic breaks, footnote definitions

### Server Lifecycle

- Started in `handleInitialize` — spawns HTTP server goroutine
- Port communicated via `djot/previewReady` notification sent after server starts
- On `didChange`: re-render and broadcast
- On `didClose`: clear preview content
- On shutdown: HTTP server stopped via `context.Context` cancellation

### WebSocket Message Protocol

**Server → Client (content update):**
```json
{"type": "content", "html": "<h1>Hello</h1><p>World</p>"}
```

**Client → Server (scroll sync):**
```json
{"type": "scroll", "line": 42}
```

## Preview HTML Page

Self-contained HTML served at `/preview`. Includes embedded CSS and JS.

### Stylesheet

Minimal, readable defaults:
- System font stack (`-apple-system, BlinkMacSystemFont, ...`)
- Sensible heading sizes
- Code blocks with background color and monospace font
- Bordered tables with padding
- Block quote left border
- Max-width content area for readability

### JavaScript

- Connects to `ws://localhost:{port}/ws` on load
- Auto-reconnects on disconnect (handles LSP restart)
- On `content` message: replaces `#content` innerHTML, preserves scroll position
- On cursor sync: finds element with matching `data-line` attribute, calls `scrollIntoView({ behavior: 'smooth', block: 'center' })`

### Scroll Position Preservation

On content update:
1. Record current `scrollTop`
2. Replace innerHTML
3. Restore `scrollTop`

This prevents the preview from jumping to the top on every keystroke.

## Scroll Sync (Editor → Preview)

### Nova JS Side

`Scripts/main.js` adds:

- Listen on `editor.onDidChangeSelection()` for the active Djot editor
- Extract current cursor line number
- Send `{"type": "scroll", "line": N}` to a WebSocket connection to `ws://localhost:{port}/ws`
- Debounced (e.g., 100ms) to avoid flooding during rapid cursor movement

### Preview Page Side

On receiving a `scroll` message:
- Find the closest element with `data-line` attribute where `data-line <= targetLine`
- Call `element.scrollIntoView({ behavior: 'smooth', block: 'center' })`

### `data-line` Injection

Block-level HTML elements rendered by godjot get a `data-line="N"` attribute where N is the 0-indexed source line number of that element. This is done by wrapping godjot's HTML conversion with custom rules that look up each node's position in the source.

Since godjot nodes don't carry position information, the renderer finds positions by searching the source text for the node's content (same strategy used in `document.go` for the LSP features).

## Nova Extension Changes

### `Scripts/main.js`

Additions to `activate()`:
- Listen for `djot/previewReady` notification, store port
- Register command `io.dwk.djot.preview` that opens `http://localhost:{port}/preview`
- Open a WebSocket connection to the server for sending scroll sync messages
- Hook `onDidChangeSelection` on active Djot editors to send scroll line

### `extension.json`

Add command definition:
```json
"commands": {
    "editor": [
        {
            "title": "Preview Djot",
            "command": "io.dwk.djot.preview",
            "when": "editorSyntax == 'djot'"
        }
    ]
}
```

## Build

Recompile `bin/djot-lsp` with the new preview code. May need a new Go dependency for WebSocket support (`nhooyr.io/websocket` or `gorilla/websocket`).

`make lsp` rebuilds the binary. No new binaries or build targets needed.

## What's NOT Changing

- Tree-sitter grammar, syntax XML, query files — unchanged
- Existing LSP features (completions, diagnostics, symbols, hover, definition) — unchanged
- Version remains 1.0.0

## Out of Scope

- Preview → editor scroll sync (clicking preview doesn't move cursor)
- Custom CSS / user theming
- Export to HTML file
- Multi-document preview (one active preview per workspace)
- Print / PDF export

## Dependencies

| Dependency | License | Purpose |
|---|---|---|
| `nhooyr.io/websocket` or `gorilla/websocket` | MIT/BSD | WebSocket server in Go |

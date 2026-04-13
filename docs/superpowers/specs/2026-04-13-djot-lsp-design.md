# Djot LSP Server Design Spec

**Date**: 2026-04-13
**Status**: Approved

## Overview

Add a Go-based Language Server Protocol (LSP) server to the Nova Djot extension, providing completions, diagnostics, document symbols, hover, and go-to-definition for `.dj` files. The server uses `sivukhin/godjot/v2` for parsing and communicates with Nova via stdio JSON-RPC.

## Architecture

```
Nova Editor
  ├── Scripts/main.js          (LanguageClient — starts/stops LSP)
  └── bin/djot-lsp             (Go binary — LSP server via stdio)
        └── uses: sivukhin/godjot/v2 (Djot parser)
```

- Nova spawns `bin/djot-lsp` as a child process when a `.dj` file is opened
- Communication: stdio (stdin/stdout JSON-RPC), Nova's default `LanguageClient` transport
- Re-parses the full document on every `textDocument/didOpen` and `textDocument/didChange`
- Single-document scope — no workspace/multi-file support

## LSP Server — Go Project Structure

```
djot-lsp/
├── go.mod                    # module github.com/dwk/djot-lsp
├── go.sum
├── main.go                   # Entry point — stdio JSON-RPC server
├── server/
│   ├── server.go             # LSP method handlers (initialize, didOpen, didChange, etc.)
│   ├── document.go           # Per-document state (source text, AST, indexes)
│   ├── completions.go        # textDocument/completion handler
│   ├── diagnostics.go        # Publish diagnostics on change
│   ├── symbols.go            # textDocument/documentSymbol handler
│   ├── hover.go              # textDocument/hover handler
│   └── definition.go         # textDocument/definition handler
└── server/
    └── *_test.go             # Tests for each handler
```

### Dependencies

- `github.com/sivukhin/godjot/v2` — Djot parser (MIT license)
- LSP protocol library — `go.lsp.dev/protocol` and `go.lsp.dev/jsonrpc2` or equivalent

### Document State Model

Each open document gets a `Document` struct:

- `uri` — document URI
- `source` — raw source text (`[]byte`)
- `lineOffsets` — newline byte offset index (for offset → line/column conversion)
- `ast` — parsed AST (`[]TreeNode[DjotNode]`)
- `headings` — index: label → position
- `footnoteDefs` — index: label → position
- `footnoteRefs` — index: label → []position
- `referenceDefs` — index: label → position
- `referenceRefs` — index: label → []position

Rebuilt on every `didChange`. godjot's `BuildDjotContext()` provides reference/footnote lookup maps; byte offsets from tokens are converted to LSP `{line, character}` positions using the newline index.

## LSP Features

### Completions (`textDocument/completion`)

| Trigger | Items Offered | Source |
|---|---|---|
| `[^` typed | Footnote labels (e.g., `[^note]`) | All `FootnoteDefNode` labels in document |
| `[text][` typed | Reference link labels (e.g., `[text][ref]`) | All `ReferenceDefNode` labels in document |
| `#` inside `{}` attribute | Heading IDs | All `HeadingNode` auto-generated IDs |

Completion items include the closing bracket/brace.

### Diagnostics (`textDocument/publishDiagnostics`)

Published automatically after every document change.

| Diagnostic | Severity | Condition |
|---|---|---|
| Undefined footnote reference | Warning | `[^label]` used but no `[^label]:` definition exists |
| Undefined link reference | Warning | `[text][ref]` used but no `[ref]:` definition exists |
| Duplicate footnote definition | Warning | Two `[^label]:` definitions with the same label |
| Duplicate reference definition | Warning | Two `[ref]:` definitions with the same label |

### Document Symbols (`textDocument/documentSymbol`)

Returns a nested tree of headings:

- `HeadingNode` → `DocumentSymbol` with `kind: String`, `detail: "H{level}"`
- `SectionNode` provides nesting hierarchy (H2 under H1, H3 under H2, etc.)
- `name` is the heading's full text content

### Hover (`textDocument/hover`)

| Hover Target | Content Shown |
|---|---|
| Footnote reference `[^note]` | The footnote's full text content |
| Link reference `[text][ref]` | The link destination URL |
| Inline link `[text](url)` | The URL |

Returns Markdown-formatted hover content.

### Go to Definition (`textDocument/definition`)

| Source | Jumps To |
|---|---|
| Footnote reference `[^note]` | The `[^note]:` definition position |
| Link reference `[text][ref]` | The `[ref]:` definition position |

## Nova Extension JavaScript

### `Scripts/main.js`

Minimal — manages LSP lifecycle only:

- `exports.activate` — creates `LanguageClient` pointing to `bin/djot-lsp`, transport stdio, syntax filter `["djot"]`, calls `start()`
- `exports.deactivate` — calls `stop()` on the client

### `extension.json` Updates

Added fields (on top of existing v1.0.0 manifest):

- `"main": "Scripts/main.js"` — extension entry point
- `"activationEvents": ["onLanguage:djot"]` — only load when Djot file opened
- `"entitlements": { "process": true }` — required to spawn the LSP binary

Version remains `1.0.0`.

## Build Process

### LSP Binary Compilation

Added to the existing `Makefile`:

```makefile
LSP_DIR    = djot-lsp
LSP_BINARY = bin/djot-lsp

lsp: $(LSP_BINARY)

$(LSP_BINARY): $(LSP_DIR)/main.go $(LSP_DIR)/server/*.go
	cd $(LSP_DIR) && GOOS=darwin GOARCH=arm64 go build -o /tmp/djot-lsp-arm64 .
	cd $(LSP_DIR) && GOOS=darwin GOARCH=amd64 go build -o /tmp/djot-lsp-amd64 .
	lipo -create /tmp/djot-lsp-arm64 /tmp/djot-lsp-amd64 -output $(LSP_BINARY)
	rm -f /tmp/djot-lsp-arm64 /tmp/djot-lsp-amd64

all: build lsp

clean:
	rm -rf $(GRAMMAR_DIR)
	rm -f $(DYLIB) $(LSP_BINARY)
```

### Build Dependencies

- Go 1.23+
- Xcode Command Line Tools (`lipo`)

### Git Strategy

- `djot-lsp/` source directory — **committed** (our code)
- `bin/djot-lsp` compiled binary — **committed** (Nova extensions are static bundles)
- `tree-sitter-djot/` — remains gitignored (third-party clone)

## LSP Capabilities Declared

The server declares these capabilities in `initialize` response:

```json
{
    "capabilities": {
        "textDocumentSync": {
            "openClose": true,
            "change": 1
        },
        "completionProvider": {
            "triggerCharacters": ["[", "#"]
        },
        "hoverProvider": true,
        "definitionProvider": true,
        "documentSymbolProvider": true
    }
}
```

`change: 1` = full document sync (send entire text on every change). Simpler than incremental sync and fast enough for single documents.

## What's NOT Changing

- Tree-sitter grammar, Djot.xml syntax definition, all query files (highlights, folds, symbols, injections) — unchanged
- Extension version remains 1.0.0

## Out of Scope

- Multi-file / workspace support
- Live preview / HTML export
- Code actions or formatting
- Rename symbol
- Incremental document sync (full sync is sufficient)

## Dependencies

| Dependency | Version | License | Purpose |
|---|---|---|---|
| `sivukhin/godjot/v2` | latest | MIT | Djot parser |
| `go.lsp.dev/protocol` | latest | BSD-3 | LSP type definitions |
| `go.lsp.dev/jsonrpc2` | latest | BSD-3 | JSON-RPC transport |

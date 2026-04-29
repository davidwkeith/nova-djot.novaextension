package server

import (
	"strings"
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var documents sync.Map
var workspaceRoot string

func NewHandler() *protocol.Handler {
	return &protocol.Handler{
		Initialize:                 handleInitialize,
		Initialized:               handleInitialized,
		Shutdown:                   handleShutdown,
		TextDocumentDidOpen:        handleDidOpen,
		TextDocumentDidChange:      handleDidChange,
		TextDocumentDidClose:       handleDidClose,
		TextDocumentDocumentSymbol: handleDocumentSymbol,
		TextDocumentCompletion:     handleCompletion,
		TextDocumentHover:          handleHover,
		TextDocumentDefinition:     handleDefinition,
	}
}

func handleInitialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	if params.RootURI != nil {
		workspaceRoot = strings.TrimPrefix(string(*params.RootURI), "file://")
	} else if params.RootPath != nil {
		workspaceRoot = *params.RootPath
	}

	syncKind := protocol.TextDocumentSyncKindFull
	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: boolPtr(true),
				Change:    &syncKind,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"[", "#"},
			},
			HoverProvider:          true,
			DefinitionProvider:     true,
			DocumentSymbolProvider: true,
		},
	}, nil
}

func handleInitialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	if workspaceRoot == "" {
		return nil
	}
	port, err := StartPreviewServer(workspaceRoot)
	if err != nil {
		return nil
	}
	ctx.Notify("djot/previewServer", map[string]interface{}{
		"port": port,
	})
	return nil
}

func handleShutdown(ctx *glsp.Context) error {
	StopPreviewServer()
	return nil
}

func handleDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	doc := NewDocument(params.TextDocument.URI, params.TextDocument.Text)
	documents.Store(params.TextDocument.URI, doc)
	publishDiagnostics(ctx, doc)
	return nil
}

func handleDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}
	last := params.ContentChanges[len(params.ContentChanges)-1]
	whole, ok := last.(protocol.TextDocumentContentChangeEventWhole)
	if !ok {
		return nil
	}
	doc := NewDocument(params.TextDocument.URI, whole.Text)
	documents.Store(params.TextDocument.URI, doc)
	publishDiagnostics(ctx, doc)
	return nil
}

func handleDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	documents.Delete(params.TextDocument.URI)
	ctx.Notify(string(protocol.ServerTextDocumentPublishDiagnostics), protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []protocol.Diagnostic{},
	})
	return nil
}

func getDocument(uri string) *Document {
	val, ok := documents.Load(uri)
	if !ok {
		return nil
	}
	return val.(*Document)
}

func boolPtr(b bool) *bool {
	return &b
}

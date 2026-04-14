package server

import (
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var documents sync.Map // uri string → *Document

func NewHandler() *protocol.Handler {
	handler := &protocol.Handler{
		Initialize:            handleInitialize,
		Initialized:           handleInitialized,
		Shutdown:              handleShutdown,
		TextDocumentDidOpen:          handleDidOpen,
		TextDocumentDidChange:        handleDidChange,
		TextDocumentDidClose:         handleDidClose,
		TextDocumentDocumentSymbol:   handleDocumentSymbol,
		TextDocumentCompletion:        handleCompletion,
		TextDocumentHover:             handleHover,
	}
	return handler
}

func handleInitialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
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
	return nil
}

func handleShutdown(ctx *glsp.Context) error {
	return nil
}

func handleDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	doc := NewDocument(params.TextDocument.URI, params.TextDocument.Text)
	documents.Store(params.TextDocument.URI, doc)
	return publishDiagnostics(ctx, doc)
}

func handleDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}
	// Full sync — last content change has the full text
	last := params.ContentChanges[len(params.ContentChanges)-1]
	whole, ok := last.(protocol.TextDocumentContentChangeEventWhole)
	if !ok {
		return nil
	}
	doc := NewDocument(params.TextDocument.URI, whole.Text)
	documents.Store(params.TextDocument.URI, doc)
	return publishDiagnostics(ctx, doc)
}

func handleDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	documents.Delete(params.TextDocument.URI)
	// Clear diagnostics
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

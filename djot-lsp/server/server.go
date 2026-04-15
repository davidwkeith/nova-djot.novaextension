package server

import (
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var documents sync.Map // uri string → *Document

func NewHandler() *protocol.Handler {
	handler := &protocol.Handler{
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
	CleanupPreviewFile()
	return nil
}

func handleDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	doc := NewDocument(params.TextDocument.URI, params.TextDocument.Text)
	documents.Store(params.TextDocument.URI, doc)
	publishDiagnostics(ctx, doc)
	notifyPreview(ctx, doc)
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
	notifyPreview(ctx, doc)
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

// notifyPreview writes the rendered preview file and notifies the client of its path.
func notifyPreview(ctx *glsp.Context, doc *Document) {
	path, err := WritePreviewFile(doc)
	if err != nil {
		return
	}
	ctx.Notify("djot/previewFile", map[string]interface{}{
		"path": path,
	})
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

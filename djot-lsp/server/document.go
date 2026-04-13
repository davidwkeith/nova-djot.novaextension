package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Document struct {
	URI    string
	Source string
}

func NewDocument(uri string, source string) *Document {
	return &Document{URI: uri, Source: source}
}

func publishDiagnostics(ctx *glsp.Context, doc *Document) error {
	ctx.Notify(string(protocol.ServerTextDocumentPublishDiagnostics), protocol.PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: []protocol.Diagnostic{},
	})
	return nil
}

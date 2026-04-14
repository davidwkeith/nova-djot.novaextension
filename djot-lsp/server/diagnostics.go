package server

import (
	"fmt"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// computeDiagnostics returns all diagnostics for the given document:
//   - undefined footnote references
//   - undefined link references (checked against ReferenceDefs and Heading IDs)
//   - duplicate definitions (already collected in DuplicateDiagnostics during parsing)
func computeDiagnostics(doc *Document) []protocol.Diagnostic {
	var diags []protocol.Diagnostic

	// 1. Undefined footnote references
	for _, ref := range doc.FootnoteRefs {
		if _, ok := doc.FootnoteDefs[ref.Label]; !ok {
			diags = append(diags, protocol.Diagnostic{
				Range:    ref.Range,
				Severity: severityPtr(protocol.DiagnosticSeverityWarning),
				Source:   strPtr("djot-lsp"),
				Message:  fmt.Sprintf("Undefined footnote reference: %q", ref.Label),
			})
		}
	}

	// 2. Undefined link references
	for _, ref := range doc.ReferenceRefs {
		if _, ok := doc.ReferenceDefs[ref.Label]; ok {
			continue
		}
		// Also check heading IDs
		resolvedByHeading := false
		for _, h := range doc.Headings {
			if h.ID == ref.Label {
				resolvedByHeading = true
				break
			}
		}
		if !resolvedByHeading {
			diags = append(diags, protocol.Diagnostic{
				Range:    ref.Range,
				Severity: severityPtr(protocol.DiagnosticSeverityWarning),
				Source:   strPtr("djot-lsp"),
				Message:  fmt.Sprintf("Undefined link reference: %q", ref.Label),
			})
		}
	}

	// 3. Duplicate definitions (collected during index building)
	diags = append(diags, doc.DuplicateDiagnostics...)

	return diags
}

func strPtr(s string) *string {
	return &s
}

func publishDiagnostics(ctx *glsp.Context, doc *Document) error {
	diags := computeDiagnostics(doc)
	if diags == nil {
		diags = []protocol.Diagnostic{}
	}
	ctx.Notify(string(protocol.ServerTextDocumentPublishDiagnostics), protocol.PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diags,
	})
	return nil
}

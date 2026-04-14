package server

import (
	"fmt"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func handleDocumentSymbol(ctx *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	doc := getDocument(params.TextDocument.URI)
	if doc == nil {
		return []protocol.DocumentSymbol{}, nil
	}
	return computeDocumentSymbols(doc), nil
}

// computeDocumentSymbols builds a nested heading tree from the document's headings.
// Headings are nested using a stack: an H2 after an H1 becomes a child of the H1;
// an H2 after another H2 is a sibling (the stack is popped to find the parent).
func computeDocumentSymbols(doc *Document) []protocol.DocumentSymbol {
	type stackEntry struct {
		symbol *protocol.DocumentSymbol
		level  int
	}

	var roots []protocol.DocumentSymbol
	var stack []stackEntry

	for i := range doc.Headings {
		h := doc.Headings[i]
		detail := fmt.Sprintf("H%d", h.Level)
		sym := protocol.DocumentSymbol{
			Name:           h.Label,
			Detail:         &detail,
			Kind:           protocol.SymbolKindString,
			Range:          h.Range,
			SelectionRange: h.Range,
		}

		// Pop stack until we find an entry with a lower level (a proper ancestor).
		for len(stack) > 0 && stack[len(stack)-1].level >= h.Level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			// No parent — this is a root symbol.
			roots = append(roots, sym)
			// Push a pointer to the just-appended root element.
			stack = append(stack, stackEntry{symbol: &roots[len(roots)-1], level: h.Level})
		} else {
			// Attach as a child of the top of the stack.
			parent := stack[len(stack)-1].symbol
			parent.Children = append(parent.Children, sym)
			// Push a pointer to the just-appended child.
			stack = append(stack, stackEntry{
				symbol: &parent.Children[len(parent.Children)-1],
				level:  h.Level,
			})
		}
	}

	return roots
}

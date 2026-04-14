package server

import (
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func handleCompletion(ctx *glsp.Context, params *protocol.CompletionParams) (any, error) {
	doc := getDocument(params.TextDocument.URI)
	if doc == nil {
		return []protocol.CompletionItem{}, nil
	}
	return computeCompletions(doc, params.Position), nil
}

// computeCompletions determines the completion context from the text before the
// cursor and returns appropriate completion items.
func computeCompletions(doc *Document, pos protocol.Position) []protocol.CompletionItem {
	if len(doc.Source) == 0 {
		return nil
	}

	// Guard against out-of-range positions.
	if int(pos.Line) >= len(doc.LineOffsets) {
		return nil
	}

	offset := doc.PositionToOffset(pos)
	lineStart := doc.LineOffsets[int(pos.Line)]

	// Clamp offset to valid range.
	if offset > len(doc.Source) {
		offset = len(doc.Source)
	}
	if lineStart > offset {
		lineStart = offset
	}

	textBeforeCursor := doc.Source[lineStart:offset]

	switch {
	case strings.HasSuffix(textBeforeCursor, "[^"):
		return footnoteCompletions(doc)
	case strings.HasSuffix(textBeforeCursor, "]["):
		return referenceCompletions(doc)
	case strings.Contains(textBeforeCursor, "{") && strings.HasSuffix(textBeforeCursor, "#"):
		return headingIDCompletions(doc)
	}

	return nil
}

func footnoteCompletions(doc *Document) []protocol.CompletionItem {
	refKind := protocol.CompletionItemKindReference
	var items []protocol.CompletionItem
	for label, def := range doc.FootnoteDefs {
		lbl := label
		detail := truncate(def.Text, 60)
		insertText := lbl + "]"
		items = append(items, protocol.CompletionItem{
			Label:      lbl,
			Kind:       &refKind,
			Detail:     &detail,
			InsertText: &insertText,
		})
	}
	return items
}

func referenceCompletions(doc *Document) []protocol.CompletionItem {
	refKind := protocol.CompletionItemKindReference
	var items []protocol.CompletionItem
	for label, def := range doc.ReferenceDefs {
		lbl := label
		detail := def.Destination
		insertText := lbl + "]"
		items = append(items, protocol.CompletionItem{
			Label:      lbl,
			Kind:       &refKind,
			Detail:     &detail,
			InsertText: &insertText,
		})
	}
	return items
}

func headingIDCompletions(doc *Document) []protocol.CompletionItem {
	refKind := protocol.CompletionItemKindReference
	var items []protocol.CompletionItem
	for _, h := range doc.Headings {
		if h.ID == "" {
			continue
		}
		id := h.ID
		items = append(items, protocol.CompletionItem{
			Label: id,
			Kind:  &refKind,
		})
	}
	return items
}

// truncate returns s truncated to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

package server

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func handleHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	doc := getDocument(params.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}
	return computeHover(doc, params.Position), nil
}

// computeHover returns hover content for the given position, or nil if there is
// nothing meaningful to show.
func computeHover(doc *Document, pos protocol.Position) *protocol.Hover {
	if len(doc.Source) == 0 {
		return nil
	}

	offset := doc.PositionToOffset(pos)
	if offset > len(doc.Source) {
		offset = len(doc.Source)
	}

	// Try each hover kind in priority order.
	if label := findFootnoteRefAt(doc, offset); label != "" {
		def, ok := doc.FootnoteDefs[label]
		if !ok {
			return nil
		}
		return markdownHover(fmt.Sprintf("**Footnote [^%s]:** %s", label, def.Text))
	}

	if label := findLinkRefAt(doc, offset); label != "" {
		def, ok := doc.ReferenceDefs[label]
		if !ok {
			return nil
		}
		return markdownHover(fmt.Sprintf("**[%s]:** %s", label, def.Destination))
	}

	if url := findInlineLinkAt(doc, offset); url != "" {
		return markdownHover(url)
	}

	return nil
}

// markdownHover wraps a string in a Markdown MarkupContent Hover.
func markdownHover(value string) *protocol.Hover {
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: value,
		},
	}
}

// findFootnoteRefAt scans the source around offset to detect a footnote
// reference pattern "[^label]". Returns the label if the offset falls within
// the pattern, otherwise returns "".
func findFootnoteRefAt(doc *Document, offset int) string {
	src := doc.Source
	n := len(src)
	if n == 0 || offset > n {
		return ""
	}

	// Clamp offset to a valid index (for reading chars).
	idx := offset
	if idx >= n {
		idx = n - 1
	}

	// Scan backward from offset to find "[^".
	start := -1
	for i := idx; i >= 1; i-- {
		if src[i-1] == '[' && src[i] == '^' {
			start = i - 1 // points to '['
			break
		}
		// Stop if we cross a line boundary or a closing ']'.
		if src[i] == '\n' || src[i] == ']' {
			break
		}
	}
	// Also check if we're already sitting exactly on '[' or '^'
	if start == -1 {
		if idx < n-1 && src[idx] == '[' && src[idx+1] == '^' {
			start = idx
		} else if idx > 0 && src[idx-1] == '[' && src[idx] == '^' {
			start = idx - 1
		}
	}
	if start == -1 {
		return ""
	}

	// Scan forward from start+2 (after "[^") to find the closing "]".
	labelStart := start + 2
	end := strings.IndexByte(src[labelStart:], ']')
	if end < 0 {
		return ""
	}
	closeIdx := labelStart + end // index of ']'

	// The offset must be within [start, closeIdx] (inclusive).
	if offset < start || offset > closeIdx {
		return ""
	}

	// Make sure this is not a definition (followed by ":").
	afterClose := closeIdx + 1
	if afterClose < n && src[afterClose] == ':' {
		return ""
	}

	label := src[labelStart:closeIdx]
	if label == "" {
		return ""
	}
	return label
}

// findLinkRefAt scans the source around offset to detect a reference link
// label "[label]" that is preceded by "]" (i.e. the pattern is "][label]").
// Returns the label, or "".
func findLinkRefAt(doc *Document, offset int) string {
	src := doc.Source
	n := len(src)
	if n == 0 || offset > n {
		return ""
	}

	idx := offset
	if idx >= n {
		idx = n - 1
	}

	// Scan backward to find an opening "[" that is preceded by "]".
	start := -1
	for i := idx; i >= 1; i-- {
		if src[i] == '[' {
			// Check that this '[' is preceded by ']'.
			if src[i-1] == ']' {
				start = i // points to '[' of the label bracket
				break
			}
			// Some other '[' — stop scanning.
			break
		}
		if src[i] == '\n' {
			break
		}
	}
	// Also check if we're sitting right on '[' preceded by ']'.
	if start == -1 && idx > 0 && src[idx] == '[' && src[idx-1] == ']' {
		start = idx
	}
	if start == -1 {
		return ""
	}

	// Scan forward to find closing "]".
	labelStart := start + 1
	end := strings.IndexByte(src[labelStart:], ']')
	if end < 0 {
		return ""
	}
	closeIdx := labelStart + end // index of ']'

	// Cursor must be within [start, closeIdx].
	if offset < start || offset > closeIdx {
		return ""
	}

	label := src[labelStart:closeIdx]
	// Skip footnote refs (label starts with "^").
	if strings.HasPrefix(label, "^") || label == "" {
		return ""
	}
	return label
}

// findInlineLinkAt scans the source around offset to detect an inline link URL
// in the pattern "(url)" preceded by "]". Returns the URL, or "".
func findInlineLinkAt(doc *Document, offset int) string {
	src := doc.Source
	n := len(src)
	if n == 0 || offset > n {
		return ""
	}

	idx := offset
	if idx >= n {
		idx = n - 1
	}

	// Scan backward from idx to find an opening "(" preceded by "]".
	start := -1
	for i := idx; i >= 1; i-- {
		if src[i] == '(' {
			if src[i-1] == ']' {
				start = i // points to '('
				break
			}
			break
		}
		if src[i] == '\n' || src[i] == ')' {
			break
		}
	}
	// Also handle cursor sitting exactly on '('.
	if start == -1 && idx > 0 && src[idx] == '(' && src[idx-1] == ']' {
		start = idx
	}
	if start == -1 {
		return ""
	}

	// Scan forward to find closing ")".
	urlStart := start + 1
	end := strings.IndexByte(src[urlStart:], ')')
	if end < 0 {
		return ""
	}
	closeIdx := urlStart + end // index of ')'

	// Cursor must be within [start, closeIdx].
	if offset < start || offset > closeIdx {
		return ""
	}

	url := src[urlStart:closeIdx]
	if url == "" {
		return ""
	}
	return url
}

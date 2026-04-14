package server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sivukhin/godjot/djot_parser"
	"github.com/sivukhin/godjot/djot_tokenizer"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// HeadingInfo holds extracted information about a heading node.
type HeadingInfo struct {
	Label    string
	Level    int
	ID       string
	Position protocol.Position
	Range    protocol.Range
}

// DefinitionInfo holds extracted information about a link/footnote definition.
type DefinitionInfo struct {
	Label       string
	Destination string
	Position    protocol.Position
	Range       protocol.Range
	Text        string
}

// ReferenceInfo holds a reference to a footnote or link label used in the text.
type ReferenceInfo struct {
	Label    string
	Position protocol.Position
	Range    protocol.Range
}

// Document holds all per-document state for an open djot file.
type Document struct {
	URI                  string
	Source               string
	LineOffsets          []int
	AST                  []djot_parser.TreeNode[djot_parser.DjotNode]
	Headings             []HeadingInfo
	FootnoteDefs         map[string]DefinitionInfo
	FootnoteRefs         []ReferenceInfo
	ReferenceDefs        map[string]DefinitionInfo
	ReferenceRefs        []ReferenceInfo
	DuplicateDiagnostics []protocol.Diagnostic
}

// NewDocument creates a Document by parsing the source and building all indexes.
func NewDocument(uri, source string) *Document {
	doc := &Document{
		URI:           uri,
		Source:        source,
		LineOffsets:   buildLineOffsets(source),
		FootnoteDefs:  make(map[string]DefinitionInfo),
		ReferenceDefs: make(map[string]DefinitionInfo),
	}
	doc.AST = djot_parser.BuildDjotAst([]byte(source))
	doc.buildIndexes()
	return doc
}

// buildLineOffsets returns a slice where offsets[i] is the byte offset of the
// start of line i (0-indexed).
func buildLineOffsets(source string) []int {
	offsets := []int{0}
	for i, b := range []byte(source) {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

// OffsetToPosition converts a byte offset into an LSP Position (line/character).
func (d *Document) OffsetToPosition(offset int) protocol.Position {
	// Binary search for the line containing this offset.
	line := sort.Search(len(d.LineOffsets), func(i int) bool {
		return d.LineOffsets[i] > offset
	}) - 1
	if line < 0 {
		line = 0
	}
	char := offset - d.LineOffsets[line]
	return protocol.Position{
		Line:      uint32(line),
		Character: uint32(char),
	}
}

// PositionToOffset converts an LSP Position into a byte offset.
func (d *Document) PositionToOffset(pos protocol.Position) int {
	line := int(pos.Line)
	if line >= len(d.LineOffsets) {
		return len(d.Source)
	}
	return d.LineOffsets[line] + int(pos.Character)
}

// findInSource returns the byte offset of the first occurrence of pattern in
// source, or -1 if not found.
func (d *Document) findInSource(pattern string) int {
	idx := strings.Index(d.Source, pattern)
	return idx
}

// findInSourceAfter returns the byte offset of the first occurrence of pattern
// in source starting at or after fromOffset, or -1 if not found.
func (d *Document) findInSourceAfter(pattern string, fromOffset int) int {
	if fromOffset > len(d.Source) {
		return -1
	}
	idx := strings.Index(d.Source[fromOffset:], pattern)
	if idx < 0 {
		return -1
	}
	return fromOffset + idx
}

// walkNode recursively processes a node, collecting headings, definitions, and references.
func (d *Document) walkNode(node djot_parser.TreeNode[djot_parser.DjotNode]) {
	switch node.Type {
	case djot_parser.SectionNode:
		// The section ID (auto-generated from the heading text) lives on the
		// SectionNode, not on the child HeadingNode. We handle it here so we can
		// pass the ID down to indexHeading.
		sectionID := node.Attributes.Get(djot_parser.IdKey)
		for _, child := range node.Children {
			if child.Type == djot_parser.HeadingNode {
				d.indexHeadingWithID(child, sectionID)
			} else {
				d.walkNode(child)
			}
		}
		return // children already handled above

	case djot_parser.HeadingNode:
		// HeadingNode encountered outside a SectionNode (rare); index without an ID.
		d.indexHeadingWithID(node, "")

	case djot_parser.FootnoteDefNode:
		d.indexFootnoteDef(node)

	case djot_parser.LinkNode:
		// Footnote refs and reference refs are collected by source scanning passes.
		_ = node
	}

	// Recurse into children
	for _, child := range node.Children {
		d.walkNode(child)
	}
}

func (d *Document) indexHeadingWithID(node djot_parser.TreeNode[djot_parser.DjotNode], id string) {
	levelStr := node.Attributes.Get(djot_parser.HeadingLevelKey)
	level := len(levelStr) // "#" = 1, "##" = 2, etc.
	text := strings.TrimSpace(string(node.FullText()))

	// Find the heading in source by searching for the level markers + text.
	// e.g. "## Heading Two"
	marker := strings.Repeat("#", level) + " " + text
	startOffset := d.findInSource(marker)

	var startPos, endPos protocol.Position
	if startOffset >= 0 {
		startPos = d.OffsetToPosition(startOffset)
		endPos = d.OffsetToPosition(startOffset + len(marker))
	}

	info := HeadingInfo{
		Label:    text,
		Level:    level,
		ID:       id,
		Position: startPos,
		Range: protocol.Range{
			Start: startPos,
			End:   endPos,
		},
	}
	d.Headings = append(d.Headings, info)
}

func (d *Document) indexFootnoteDef(node djot_parser.TreeNode[djot_parser.DjotNode]) {
	label := node.Attributes.Get(djot_tokenizer.ReferenceKey)
	if label == "" {
		return
	}

	// Search for "[^label]:" in source
	pattern := fmt.Sprintf("[^%s]:", label)
	startOffset := d.findInSource(pattern)

	var startPos, endPos protocol.Position
	if startOffset >= 0 {
		startPos = d.OffsetToPosition(startOffset)
		endPos = d.OffsetToPosition(startOffset + len(pattern))
	}

	info := DefinitionInfo{
		Label:    label,
		Position: startPos,
		Range: protocol.Range{
			Start: startPos,
			End:   endPos,
		},
		Text: strings.TrimSpace(string(node.FullText())),
	}

	if _, exists := d.FootnoteDefs[label]; exists {
		// Duplicate — emit a diagnostic
		diag := protocol.Diagnostic{
			Range:    info.Range,
			Severity: severityPtr(protocol.DiagnosticSeverityWarning),
			Message:  fmt.Sprintf("duplicate footnote definition: %q", label),
		}
		d.DuplicateDiagnostics = append(d.DuplicateDiagnostics, diag)
		return
	}
	d.FootnoteDefs[label] = info
}

// scanReferenceDefs scans source for "[label]: url" reference definition patterns.
// godjot resolves references in BuildDjotContext and does NOT emit ReferenceDefNode
// in the AST, so we must extract definitions directly from the source text.
func (d *Document) scanReferenceDefs() {
	src := d.Source
	// A reference definition is: at start of line (or start of file), "[label]: url"
	// We use a line-by-line scan.
	lines := strings.Split(src, "\n")
	offset := 0
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "[^") {
			closeBracket := strings.Index(trimmed, "]:")
			if closeBracket > 0 {
				label := trimmed[1:closeBracket]
				rest := strings.TrimSpace(trimmed[closeBracket+2:])
				// rest is the URL/destination
				destination := rest

				// Find this pattern in source at the correct offset
				patternOffset := offset + (len(line) - len(trimmed))
				startPos := d.OffsetToPosition(patternOffset)
				endOffset := patternOffset + closeBracket + 2 // through "]:"
				endPos := d.OffsetToPosition(endOffset)

				info := DefinitionInfo{
					Label:       label,
					Destination: destination,
					Position:    startPos,
					Range: protocol.Range{
						Start: startPos,
						End:   endPos,
					},
				}

				if _, exists := d.ReferenceDefs[label]; exists {
					diag := protocol.Diagnostic{
						Range:    info.Range,
						Severity: severityPtr(protocol.DiagnosticSeverityWarning),
						Message:  fmt.Sprintf("duplicate reference definition: %q", label),
					}
					d.DuplicateDiagnostics = append(d.DuplicateDiagnostics, diag)
				} else {
					d.ReferenceDefs[label] = info
				}
			}
		}
		offset += len(line) + 1 // +1 for the "\n" we split on
	}
}

// buildIndexes walks the AST and scans source for reference usages.
func (d *Document) buildIndexes() {
	for _, node := range d.AST {
		d.walkNode(node)
	}
	d.scanReferenceDefs()
	d.scanFootnoteRefs()
	d.scanReferenceRefs()
}

// scanFootnoteRefs scans the source for "[^label]" patterns not followed by ":"
// to build the FootnoteRefs list.
func (d *Document) scanFootnoteRefs() {
	src := d.Source
	i := 0
	for i < len(src) {
		// Find "[^"
		idx := strings.Index(src[i:], "[^")
		if idx < 0 {
			break
		}
		start := i + idx
		// Find closing "]"
		end := strings.Index(src[start:], "]")
		if end < 0 {
			break
		}
		end = start + end // index of ']'
		label := src[start+2 : end]
		// Check that the character after "]" is not ":"
		afterEnd := end + 1
		if afterEnd < len(src) && src[afterEnd] == ':' {
			// This is a definition, skip
			i = afterEnd + 1
			continue
		}
		if label == "" {
			i = end + 1
			continue
		}
		startPos := d.OffsetToPosition(start)
		endPos := d.OffsetToPosition(end + 1)
		d.FootnoteRefs = append(d.FootnoteRefs, ReferenceInfo{
			Label:    label,
			Position: startPos,
			Range: protocol.Range{
				Start: startPos,
				End:   endPos,
			},
		})
		i = end + 1
	}
}

// scanReferenceRefs scans the source for "[text][label]" reference usage patterns
// to build the ReferenceRefs list.
func (d *Document) scanReferenceRefs() {
	src := d.Source
	i := 0
	for i < len(src) {
		// Find "][" which indicates a reference link [text][label]
		idx := strings.Index(src[i:], "][")
		if idx < 0 {
			break
		}
		// The second "[" is at i+idx+1
		labelStart := i + idx + 1 // points to '['
		// Find closing "]"
		end := strings.Index(src[labelStart:], "]")
		if end < 0 {
			break
		}
		end = labelStart + end // index of ']'
		label := src[labelStart+1 : end]
		if label == "" {
			i = end + 1
			continue
		}
		// Skip if this looks like a footnote ref "[^..."
		if strings.HasPrefix(label, "^") {
			i = end + 1
			continue
		}
		startPos := d.OffsetToPosition(labelStart)
		endPos := d.OffsetToPosition(end + 1)
		d.ReferenceRefs = append(d.ReferenceRefs, ReferenceInfo{
			Label:    label,
			Position: startPos,
			Range: protocol.Range{
				Start: startPos,
				End:   endPos,
			},
		})
		i = end + 1
	}
}

func severityPtr(s protocol.DiagnosticSeverity) *protocol.DiagnosticSeverity {
	return &s
}

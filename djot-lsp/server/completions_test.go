package server

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// pos is a helper to build a protocol.Position.
func pos(line, char uint32) protocol.Position {
	return protocol.Position{Line: line, Character: char}
}

func TestComputeCompletions_NoCompletionsInPlainText(t *testing.T) {
	source := "This is just plain text with no special syntax.\n"
	doc := NewDocument("file:///test.djot", source)
	items := computeCompletions(doc, pos(0, 10))
	if len(items) != 0 {
		t.Errorf("expected 0 completions in plain text, got %d", len(items))
	}
}

func TestComputeCompletions_FootnoteLabels(t *testing.T) {
	// Single footnote definition — parser reliably indexes this.
	source := "Text [^fn1] here.\n\n[^fn1]: The footnote text.\n"
	doc := NewDocument("file:///test.djot", source)

	// Line 0: "Text [^fn1] here."
	// Cursor at character 7 → textBeforeCursor = "Text [^"
	items := computeCompletions(doc, pos(0, 7))
	if len(items) != 1 {
		t.Fatalf("expected 1 footnote completion, got %d", len(items))
	}
	if items[0].Label != "fn1" {
		t.Errorf("expected label 'fn1', got %q", items[0].Label)
	}
}

func TestComputeCompletions_FootnoteInsertText(t *testing.T) {
	source := "Text [^fn].\n\n[^fn]: The footnote.\n"
	doc := NewDocument("file:///test.djot", source)

	// cursor after "[^" on line 0: "Text [^"
	items := computeCompletions(doc, pos(0, 7))
	if len(items) != 1 {
		t.Fatalf("expected 1 footnote completion, got %d", len(items))
	}
	if items[0].InsertText == nil || *items[0].InsertText != "fn]" {
		got := "<nil>"
		if items[0].InsertText != nil {
			got = *items[0].InsertText
		}
		t.Errorf("expected InsertText 'fn]', got %q", got)
	}
}

func TestComputeCompletions_FootnoteDetail(t *testing.T) {
	source := "Text [^fn].\n\n[^fn]: Short note.\n"
	doc := NewDocument("file:///test.djot", source)

	items := computeCompletions(doc, pos(0, 7))
	if len(items) != 1 {
		t.Fatalf("expected 1 footnote completion, got %d", len(items))
	}
	if items[0].Detail == nil {
		t.Error("expected Detail to be set for footnote completion")
	}
}

func TestComputeCompletions_FootnoteKind(t *testing.T) {
	source := "Text [^fn].\n\n[^fn]: A footnote.\n"
	doc := NewDocument("file:///test.djot", source)

	items := computeCompletions(doc, pos(0, 7))
	if len(items) != 1 {
		t.Fatalf("expected 1 footnote completion, got %d", len(items))
	}
	if items[0].Kind == nil {
		t.Error("expected Kind to be set")
	} else if *items[0].Kind != protocol.CompletionItemKindReference {
		t.Errorf("expected Kind Reference, got %v", *items[0].Kind)
	}
}

func TestComputeCompletions_ReferenceLabels(t *testing.T) {
	source := "See [link text][ref1] and [other][ref2].\n\n[ref1]: https://example.com\n[ref2]: https://other.com\n"
	doc := NewDocument("file:///test.djot", source)

	// Line 0: "See [link text][ref1] and [other][ref2]."
	// "[link text][" ends at index 16 → cursor at character 16
	// textBeforeCursor = "See [link text]["  — ends with "]["
	items := computeCompletions(doc, pos(0, 16))
	if len(items) != 2 {
		t.Fatalf("expected 2 reference completions, got %d", len(items))
	}

	labels := make(map[string]bool)
	for _, item := range items {
		labels[item.Label] = true
	}
	if !labels["ref1"] {
		t.Errorf("expected 'ref1' in completions, got %v", labels)
	}
	if !labels["ref2"] {
		t.Errorf("expected 'ref2' in completions, got %v", labels)
	}
}

func TestComputeCompletions_ReferenceInsertText(t *testing.T) {
	source := "See [link][myref].\n\n[myref]: https://example.com\n"
	doc := NewDocument("file:///test.djot", source)

	// Line 0: "See [link][" — cursor at character 11
	items := computeCompletions(doc, pos(0, 11))
	if len(items) != 1 {
		t.Fatalf("expected 1 reference completion, got %d", len(items))
	}
	if items[0].InsertText == nil || *items[0].InsertText != "myref]" {
		got := "<nil>"
		if items[0].InsertText != nil {
			got = *items[0].InsertText
		}
		t.Errorf("expected InsertText 'myref]', got %q", got)
	}
}

func TestComputeCompletions_ReferenceDetail(t *testing.T) {
	source := "See [link][ref].\n\n[ref]: https://example.com\n"
	doc := NewDocument("file:///test.djot", source)

	items := computeCompletions(doc, pos(0, 11))
	if len(items) != 1 {
		t.Fatalf("expected 1 reference completion, got %d", len(items))
	}
	if items[0].Detail == nil || *items[0].Detail != "https://example.com" {
		got := "<nil>"
		if items[0].Detail != nil {
			got = *items[0].Detail
		}
		t.Errorf("expected Detail 'https://example.com', got %q", got)
	}
}

func TestComputeCompletions_HeadingIDs(t *testing.T) {
	// Headings get auto-generated IDs from their text ("First Heading" → "First-Heading").
	source := "# First Heading\n\n## Second Heading\n\nSee [](#"
	doc := NewDocument("file:///test.djot", source)

	// Line 4: "See [](#" — character 8
	// textBeforeCursor contains "{" (from "[](#") — wait, "[](#" contains "#" but no "{"
	// We need "{" in textBeforeCursor. Use "{#" context directly.
	// Use a line like "span {#" which has "{" and ends with "#"
	source2 := "# First Heading\n\n## Second Heading\n\nspan {#"
	doc2 := NewDocument("file:///test.djot", source2)

	// Line 4: "span {#" — character 7
	items := computeCompletions(doc2, pos(4, 7))
	if len(items) != 2 {
		t.Fatalf("expected 2 heading ID completions, got %d: %v", len(items), items)
	}

	ids := make(map[string]bool)
	for _, item := range items {
		ids[item.Label] = true
	}
	if !ids["First-Heading"] {
		t.Errorf("expected 'First-Heading' in completions, got %v", ids)
	}
	if !ids["Second-Heading"] {
		t.Errorf("expected 'Second-Heading' in completions, got %v", ids)
	}

	// Suppress unused variable warning
	_ = doc
}

func TestComputeCompletions_HeadingIDsOnlyWhenHasID(t *testing.T) {
	// Headings without auto-generated IDs shouldn't appear. But since the parser
	// always generates IDs from heading text, we verify headings with non-empty IDs appear.
	// Use source with headings that have known auto-generated IDs.
	source := "# Alpha\n\n## Beta\n\nsome {#"
	doc := NewDocument("file:///test.djot", source)

	// Line 4: "some {#" — character 7
	items := computeCompletions(doc, pos(4, 7))
	// Both headings have auto-generated IDs "Alpha" and "Beta"
	if len(items) < 1 {
		t.Fatalf("expected at least 1 heading ID completion, got %d", len(items))
	}
	ids := make(map[string]bool)
	for _, item := range items {
		ids[item.Label] = true
	}
	if !ids["Alpha"] {
		t.Errorf("expected 'Alpha' in completions, got %v", ids)
	}
	if !ids["Beta"] {
		t.Errorf("expected 'Beta' in completions, got %v", ids)
	}
}

func TestComputeCompletions_EmptyDocument(t *testing.T) {
	doc := NewDocument("file:///test.djot", "")
	items := computeCompletions(doc, pos(0, 0))
	if len(items) != 0 {
		t.Errorf("expected 0 completions for empty document, got %d", len(items))
	}
}

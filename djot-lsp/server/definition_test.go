package server

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestComputeDefinition_FootnoteRef(t *testing.T) {
	source := "Here is a footnote ref [^note] in text.\n\n[^note]: The footnote body.\n"
	doc := NewDocument("file:///test.djot", source)

	// Position cursor on "note" inside "[^note]" on line 0 (offset ~25)
	pos := protocol.Position{Line: 0, Character: 25}
	loc := computeDefinition(doc, pos)

	if loc == nil {
		t.Fatal("expected a location, got nil")
	}
	if loc.URI != "file:///test.djot" {
		t.Errorf("expected URI %q, got %q", "file:///test.djot", loc.URI)
	}
	// The definition "[^note]:" is on line 2 (0-indexed)
	if loc.Range.Start.Line != 2 {
		t.Errorf("expected definition on line 2, got line %d", loc.Range.Start.Line)
	}
}

func TestComputeDefinition_LinkRef(t *testing.T) {
	source := "See [the site][example] for details.\n\n[example]: https://example.com\n"
	doc := NewDocument("file:///test.djot", source)

	// Position cursor on "example" inside the second "[example]" on line 0
	// "[the site]" ends at offset 10, then "[example]" starts at 10
	// cursor at character 16 (inside "example")
	pos := protocol.Position{Line: 0, Character: 16}
	loc := computeDefinition(doc, pos)

	if loc == nil {
		t.Fatal("expected a location, got nil")
	}
	if loc.URI != "file:///test.djot" {
		t.Errorf("expected URI %q, got %q", "file:///test.djot", loc.URI)
	}
	// The definition "[example]:" is on line 2
	if loc.Range.Start.Line != 2 {
		t.Errorf("expected definition on line 2, got line %d", loc.Range.Start.Line)
	}
}

func TestComputeDefinition_PlainText(t *testing.T) {
	source := "Just some plain text with nothing special here.\n"
	doc := NewDocument("file:///test.djot", source)

	pos := protocol.Position{Line: 0, Character: 5}
	loc := computeDefinition(doc, pos)

	if loc != nil {
		t.Errorf("expected nil for plain text, got %+v", loc)
	}
}

func TestComputeDefinition_UndefinedFootnote(t *testing.T) {
	// Footnote reference present, but no definition
	source := "See [^missing] here.\n"
	doc := NewDocument("file:///test.djot", source)

	pos := protocol.Position{Line: 0, Character: 6}
	loc := computeDefinition(doc, pos)

	if loc != nil {
		t.Errorf("expected nil for undefined footnote, got %+v", loc)
	}
}

func TestComputeDefinition_UndefinedLinkRef(t *testing.T) {
	// Link reference present, but no definition
	source := "See [text][nowhere] here.\n"
	doc := NewDocument("file:///test.djot", source)

	pos := protocol.Position{Line: 0, Character: 12}
	loc := computeDefinition(doc, pos)

	if loc != nil {
		t.Errorf("expected nil for undefined link ref, got %+v", loc)
	}
}

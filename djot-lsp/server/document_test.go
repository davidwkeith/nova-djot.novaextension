package server

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

const testDoc = `# Heading One

## Heading Two

Some paragraph text with a [link][ref1] and a footnote[^fn1].

[ref1]: https://example.com

[^fn1]: This is the footnote content.

### Heading Three

Another paragraph with [^fn1] again.
`

func TestNewDocument_HeadingCount(t *testing.T) {
	doc := NewDocument("file:///test.djot", testDoc)
	if len(doc.Headings) != 3 {
		t.Errorf("expected 3 headings, got %d", len(doc.Headings))
	}
}

func TestNewDocument_HeadingLabels(t *testing.T) {
	doc := NewDocument("file:///test.djot", testDoc)
	if len(doc.Headings) < 1 {
		t.Fatal("no headings found")
	}
	if doc.Headings[0].Label != "Heading One" {
		t.Errorf("expected 'Heading One', got %q", doc.Headings[0].Label)
	}
	if doc.Headings[1].Label != "Heading Two" {
		t.Errorf("expected 'Heading Two', got %q", doc.Headings[1].Label)
	}
	if doc.Headings[2].Label != "Heading Three" {
		t.Errorf("expected 'Heading Three', got %q", doc.Headings[2].Label)
	}
}

func TestNewDocument_HeadingLevels(t *testing.T) {
	doc := NewDocument("file:///test.djot", testDoc)
	if len(doc.Headings) < 3 {
		t.Fatal("not enough headings")
	}
	if doc.Headings[0].Level != 1 {
		t.Errorf("expected level 1, got %d", doc.Headings[0].Level)
	}
	if doc.Headings[1].Level != 2 {
		t.Errorf("expected level 2, got %d", doc.Headings[1].Level)
	}
	if doc.Headings[2].Level != 3 {
		t.Errorf("expected level 3, got %d", doc.Headings[2].Level)
	}
}

func TestNewDocument_FootnoteDefs(t *testing.T) {
	doc := NewDocument("file:///test.djot", testDoc)
	if len(doc.FootnoteDefs) != 1 {
		t.Errorf("expected 1 footnote def, got %d", len(doc.FootnoteDefs))
	}
	def, ok := doc.FootnoteDefs["fn1"]
	if !ok {
		t.Error("expected footnote def 'fn1' not found")
		return
	}
	if def.Label != "fn1" {
		t.Errorf("expected label 'fn1', got %q", def.Label)
	}
}

func TestNewDocument_ReferenceDefs(t *testing.T) {
	doc := NewDocument("file:///test.djot", testDoc)
	if len(doc.ReferenceDefs) != 1 {
		t.Errorf("expected 1 reference def, got %d", len(doc.ReferenceDefs))
	}
	def, ok := doc.ReferenceDefs["ref1"]
	if !ok {
		t.Error("expected reference def 'ref1' not found")
		return
	}
	if def.Label != "ref1" {
		t.Errorf("expected label 'ref1', got %q", def.Label)
	}
	if def.Destination != "https://example.com" {
		t.Errorf("expected destination 'https://example.com', got %q", def.Destination)
	}
}

func TestOffsetToPosition_FirstLine(t *testing.T) {
	doc := NewDocument("file:///test.djot", "hello\nworld\n")
	pos := doc.OffsetToPosition(0)
	if pos.Line != 0 || pos.Character != 0 {
		t.Errorf("expected 0:0, got %d:%d", pos.Line, pos.Character)
	}
}

func TestOffsetToPosition_SecondLine(t *testing.T) {
	doc := NewDocument("file:///test.djot", "hello\nworld\n")
	pos := doc.OffsetToPosition(6)
	if pos.Line != 1 || pos.Character != 0 {
		t.Errorf("expected 1:0, got %d:%d", pos.Line, pos.Character)
	}
}

func TestOffsetToPosition_MidLine(t *testing.T) {
	doc := NewDocument("file:///test.djot", "hello\nworld\n")
	pos := doc.OffsetToPosition(8) // 'r' in 'world'
	if pos.Line != 1 || pos.Character != 2 {
		t.Errorf("expected 1:2, got %d:%d", pos.Line, pos.Character)
	}
}

func TestPositionToOffset_RoundTrip(t *testing.T) {
	source := "hello\nworld\nfoo\n"
	doc := NewDocument("file:///test.djot", source)
	offsets := []int{0, 3, 6, 9, 12, 14}
	for _, offset := range offsets {
		pos := doc.OffsetToPosition(offset)
		back := doc.PositionToOffset(pos)
		if back != offset {
			t.Errorf("round-trip failed for offset %d: got %d (via %d:%d)", offset, back, pos.Line, pos.Character)
		}
	}
}

func TestPositionToOffset_KnownPosition(t *testing.T) {
	doc := NewDocument("file:///test.djot", "hello\nworld\n")
	pos := protocol.Position{Line: 1, Character: 3}
	offset := doc.PositionToOffset(pos)
	if offset != 9 { // 6 + 3
		t.Errorf("expected offset 9, got %d", offset)
	}
}

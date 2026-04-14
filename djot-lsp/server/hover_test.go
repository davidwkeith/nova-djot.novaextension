package server

import (
	"strings"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// hoverPos is a helper to build a protocol.Position for hover tests.
// (pos() is already defined in completions_test.go so we reuse it.)

func TestComputeHover_FootnoteRef(t *testing.T) {
	source := "Text [^fn1] here.\n\n[^fn1]: The footnote text.\n"
	doc := NewDocument("file:///test.djot", source)

	// Line 0: "Text [^fn1] here."
	// "[^fn1]" spans characters 5..10. Place cursor inside at character 7 (on "f").
	result := computeHover(doc, pos(0, 7))
	if result == nil {
		t.Fatal("expected non-nil hover for footnote reference")
	}

	mc, ok := result.Contents.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("expected MarkupContent, got %T", result.Contents)
	}
	if mc.Kind != protocol.MarkupKindMarkdown {
		t.Errorf("expected Markdown kind, got %q", mc.Kind)
	}
	if !strings.Contains(mc.Value, "fn1") {
		t.Errorf("expected hover to mention label 'fn1', got %q", mc.Value)
	}
	if !strings.Contains(mc.Value, "The footnote text") {
		t.Errorf("expected hover to contain footnote text, got %q", mc.Value)
	}
}

func TestComputeHover_PlainText_Nil(t *testing.T) {
	source := "This is plain text with no links or footnotes.\n"
	doc := NewDocument("file:///test.djot", source)

	result := computeHover(doc, pos(0, 5))
	if result != nil {
		t.Errorf("expected nil hover for plain text, got %+v", result)
	}
}

func TestComputeHover_LinkReference(t *testing.T) {
	source := "See [link text][myref] for details.\n\n[myref]: https://example.com\n"
	doc := NewDocument("file:///test.djot", source)

	// Line 0: "See [link text][myref] for details."
	// "[myref]" spans characters 15..21. Cursor at character 17 (on "r").
	result := computeHover(doc, pos(0, 17))
	if result == nil {
		t.Fatal("expected non-nil hover for link reference")
	}

	mc, ok := result.Contents.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("expected MarkupContent, got %T", result.Contents)
	}
	if mc.Kind != protocol.MarkupKindMarkdown {
		t.Errorf("expected Markdown kind, got %q", mc.Kind)
	}
	if !strings.Contains(mc.Value, "myref") {
		t.Errorf("expected hover to mention label 'myref', got %q", mc.Value)
	}
	if !strings.Contains(mc.Value, "https://example.com") {
		t.Errorf("expected hover to contain URL, got %q", mc.Value)
	}
}

func TestComputeHover_InlineLink(t *testing.T) {
	source := "Visit [my site](https://example.org) today.\n"
	doc := NewDocument("file:///test.djot", source)

	// Line 0: "Visit [my site](https://example.org) today."
	// "(https://example.org)" spans characters 15..35. Cursor at character 20 (on "e" in example).
	result := computeHover(doc, pos(0, 20))
	if result == nil {
		t.Fatal("expected non-nil hover for inline link")
	}

	mc, ok := result.Contents.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("expected MarkupContent, got %T", result.Contents)
	}
	if !strings.Contains(mc.Value, "https://example.org") {
		t.Errorf("expected hover to contain URL, got %q", mc.Value)
	}
}

func TestComputeHover_FootnoteRef_UnknownLabel(t *testing.T) {
	// Footnote ref present but no definition — should return nil (nothing useful to show)
	source := "Text [^missing] here.\n"
	doc := NewDocument("file:///test.djot", source)

	// Cursor inside "[^missing]"
	result := computeHover(doc, pos(0, 8))
	if result != nil {
		t.Errorf("expected nil hover for unknown footnote label, got %+v", result)
	}
}

func TestComputeHover_LinkRef_UnknownLabel(t *testing.T) {
	// Reference link present but no definition — should return nil
	source := "See [link text][ghost] here.\n"
	doc := NewDocument("file:///test.djot", source)

	// Cursor inside "[ghost]" — character 17
	result := computeHover(doc, pos(0, 17))
	if result != nil {
		t.Errorf("expected nil hover for unknown reference label, got %+v", result)
	}
}

func TestComputeHover_EmptyDocument(t *testing.T) {
	doc := NewDocument("file:///test.djot", "")
	result := computeHover(doc, pos(0, 0))
	if result != nil {
		t.Errorf("expected nil hover for empty document, got %+v", result)
	}
}

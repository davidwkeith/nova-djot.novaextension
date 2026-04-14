package server

import (
	"testing"
)

func TestComputeDiagnostics_UndefinedFootnoteRef(t *testing.T) {
	src := `Some text with [^missing] footnote reference.
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Message != `Undefined footnote reference: "missing"` {
		t.Errorf("unexpected message: %q", diags[0].Message)
	}
}

func TestComputeDiagnostics_DefinedFootnoteRef(t *testing.T) {
	src := `Some text with [^fn1] footnote reference.

[^fn1]: This is the footnote.
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestComputeDiagnostics_UndefinedLinkRef(t *testing.T) {
	src := `A [link][missing] in text.
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Message != `Undefined link reference: "missing"` {
		t.Errorf("unexpected message: %q", diags[0].Message)
	}
}

func TestComputeDiagnostics_DefinedLinkRef(t *testing.T) {
	src := `A [link][ref1] in text.

[ref1]: https://example.com
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestComputeDiagnostics_LinkRefToHeadingID(t *testing.T) {
	// godjot generates the heading ID as "My-Heading" (preserving case, replacing
	// spaces with hyphens). A link reference using that exact ID should resolve.
	src := `# My Heading

A [link][My-Heading] resolving to a heading ID.
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics (heading ID resolves ref), got %d: %v", len(diags), diags)
	}
}

func TestComputeDiagnostics_CleanDocument(t *testing.T) {
	src := `# Heading

A paragraph with no references.
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestComputeDiagnostics_DuplicateRefDef(t *testing.T) {
	// Duplicate reference definitions are detected by source scanning and stored
	// in DuplicateDiagnostics during index building.
	src := `A [link][ref1] in text.

[ref1]: https://example.com

[ref1]: https://other.com
`
	doc := NewDocument("file:///test.djot", src)
	diags := computeDiagnostics(doc)
	// Should have at least 1 diagnostic for the duplicate reference definition
	found := false
	for _, d := range diags {
		if d.Message == `duplicate reference definition: "ref1"` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate reference definition diagnostic, got: %v", diags)
	}
}

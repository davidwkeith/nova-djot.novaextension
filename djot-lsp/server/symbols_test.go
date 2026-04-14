package server

import (
	"testing"
)

func TestComputeDocumentSymbols_Empty(t *testing.T) {
	doc := NewDocument("file:///test.djot", "")
	symbols := computeDocumentSymbols(doc)
	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols for empty document, got %d", len(symbols))
	}
}

func TestComputeDocumentSymbols_FlatHeadings(t *testing.T) {
	source := "# First\n\n# Second\n"
	doc := NewDocument("file:///test.djot", source)
	symbols := computeDocumentSymbols(doc)
	if len(symbols) != 2 {
		t.Fatalf("expected 2 top-level symbols, got %d", len(symbols))
	}
	if symbols[0].Name != "First" {
		t.Errorf("expected first symbol name 'First', got %q", symbols[0].Name)
	}
	if symbols[1].Name != "Second" {
		t.Errorf("expected second symbol name 'Second', got %q", symbols[1].Name)
	}
	if symbols[0].Children != nil {
		t.Errorf("expected no children on first H1, got %d", len(symbols[0].Children))
	}
	if symbols[1].Children != nil {
		t.Errorf("expected no children on second H1, got %d", len(symbols[1].Children))
	}
}

func TestComputeDocumentSymbols_NestedHeadings(t *testing.T) {
	source := "# Top\n\n## Middle\n\n### Bottom\n"
	doc := NewDocument("file:///test.djot", source)
	symbols := computeDocumentSymbols(doc)
	if len(symbols) != 1 {
		t.Fatalf("expected 1 top-level symbol, got %d", len(symbols))
	}
	top := symbols[0]
	if top.Name != "Top" {
		t.Errorf("expected top symbol 'Top', got %q", top.Name)
	}
	if len(top.Children) != 1 {
		t.Fatalf("expected 1 child of Top, got %d", len(top.Children))
	}
	mid := top.Children[0]
	if mid.Name != "Middle" {
		t.Errorf("expected child symbol 'Middle', got %q", mid.Name)
	}
	if len(mid.Children) != 1 {
		t.Fatalf("expected 1 grandchild of Middle, got %d", len(mid.Children))
	}
	bottom := mid.Children[0]
	if bottom.Name != "Bottom" {
		t.Errorf("expected grandchild symbol 'Bottom', got %q", bottom.Name)
	}
}

func TestComputeDocumentSymbols_SiblingH2s(t *testing.T) {
	source := "# Top\n\n## First Section\n\n## Second Section\n"
	doc := NewDocument("file:///test.djot", source)
	symbols := computeDocumentSymbols(doc)
	if len(symbols) != 1 {
		t.Fatalf("expected 1 top-level symbol, got %d", len(symbols))
	}
	top := symbols[0]
	if len(top.Children) != 2 {
		t.Fatalf("expected 2 children of Top, got %d", len(top.Children))
	}
	if top.Children[0].Name != "First Section" {
		t.Errorf("expected 'First Section', got %q", top.Children[0].Name)
	}
	if top.Children[1].Name != "Second Section" {
		t.Errorf("expected 'Second Section', got %q", top.Children[1].Name)
	}
}

func TestComputeDocumentSymbols_Detail(t *testing.T) {
	source := "## A Section\n"
	doc := NewDocument("file:///test.djot", source)
	symbols := computeDocumentSymbols(doc)
	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}
	if symbols[0].Detail == nil || *symbols[0].Detail != "H2" {
		detail := "<nil>"
		if symbols[0].Detail != nil {
			detail = *symbols[0].Detail
		}
		t.Errorf("expected detail 'H2', got %q", detail)
	}
}

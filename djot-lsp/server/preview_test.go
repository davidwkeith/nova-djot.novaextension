package server

import (
	"strings"
	"testing"
)

func TestRenderHTML_BasicParagraph(t *testing.T) {
	doc := NewDocument("file:///test.djot", "Hello world.\n")
	html := renderHTML(doc)
	if !strings.Contains(html, "<p") {
		t.Errorf("expected <p tag, got: %s", html)
	}
	if !strings.Contains(html, "Hello world.") {
		t.Errorf("expected 'Hello world.' in output, got: %s", html)
	}
}

func TestRenderHTML_Heading(t *testing.T) {
	doc := NewDocument("file:///test.djot", "# Hello\n")
	html := renderHTML(doc)
	if !strings.Contains(html, "<h1") {
		t.Errorf("expected <h1 tag, got: %s", html)
	}
	if !strings.Contains(html, "Hello") {
		t.Errorf("expected 'Hello' in output, got: %s", html)
	}
}

func TestRenderHTML_CodeBlock(t *testing.T) {
	doc := NewDocument("file:///test.djot", "```\nfoo bar\n```\n")
	html := renderHTML(doc)
	if !strings.Contains(html, "<pre") {
		t.Errorf("expected <pre tag, got: %s", html)
	}
	if !strings.Contains(html, "<code") {
		t.Errorf("expected <code tag, got: %s", html)
	}
}

func TestRenderHTML_EmptyDocument(t *testing.T) {
	doc := NewDocument("file:///test.djot", "")
	html := renderHTML(doc)
	if html != "" {
		t.Errorf("expected empty string for empty document, got: %q", html)
	}
}

func TestInjectDataLines_AddsAttributes(t *testing.T) {
	doc := NewDocument("file:///test.djot", "Hello world.\n")
	html := renderHTMLWithLineNumbers(doc)
	if !strings.Contains(html, `data-line="`) {
		t.Errorf("expected data-line attribute, got: %s", html)
	}
}

func TestInjectDataLines_MultipleElements(t *testing.T) {
	source := "# Heading\n\nFirst paragraph.\n\nSecond paragraph.\n"
	doc := NewDocument("file:///test.djot", source)
	html := renderHTMLWithLineNumbers(doc)
	count := strings.Count(html, `data-line="`)
	if count < 3 {
		t.Errorf("expected at least 3 data-line attributes, got %d in: %s", count, html)
	}
}

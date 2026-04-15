package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sivukhin/godjot/djot_parser"
	"github.com/sivukhin/godjot/html_writer"
)

// blockTagRe matches opening block-level HTML tags (capturing the tag name and
// position so we can inject data-line attributes).
var blockTagRe = regexp.MustCompile(`(?i)<(p|h[1-6]|pre|blockquote|ul|ol|li|table|div|hr|dl|dt|dd)(\s[^>]*|)>`)

// renderHTML renders doc.AST to an HTML string using godjot's default converter.
// Returns an empty string when the AST is empty (e.g. the document is empty).
func renderHTML(doc *Document) string {
	if len(doc.AST) == 0 {
		return ""
	}
	context := djot_parser.NewConversionContext("html", djot_parser.DefaultConversionRegistry)
	return context.ConvertDjotToHtml(&html_writer.HtmlWriter{}, doc.AST...)
}

// renderHTMLWithLineNumbers renders the document to HTML and then injects
// data-line attributes on every block-level opening tag.
func renderHTMLWithLineNumbers(doc *Document) string {
	html := renderHTML(doc)
	if html == "" {
		return ""
	}
	return injectDataLines(html, doc)
}

// injectDataLines walks through html, finds block-level opening tags, extracts
// a short plain-text sample from the content that follows each tag, locates
// that sample in doc.Source to determine the source line number, and rewrites
// the tag to include a data-line="N" attribute (1-indexed).
func injectDataLines(html string, doc *Document) string {
	// Find all match locations up front so we know each tag's position in html.
	locs := blockTagRe.FindAllStringIndex(html, -1)
	if len(locs) == 0 {
		return html
	}

	// sourceSearchStart tracks where in doc.Source we look next, so that
	// repeated content (e.g. two identical paragraphs) maps to successive
	// occurrences rather than always the first.
	sourceSearchStart := 0

	var sb strings.Builder
	prev := 0
	for _, loc := range locs {
		start, end := loc[0], loc[1]
		// Copy everything between the previous tag end and this tag start.
		sb.WriteString(html[prev:start])

		match := html[start:end]
		contentStart := end

		sample := extractTextSample(html, contentStart, 30)

		replacement := match // default: leave tag unchanged
		if sample != "" {
			// Search for the sample in doc.Source starting after our previous hit.
			idx := strings.Index(doc.Source[sourceSearchStart:], sample)
			if idx < 0 {
				// Fall back to a full search when the window-based search misses.
				idx = strings.Index(doc.Source, sample)
			} else {
				idx += sourceSearchStart
			}

			if idx >= 0 {
				sourceSearchStart = idx

				pos := doc.OffsetToPosition(idx)
				lineNum := int(pos.Line) + 1 // convert to 1-indexed

				// Parse the original tag to inject data-line without duplication.
				submatches := blockTagRe.FindStringSubmatch(match)
				if submatches != nil {
					tagName := submatches[1]
					existingAttrs := submatches[2]
					replacement = fmt.Sprintf("<%s%s data-line=\"%d\">", tagName, existingAttrs, lineNum)
				}
			}
		}

		sb.WriteString(replacement)
		prev = end
	}
	// Append any remaining HTML after the last tag.
	sb.WriteString(html[prev:])
	return sb.String()
}

// extractTextSample strips HTML tags from the content of html starting at pos
// and returns up to maxLen characters of plain text. It stops at newlines so
// that samples stay on a single source line, making them easier to locate in
// the original source. Leading/trailing whitespace is trimmed.
func extractTextSample(html string, pos int, maxLen int) string {
	if pos >= len(html) {
		return ""
	}

	var buf strings.Builder
	inTag := false
	for i := pos; i < len(html) && buf.Len() < maxLen; i++ {
		ch := html[i]
		switch {
		case ch == '<':
			inTag = true
		case ch == '>':
			inTag = false
		case ch == '\n':
			// Stop at newlines so the sample stays within a single source line.
			if buf.Len() > 0 {
				break
			}
		case !inTag:
			buf.WriteByte(ch)
		}
		if ch == '\n' && buf.Len() > 0 {
			break
		}
	}

	return strings.TrimSpace(buf.String())
}

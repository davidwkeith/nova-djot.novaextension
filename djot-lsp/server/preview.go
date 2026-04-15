package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sivukhin/godjot/djot_parser"
	"github.com/sivukhin/godjot/html_writer"
)

const previewDir = "/tmp/djot-preview"

// ---------------------------------------------------------------------------
// HTML rendering
// ---------------------------------------------------------------------------

func renderHTML(doc *Document) string {
	if len(doc.AST) == 0 {
		return ""
	}
	context := djot_parser.NewConversionContext("html", djot_parser.DefaultConversionRegistry)
	return context.ConvertDjotToHtml(&html_writer.HtmlWriter{}, doc.AST...)
}

func renderHTMLWithLineNumbers(doc *Document) string {
	html := renderHTML(doc)
	if html == "" {
		return ""
	}
	return injectDataLines(html, doc)
}

// WritePreviewFile renders the document to HTML and writes it to
// /tmp/djot-preview/{relative-path}.html, mirroring the project structure.
// Returns the output file path.
func WritePreviewFile(doc *Document, workspaceRoot string) string {
	body := renderHTMLWithLineNumbers(doc)
	page := previewDocument(body)

	// Convert URI to file path
	docPath := strings.TrimPrefix(doc.URI, "file://")

	// Compute relative path from workspace root
	var relPath string
	if workspaceRoot != "" && strings.HasPrefix(docPath, workspaceRoot) {
		relPath = strings.TrimPrefix(docPath, workspaceRoot)
	} else {
		relPath = "/" + filepath.Base(docPath)
	}

	// Change .dj extension to .html
	if strings.HasSuffix(relPath, ".dj") {
		relPath = relPath[:len(relPath)-3] + ".html"
	}

	outPath := filepath.Join(previewDir, relPath)

	// Ensure parent directory exists
	os.MkdirAll(filepath.Dir(outPath), 0755)

	os.WriteFile(outPath, []byte(page), 0644)
	return outPath
}

// previewDocument wraps an HTML body fragment in a full HTML page with styles.
func previewDocument(body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Djot Preview</title>
<style>
*, *::before, *::after { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  font-size: 1rem;
  line-height: 1.6;
  color: #24292f;
  background: #ffffff;
  max-width: 48em;
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem;
}
h1, h2, h3, h4, h5, h6 { margin-top: 1.5em; margin-bottom: 0.5em; font-weight: 600; }
h1 { font-size: 2em;   border-bottom: 1px solid #d0d7de; padding-bottom: 0.3em; }
h2 { font-size: 1.5em; border-bottom: 1px solid #d0d7de; padding-bottom: 0.3em; }
p  { margin: 0.8em 0; }
a  { color: #0969da; }
code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.875em;
  background: #f6f8fa;
  padding: 0.2em 0.4em;
  border-radius: 4px;
}
pre {
  background: #f6f8fa;
  border-radius: 6px;
  padding: 1em;
  overflow-x: auto;
  font-size: 0.875em;
}
pre code { background: transparent; padding: 0; border-radius: 0; }
blockquote {
  margin: 1em 0;
  padding: 0.5em 1em;
  border-left: 4px solid #d0d7de;
  color: #57606a;
}
table { border-collapse: collapse; width: 100%%; margin: 1em 0; }
th, td { border: 1px solid #d0d7de; padding: 0.5em 0.75em; text-align: left; }
th { background: #f6f8fa; font-weight: 600; }
ins  { text-decoration: underline; }
del  { text-decoration: line-through; }
sup  { vertical-align: super;  font-size: 0.75em; }
sub  { vertical-align: sub;    font-size: 0.75em; }
@media (prefers-color-scheme: dark) {
  body       { background: #0d1117; color: #e6edf3; }
  a          { color: #58a6ff; }
  h1, h2     { border-bottom-color: #30363d; }
  code, pre  { background: #161b22; }
  blockquote { border-left-color: #30363d; color: #8b949e; }
  th         { background: #161b22; }
  th, td     { border-color: #30363d; }
}
</style>
</head>
<body>
%s
</body>
</html>`, body)
}

// ---------------------------------------------------------------------------
// data-line injection
// ---------------------------------------------------------------------------

var blockTagRe = regexp.MustCompile(`(?i)<(p|h[1-6]|pre|blockquote|ul|ol|li|table|div|hr|dl|dt|dd)(\s[^>]*|)>`)

func injectDataLines(html string, doc *Document) string {
	locs := blockTagRe.FindAllStringIndex(html, -1)
	if len(locs) == 0 {
		return html
	}

	sourceSearchStart := 0
	var sb strings.Builder
	prev := 0
	for _, loc := range locs {
		start, end := loc[0], loc[1]
		sb.WriteString(html[prev:start])

		match := html[start:end]
		sample := extractTextSample(html, end, 30)

		replacement := match
		if sample != "" {
			idx := strings.Index(doc.Source[sourceSearchStart:], sample)
			if idx < 0 {
				idx = strings.Index(doc.Source, sample)
			} else {
				idx += sourceSearchStart
			}

			if idx >= 0 {
				sourceSearchStart = idx
				pos := doc.OffsetToPosition(idx)
				lineNum := int(pos.Line) + 1

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
	sb.WriteString(html[prev:])
	return sb.String()
}

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
			if buf.Len() > 0 {
				goto done
			}
		case !inTag:
			buf.WriteByte(ch)
		}
	}
done:
	return strings.TrimSpace(buf.String())
}

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
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

// ---------------------------------------------------------------------------
// HTTP server + WebSocket hub
// ---------------------------------------------------------------------------

// PreviewServer holds the HTTP server, WebSocket upgrader, and the set of
// currently connected WebSocket clients.
type PreviewServer struct {
	mu       sync.RWMutex
	clients  map[*websocket.Conn]bool
	port     int
	server   *http.Server
	upgrader websocket.Upgrader
}

// preview is the package-level singleton PreviewServer instance.
var preview *PreviewServer

// StartPreviewServer starts an HTTP server on a random localhost port and
// returns the port number. It is safe to call more than once; subsequent calls
// are no-ops when a server is already running.
func StartPreviewServer() (int, error) {
	if preview != nil {
		return preview.port, nil
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("preview server listen: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	ps := &PreviewServer{
		clients: make(map[*websocket.Conn]bool),
		port:    port,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/preview", ps.handlePreviewPage)
	mux.HandleFunc("/ws", ps.handleWebSocket)

	ps.server = &http.Server{Handler: mux}
	preview = ps

	go ps.server.Serve(listener) //nolint:errcheck

	return port, nil
}

// StopPreviewServer gracefully shuts down the preview HTTP server if it is
// running.
func StopPreviewServer() {
	if preview == nil {
		return
	}
	preview.server.Shutdown(context.Background()) //nolint:errcheck
	preview = nil
}

// BroadcastContent renders the document and sends the resulting HTML to every
// connected WebSocket client as a JSON content message.
func BroadcastContent(doc *Document) {
	if preview == nil {
		return
	}
	html := renderHTMLWithLineNumbers(doc)
	msg := fmt.Sprintf(`{"type":"content","html":"%s"}`, jsonEscape(html))
	preview.broadcast([]byte(msg))
}

// BroadcastScroll sends a scroll-sync message to every connected WebSocket
// client telling them to scroll to the given 1-indexed source line.
func BroadcastScroll(line int) {
	if preview == nil {
		return
	}
	msg := fmt.Sprintf(`{"type":"scroll","line":%d}`, line)
	preview.broadcast([]byte(msg))
}

// broadcast sends msg to every connected client. Clients that fail to receive
// the message are silently removed from the set.
func (ps *PreviewServer) broadcast(msg []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for conn := range ps.clients {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			conn.Close()
			delete(ps.clients, conn)
		}
	}
}

// handleWebSocket upgrades the connection, registers the client, discards
// incoming frames, and removes the client on disconnect.
func (ps *PreviewServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ps.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ps.mu.Lock()
	ps.clients[conn] = true
	ps.mu.Unlock()

	defer func() {
		ps.mu.Lock()
		delete(ps.clients, conn)
		ps.mu.Unlock()
		conn.Close()
	}()

	// Read loop: discard all incoming messages; exit on any error (disconnect).
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// handlePreviewPage serves the self-contained preview HTML page.
func (ps *PreviewServer) handlePreviewPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, previewPageHTML(ps.port))
}

// jsonEscape returns s with characters that would break a JSON string literal
// replaced by their escape sequences.
func jsonEscape(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// previewPageHTML returns a complete, self-contained HTML page that connects to
// the WebSocket at ws://localhost:{port}/ws and renders incoming content.
//
// Security note: the HTML content inserted via innerHTML is produced by the
// godjot renderer from the user's own local Djot source and delivered
// exclusively over a localhost-only WebSocket. No external input pathway exists.
func previewPageHTML(port int) string {
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
pre code {
  background: transparent;
  padding: 0;
  border-radius: 0;
}
blockquote {
  margin: 1em 0;
  padding: 0.5em 1em;
  border-left: 4px solid #d0d7de;
  color: #57606a;
}
table {
  border-collapse: collapse;
  width: 100%%;
  margin: 1em 0;
}
th, td {
  border: 1px solid #d0d7de;
  padding: 0.5em 0.75em;
  text-align: left;
}
th { background: #f6f8fa; font-weight: 600; }
ins  { text-decoration: underline; }
del  { text-decoration: line-through; }
sup  { vertical-align: super;  font-size: 0.75em; }
sub  { vertical-align: sub;    font-size: 0.75em; }
#status {
  position: fixed;
  top: 0; left: 0; right: 0;
  padding: 0.4em 1em;
  background: #fff3cd;
  color: #664d03;
  font-size: 0.85em;
  text-align: center;
  z-index: 1000;
  display: none;
}
#status.disconnected { display: block; }
@media (prefers-color-scheme: dark) {
  body       { background: #0d1117; color: #e6edf3; }
  a          { color: #58a6ff; }
  h1, h2     { border-bottom-color: #30363d; }
  code, pre  { background: #161b22; }
  blockquote { border-left-color: #30363d; color: #8b949e; }
  th         { background: #161b22; }
  th, td     { border-color: #30363d; }
  #status    { background: #3d2e00; color: #e3b341; }
}
</style>
</head>
<body>
<div id="status" class="disconnected">Reconnecting...</div>
<div id="content"><p><em>Waiting for content...</em></p></div>
<script>
(function () {
  var port = %d;
  var statusEl  = document.getElementById('status');
  var contentEl = document.getElementById('content');
  var ws;

  function connect() {
    ws = new WebSocket('ws://localhost:' + port + '/ws');

    ws.onopen = function () {
      statusEl.classList.remove('disconnected');
    };

    ws.onclose = function () {
      statusEl.classList.add('disconnected');
      setTimeout(connect, 1000);
    };

    ws.onerror = function () {
      ws.close();
    };

    ws.onmessage = function (event) {
      var msg;
      try { msg = JSON.parse(event.data); } catch (e) { return; }

      if (msg.type === 'content') {
        var savedTop = document.documentElement.scrollTop || document.body.scrollTop;
        // trusted local content from godjot renderer via localhost WebSocket
        contentEl.innerHTML = msg.html;
        document.documentElement.scrollTop = savedTop;
        document.body.scrollTop = savedTop;
      } else if (msg.type === 'scroll') {
        var line = msg.line;
        var els = contentEl.querySelectorAll('[data-line]');
        var target = null;
        for (var i = 0; i < els.length; i++) {
          if (parseInt(els[i].getAttribute('data-line'), 10) <= line) {
            target = els[i];
          } else {
            break;
          }
        }
        if (target) {
          target.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }
    };
  }

  connect();
}());
</script>
</body>
</html>`, port)
}

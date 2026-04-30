var langClient = null;
var previewPort = null;

// ---------------------------------------------------------------------------
// Editor command helpers
// ---------------------------------------------------------------------------

// Toggle delimiters around each selected range. Three behaviors:
//   - Empty range: insert open+close, place cursor between them.
//   - Selection already wrapped (inner): strip delimiters from selected text.
//   - Selection's neighbors form delimiters (outer): expand range, strip them.
//   - Otherwise: wrap the selection.
// All edits batched into one undo step. Multi-cursor safe via reverse-order
// processing. Cursor repositioning only applied for the single-empty-range
// case (other multi-cursor combinations rely on Nova's default behavior).
function wrapSelection(editor, open, close) {
    var ranges = editor.selectedRanges.slice().sort(function(a, b) {
        return b.start - a.start;
    });
    var openLen = open.length;
    var closeLen = close.length;
    var docLen = editor.document.length;

    var soleEmpty = (ranges.length === 1 && ranges[0].start === ranges[0].end);
    var soleEmptyStart = soleEmpty ? ranges[0].start : -1;

    editor.edit(function(e) {
        for (var i = 0; i < ranges.length; i++) {
            var r = ranges[i];

            if (r.start === r.end) {
                e.insert(r.start, open + close);
                continue;
            }

            var selectedText = editor.document.getTextInRange(r);

            if (selectedText.length >= openLen + closeLen &&
                selectedText.substring(0, openLen) === open &&
                selectedText.substring(selectedText.length - closeLen) === close) {
                var inner = selectedText.substring(openLen, selectedText.length - closeLen);
                e.replace(r, inner);
                continue;
            }

            if (r.start >= openLen && r.end + closeLen <= docLen) {
                var outerRange = new Range(r.start - openLen, r.end + closeLen);
                var outerText = editor.document.getTextInRange(outerRange);
                if (outerText.substring(0, openLen) === open &&
                    outerText.substring(outerText.length - closeLen) === close) {
                    e.replace(outerRange, selectedText);
                    continue;
                }
            }

            e.replace(r, open + selectedText + close);
        }
    });

    if (soleEmpty) {
        var pos = soleEmptyStart + openLen;
        editor.selectedRanges = [new Range(pos, pos)];
    }
}

// Toggle a line prefix on every line touched by any selected range.
// Detection rule: if every NON-EMPTY affected line already starts with prefix,
// strip the prefix from those lines (leave blank lines alone). Otherwise,
// prepend the prefix to every affected line that does not already have it
// (blank lines included — `> ` on a blank line is valid blockquote
// continuation).
function prependLines(editor, prefix) {
    var doc = editor.document;
    var docText = doc.getTextInRange(new Range(0, doc.length));
    var ranges = editor.selectedRanges;

    var lineStartSet = {};
    for (var i = 0; i < ranges.length; i++) {
        var r = ranges[i];
        var firstLineStart = lineStartFor(docText, r.start);
        var pos = firstLineStart;
        while (true) {
            lineStartSet[pos] = true;
            var nl = docText.indexOf("\n", pos);
            if (nl === -1 || nl >= r.end) break;
            pos = nl + 1;
            // If the selection ends exactly at a line start, do not include
            // that next line — match the convention of "lines the user is
            // visibly working on".
            if (pos > r.end) break;
        }
    }

    var lineStarts = Object.keys(lineStartSet)
        .map(Number)
        .sort(function(a, b) { return b - a; });

    if (lineStarts.length === 0) return;

    var allHavePrefix = true;
    var anyNonEmpty = false;
    for (var j = 0; j < lineStarts.length; j++) {
        var ls = lineStarts[j];
        var lineEnd = lineEndFor(docText, ls);
        var lineText = docText.substring(ls, lineEnd);
        if (lineText.length === 0) continue;
        anyNonEmpty = true;
        if (lineText.substring(0, prefix.length) !== prefix) {
            allHavePrefix = false;
        }
    }
    var stripping = anyNonEmpty && allHavePrefix;

    editor.edit(function(e) {
        for (var k = 0; k < lineStarts.length; k++) {
            var ls2 = lineStarts[k];
            var lineEnd2 = lineEndFor(docText, ls2);
            var lineText2 = docText.substring(ls2, lineEnd2);

            if (stripping) {
                if (lineText2.substring(0, prefix.length) === prefix) {
                    e.replace(new Range(ls2, ls2 + prefix.length), "");
                }
            } else {
                if (lineText2.substring(0, prefix.length) !== prefix) {
                    e.insert(ls2, prefix);
                }
            }
        }
    });
}

function lineStartFor(docText, pos) {
    var prevNl = docText.lastIndexOf("\n", pos - 1);
    return prevNl === -1 ? 0 : prevNl + 1;
}

function lineEndFor(docText, lineStart) {
    var nextNl = docText.indexOf("\n", lineStart);
    return nextNl === -1 ? docText.length : nextNl;
}

exports.activate = function() {
    var serverPath = nova.path.join(nova.extension.path, "bin", "djot-lsp");

    var serverOptions = {
        path: serverPath,
        type: "stdio"
    };

    var clientOptions = {
        syntaxes: ["djot"]
    };

    langClient = new LanguageClient(
        "io.dwk.djot",
        "Djot Language Server",
        serverOptions,
        clientOptions
    );

    langClient.onNotification("djot/previewServer", function(params) {
        previewPort = params.port;
    });

    langClient.start();

    nova.commands.register("io.dwk.djot.preview", function() {
        if (previewPort === null) {
            nova.workspace.showWarningMessage(
                "Preview server not ready. Open a .dj file first."
            );
            return;
        }
        var editor = nova.workspace.activeTextEditor;
        if (!editor || !editor.document || !editor.document.path) {
            nova.workspace.showWarningMessage(
                "Open a .dj file to preview it."
            );
            return;
        }
        var workRoot = nova.workspace.path;
        var docPath = editor.document.path;
        var relPath = docPath;
        if (workRoot && docPath.startsWith(workRoot)) {
            relPath = docPath.substring(workRoot.length);
        }
        if (!relPath.startsWith("/")) {
            relPath = "/" + relPath;
        }
        nova.openURL("http://localhost:" + previewPort + relPath);
    });

    nova.commands.register("io.dwk.djot.toggleEmphasis", function(editor) {
        wrapSelection(editor, "_", "_");
    });

    nova.commands.register("io.dwk.djot.toggleStrong", function(editor) {
        wrapSelection(editor, "*", "*");
    });

    nova.commands.register("io.dwk.djot.toggleInlineCode", function(editor) {
        wrapSelection(editor, "`", "`");
    });

    nova.commands.register("io.dwk.djot.toggleHighlight", function(editor) {
        wrapSelection(editor, "{=", "=}");
    });

    nova.commands.register("io.dwk.djot.toggleBlockquote", function(editor) {
        prependLines(editor, "> ");
    });

    nova.commands.register("io.dwk.djot.insertLink", function(editor) {
        var firstRange = editor.selectedRanges[0];
        var selectedText = editor.document.getTextInRange(firstRange);
        var insertText = "[" + selectedText + "](url)";

        editor.edit(function(e) {
            e.replace(firstRange, insertText);
        });

        var urlStart = firstRange.start + 1 + selectedText.length + 2;
        var urlEnd = urlStart + 3;
        editor.selectedRanges = [new Range(urlStart, urlEnd)];
    });
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

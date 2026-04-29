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
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

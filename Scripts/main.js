var langClient = null;
var previewPort = null;
var scrollTimer = null;
var editorDisposable = null;

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

    langClient.onNotification("djot/previewReady", function(params) {
        previewPort = params.port;
    });

    langClient.start();

    nova.commands.register("io.dwk.djot.preview", function(editor) {
        if (!previewPort) {
            nova.workspace.showWarningMessage("Preview server not ready yet. Try again in a moment.");
            return;
        }
        nova.openURL("http://localhost:" + previewPort + "/preview");
        startScrollSync(editor);
    });
};

exports.deactivate = function() {
    stopScrollSync();
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

function startScrollSync(initialEditor) {
    stopScrollSync();

    // Hook the current editor immediately if it is a djot file
    if (initialEditor && initialEditor.document && initialEditor.document.syntax === "djot") {
        hookEditor(initialEditor);
    }

    // Also hook any future active editors
    editorDisposable = nova.workspace.onDidChangeActiveTextEditor(function(editor) {
        if (!editor) return;
        if (editor.document.syntax !== "djot") return;
        hookEditor(editor);
    });
}

function hookEditor(editor) {
    editor.onDidChangeSelection(function() {
        if (!previewPort || !langClient) return;

        // Debounce: wait 100ms after the last cursor move
        if (scrollTimer) clearTimeout(scrollTimer);
        scrollTimer = setTimeout(function() {
            var cursorPos = editor.selectedRange.start;
            var fullText = editor.document.getTextInRange(new Range(0, cursorPos));
            var lineNum = 0;
            for (var i = 0; i < fullText.length; i++) {
                if (fullText.charCodeAt(i) === 10) lineNum++;
            }
            langClient.sendNotification("djot/scrollTo", { line: lineNum });
        }, 100);
    });
}

function stopScrollSync() {
    if (editorDisposable) {
        editorDisposable.dispose();
        editorDisposable = null;
    }
    if (scrollTimer) {
        clearTimeout(scrollTimer);
        scrollTimer = null;
    }
}

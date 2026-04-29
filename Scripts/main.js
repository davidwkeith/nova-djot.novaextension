var langClient = null;
var previewPort = null;

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
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

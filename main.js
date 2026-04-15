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

    nova.commands.register("io.dwk.djot.preview", function(editor) {
        if (!previewPort) {
            nova.workspace.showWarningMessage(
                "Preview server not ready. Make sure a .dj file is open."
            );
            return;
        }

        var url = "http://localhost:" + previewPort;
        if (editor && editor.document && editor.document.path) {
            var docPath = editor.document.path;
            var workRoot = nova.workspace.path;
            if (workRoot && docPath.startsWith(workRoot)) {
                url += docPath.substring(workRoot.length);
            }
        }
        nova.openURL(url);
    });
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

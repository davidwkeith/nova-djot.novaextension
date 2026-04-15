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
        console.log("djot: preview server ready on port " + previewPort);

        // Try to set Nova's preview URL to our server
        try {
            nova.workspace.config.set("preview.url", "http://localhost:" + previewPort);
            console.log("djot: set preview.url");
        } catch (e) {
            console.error("djot: failed to set preview.url: " + e);
        }
    });

    langClient.start();

    // Register our own preview command as fallback
    nova.commands.register("io.dwk.djot.preview", function(editor) {
        if (!previewPort) {
            console.warn("djot: preview port not set, LSP notification may not have arrived");
            nova.workspace.showWarningMessage(
                "Preview server not ready. Make sure a .dj file is open."
            );
            return;
        }

        // Open in system browser
        var url = "http://localhost:" + previewPort + "/";
        if (editor && editor.document && editor.document.path) {
            // Build the path relative to workspace root
            var docPath = editor.document.path;
            var workRoot = nova.workspace.path;
            if (workRoot && docPath.startsWith(workRoot)) {
                url += docPath.substring(workRoot.length);
            } else {
                url += docPath;
            }
        }
        console.log("djot: opening preview at " + url);
        nova.openURL(url);
    });
};

exports.deactivate = function() {
    // Clean up preview URL
    try {
        nova.workspace.config.remove("preview.url");
    } catch (e) {}

    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

var langClient = null;
var previewPort = null;

console.log("djot: main.js loaded");

exports.activate = function() {
    console.log("djot: activate() called");

    // Log current preview state for debugging
    console.log("djot: workspace.path = " + nova.workspace.path);
    console.log("djot: workspace.previewURL = " + (nova.workspace.previewURL || "null"));
    console.log("djot: workspace.previewRootPath = " + (nova.workspace.previewRootPath || "null"));

    // Log all config keys we can find related to preview
    var keysToCheck = [
        "preview.url", "previewURL", "preview.server",
        "nova.preview.url", "nova.preview.server",
        "nova.workspace.preview.url"
    ];
    keysToCheck.forEach(function(key) {
        var val = nova.workspace.config.get(key);
        if (val) {
            console.log("djot: config[" + key + "] = " + val);
        }
    });

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

        // Try various config keys to set Nova's preview URL
        var url = "http://localhost:" + previewPort;
        var keys = [
            "preview.url", "previewURL", "preview.server",
            "nova.preview.url", "nova.preview.server"
        ];
        keys.forEach(function(key) {
            try {
                nova.workspace.config.set(key, url);
                console.log("djot: set config[" + key + "] = " + url);
            } catch (e) {
                console.error("djot: failed config[" + key + "]: " + e);
            }
        });

        // Log preview URL after setting
        console.log("djot: workspace.previewURL after set = " + (nova.workspace.previewURL || "null"));
    });

    langClient.start();

    // Register preview command (opens in browser as fallback)
    nova.commands.register("io.dwk.djot.preview", function(editor) {
        if (!previewPort) {
            nova.workspace.showWarningMessage(
                "Preview server not ready. Make sure a .dj file is open."
            );
            return;
        }

        var url = "http://localhost:" + previewPort + "/";
        if (editor && editor.document && editor.document.path) {
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
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewPort = null;
};

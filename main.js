var langClient = null;
var lastPreviewPath = null;

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

    langClient.onNotification("djot/previewFile", function(params) {
        lastPreviewPath = params.path;
    });

    langClient.start();

    nova.commands.register("io.dwk.djot.preview", function() {
        if (!lastPreviewPath) {
            nova.workspace.showWarningMessage(
                "Preview not ready. Open a .dj file first."
            );
            return;
        }
        nova.openURL("file://" + lastPreviewPath);
    });
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    lastPreviewPath = null;
};

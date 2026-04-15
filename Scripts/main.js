var langClient = null;
var previewFilePath = null;

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
        previewFilePath = params.path;
    });

    langClient.start();

    nova.commands.register("io.dwk.djot.preview", function(editor) {
        if (!previewFilePath) {
            nova.workspace.showWarningMessage("Preview not ready yet. Try again in a moment.");
            return;
        }
        nova.openURL("file://" + previewFilePath);
    });
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
    previewFilePath = null;
};

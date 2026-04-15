var langClient = null;
var previewPath = nova.path.join(nova.environment.home, "tmp", "djot-preview.html");

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

    langClient.start();

    nova.commands.register("io.dwk.djot.preview", function(editor) {
        // The LSP server writes rendered HTML to /tmp/djot-preview.html
        // on every document change. Open it in the default browser.
        nova.openURL("file:///tmp/djot-preview.html");
    });
};

exports.deactivate = function() {
    if (langClient) {
        langClient.stop();
        langClient = null;
    }
};

GRAMMAR_REPO = https://github.com/treeman/tree-sitter-djot.git
GRAMMAR_DIR  = tree-sitter-djot
SRC_DIR      = $(GRAMMAR_DIR)/src
DYLIB        = Syntaxes/libtree-sitter-djot.dylib

CC           = cc
CFLAGS       = -shared -Os -fPIC -I $(SRC_DIR)

LSP_DIR    = djot-lsp
LSP_BINARY = bin/djot-lsp

.PHONY: build clean lsp all

build: $(DYLIB)

$(DYLIB): $(SRC_DIR)/parser.c $(SRC_DIR)/scanner.c
	$(CC) -arch arm64 $(CFLAGS) -o /tmp/libtree-sitter-djot-arm64.dylib $(SRC_DIR)/parser.c $(SRC_DIR)/scanner.c
	$(CC) -arch x86_64 $(CFLAGS) -o /tmp/libtree-sitter-djot-x86_64.dylib $(SRC_DIR)/parser.c $(SRC_DIR)/scanner.c
	lipo -create /tmp/libtree-sitter-djot-arm64.dylib /tmp/libtree-sitter-djot-x86_64.dylib -output $(DYLIB)
	rm -f /tmp/libtree-sitter-djot-arm64.dylib /tmp/libtree-sitter-djot-x86_64.dylib

$(SRC_DIR)/parser.c:
	git clone --depth 1 $(GRAMMAR_REPO) $(GRAMMAR_DIR)

lsp: $(LSP_BINARY)

$(LSP_BINARY): $(LSP_DIR)/main.go $(LSP_DIR)/server/*.go
	mkdir -p bin
	cd $(LSP_DIR) && GOOS=darwin GOARCH=arm64 go build -o /tmp/djot-lsp-arm64 .
	cd $(LSP_DIR) && GOOS=darwin GOARCH=amd64 go build -o /tmp/djot-lsp-amd64 .
	lipo -create /tmp/djot-lsp-arm64 /tmp/djot-lsp-amd64 -output $(LSP_BINARY)
	rm -f /tmp/djot-lsp-arm64 /tmp/djot-lsp-amd64

all: build lsp

clean:
	rm -rf $(GRAMMAR_DIR)
	rm -f $(DYLIB)
	rm -f $(LSP_BINARY)

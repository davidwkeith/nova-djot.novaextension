GRAMMAR_REPO = https://github.com/treeman/tree-sitter-djot.git
GRAMMAR_DIR  = tree-sitter-djot
SRC_DIR      = $(GRAMMAR_DIR)/src
DYLIB        = Syntaxes/libtree-sitter-djot.dylib

CC           = cc
CFLAGS       = -shared -Os -fPIC -I $(SRC_DIR)

.PHONY: build clean

build: $(DYLIB)

$(DYLIB): $(SRC_DIR)/parser.c $(SRC_DIR)/scanner.c
	$(CC) -arch arm64 $(CFLAGS) -o /tmp/libtree-sitter-djot-arm64.dylib $(SRC_DIR)/parser.c $(SRC_DIR)/scanner.c
	$(CC) -arch x86_64 $(CFLAGS) -o /tmp/libtree-sitter-djot-x86_64.dylib $(SRC_DIR)/parser.c $(SRC_DIR)/scanner.c
	lipo -create /tmp/libtree-sitter-djot-arm64.dylib /tmp/libtree-sitter-djot-x86_64.dylib -output $(DYLIB)
	rm -f /tmp/libtree-sitter-djot-arm64.dylib /tmp/libtree-sitter-djot-x86_64.dylib

$(SRC_DIR)/parser.c:
	git clone --depth 1 $(GRAMMAR_REPO) $(GRAMMAR_DIR)

clean:
	rm -rf $(GRAMMAR_DIR)
	rm -f $(DYLIB)

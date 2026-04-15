**Djot** adds full [Djot](https://djot.net/) markup language support to [Nova](https://nova.app), including syntax highlighting, an integrated language server, and code intelligence features.

## Getting Started

1. Install the extension from the Nova Extension Library
2. Open or create a file with the `.dj` extension
3. The language server starts automatically when a Djot file is opened

## Syntax Highlighting

All Djot constructs are highlighted using a [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) grammar, providing accurate, parse-tree-based coloring:

- **Block-level** — headings (H1-H6), paragraphs, fenced code blocks, raw blocks, block quotes, ordered/bullet/task/definition lists, tables, divs, thematic breaks, footnote definitions, frontmatter
- **Inline-level** — emphasis, strong, insert, delete, highlight, superscript, subscript, inline code, math, links, images, autolinks, footnote references, smart punctuation, symbols, raw inline, spans with attributes
- **Attributes** — classes, identifiers, key-value pairs highlighted individually

## Code Folding

Collapse and expand document sections in the editor gutter:

- Sections (heading + content through next same-or-higher-level heading)
- Fenced code blocks and raw blocks
- Lists (bullet, ordered, task, definition)
- Fenced divs, block quotes, tables, footnote definitions

## Language Injection

Embedded languages inside Djot are highlighted with their own syntax when the corresponding Nova extension is installed:

- Fenced code blocks with a language tag (e.g., ` ```python `)
- Raw blocks with a format specifier (e.g., `::: =html`)
- Raw inline with a format attribute (e.g., `` `<b>bold</b>`{=html} ``)
- Math expressions (injected as LaTeX)
- Frontmatter with a language tag (e.g., `---toml`)

## Language Server

A built-in language server provides code intelligence without any external dependencies:

### Completions

Type `[^` to get a list of footnote labels defined in the document. Type `][` after link text to complete reference link labels. Type `{#` to complete heading IDs.

### Diagnostics

Warnings appear in the editor for:

- Undefined footnote references (e.g., `[^missing]` with no matching definition)
- Undefined link references (e.g., `[text][nowhere]` with no matching definition)
- Duplicate footnote or reference definitions

### Document Symbols

The Symbols sidebar shows a nested heading outline for the document. H2 headings nest under H1, H3 under H2, and so on.

### Hover

Hover over a footnote reference to see the footnote's content. Hover over a reference link to see its destination URL. Hover over an inline link to see the URL.

### Go to Definition

Cmd-click (or right-click > Jump to Definition) on a footnote reference to jump to its `[^label]:` definition. Works the same for reference links.

## Live Preview

Run **Editor > Preview Djot** (or find it in the command palette) to open a live HTML preview of the current document.

The preview updates in real-time as you type. Scroll sync keeps the preview aligned with your cursor position in the editor — move to a heading in the source and the preview scrolls to match.

The preview runs on a local HTTP server (random port on localhost) and connects via WebSocket for instant updates. No external services or network access required. Supports both light and dark mode.

## Building from Source

The extension includes pre-compiled binaries, but if you need to rebuild:

```sh
# Build the Tree-sitter grammar (requires Xcode CLI tools)
make build

# Build the language server (requires Go 1.23+)
make lsp

# Build both
make all

## TODO

- [ ] Live preview
- [ ] JavaScript scripts or commands
- [ ] Extension settings or configuration
- [ ] Djot-to-HTML export
```

## Credits

- Syntax: [tree-sitter-djot](https://github.com/treeman/tree-sitter-djot) by treeman (MIT)
- Parser: [godjot](https://github.com/sivukhin/godjot) by sivukhin (MIT)

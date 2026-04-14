**Djot** provides syntax highlighting, code folding, symbol navigation, and language intelligence for the [Djot](https://djot.net/) markup language.

## Features

### Syntax
- Full syntax highlighting for all Djot constructs
- Code folding for sections, code blocks, lists, tables, and more
- Language injection in fenced code blocks, raw blocks, math (LaTeX), and frontmatter

### Language Server
- **Completions** — footnote labels, reference link labels, heading IDs
- **Diagnostics** — warns on undefined footnote/link references and duplicate definitions
- **Document Symbols** — heading outline in the Symbols sidebar
- **Hover** — shows footnote content, link destinations
- **Go to Definition** — jump from reference to its definition

## Supported Syntax

### Block-Level
Headings, paragraphs, code blocks, raw blocks, block quotes, ordered lists, bullet lists, task lists, definition lists, tables, divs, thematic breaks, footnotes, frontmatter

### Inline-Level
Emphasis, strong, insert, delete, highlight, superscript, subscript, inline code, math, links, images, autolinks, footnote references, smart punctuation, symbols, raw inline, spans with attributes

## File Extension

Djot files use the `.dj` extension.

## Credits

- Syntax highlighting: [tree-sitter-djot](https://github.com/treeman/tree-sitter-djot) by treeman (MIT)
- Language server parser: [godjot](https://github.com/sivukhin/godjot) by sivukhin (MIT)


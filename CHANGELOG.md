## Version 1.1.0

### Editor Commands
- Toggle Emphasis, Toggle Strong, Toggle Inline Code, Toggle Highlight — wrap or unwrap inline delimiters
- Toggle Blockquote — prepend or strip `> ` on selected lines
- Insert Link, Insert Fenced Code Block, Insert Table — insert templates with the cursor placed on the most-likely-edited slot

## Version 1.0.0

### Syntax
- Syntax highlighting for all Djot block and inline constructs
- Code folding for sections, code blocks, lists, tables, divs, block quotes, footnotes
- Heading navigation in the Symbols sidebar
- Language injection for fenced code blocks, raw blocks/inline, math (LaTeX), and frontmatter
- Based on tree-sitter-djot grammar v2.0.0

### Language Server
- Completions for footnote labels, reference link labels, and heading IDs
- Diagnostics for undefined references and duplicate definitions
- Document symbols (heading outline) in the Symbols sidebar
- Hover information for footnotes, link references, and inline links
- Go-to-definition for footnote and link references
- Powered by godjot parser

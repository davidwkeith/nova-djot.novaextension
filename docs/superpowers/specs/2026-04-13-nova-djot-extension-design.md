# Nova Djot Extension Design Spec

**Date**: 2026-04-13
**Status**: Approved

## Overview

A Nova editor extension providing Djot markup language support (`.dj` files) via the existing `tree-sitter-djot` grammar. Pure syntax extension with no JavaScript runtime, language server, or commands.

## Extension Metadata

| Field | Value |
|---|---|
| Identifier | `io.dwk.djot` |
| Name | Djot |
| Organization | dwk |
| Version | 1.0.0 |
| Categories | languages |
| License | ICS |

## File Layout

```
nova-djot.novaextension/
├── extension.json
├── Makefile
├── CHANGELOG.md
├── README.md
├── .gitignore
├── Images/
│   └── extension/
│       ├── extension.png         # 128x128
│       └── extension@2x.png     # 256x256
├── Syntaxes/
│   ├── Djot.xml                  # Syntax definition
│   └── libtree-sitter-djot.dylib # Compiled grammar (universal arm64+x86_64)
└── Queries/
    ├── highlights.scm
    ├── folds.scm
    ├── symbols.scm
    └── injections.scm
```

## Syntax Definition (`Djot.xml`)

- `<type>markup</type>` — enables soft wrap, markup-appropriate editor behavior
- Detects `.dj` file extension with priority 1.0
- Indentation: auto-indent continuation after list markers (`^\s*([-*+]|\d+[.)]) `)
- Comments: multiline `{% ... %}` (Djot's comment syntax)
- Tree-sitter language name: `djot`
- References all four query files: highlights, folds, symbols, injections

## Highlights Query

Translated from `treeman/tree-sitter-djot` Neovim queries to Nova capture names.

### Capture Mapping

| Djot Construct | Neovim Capture | Nova Capture |
|---|---|---|
| Headings | `@markup.heading` | `@markup.heading` |
| Heading levels 1-6 | `@markup.heading.1` … `.6` | `@markup.heading.1` … `.6` |
| Emphasis | `@markup.italic` | `@markup.italic` |
| Strong | `@markup.strong` | `@markup.bold` |
| Strikethrough/Delete | `@markup.strikethrough` | `@markup.strikethrough` |
| Underline/Insert | `@markup.underline` | `@markup.underline` |
| Code blocks | `@markup.raw.block` | `@markup.code` |
| Inline code | `@markup.raw` | `@markup.code` |
| Math | `@markup.math` | `@value.number` |
| Links | `@markup.link.url` | `@markup.link` |
| Link text | `@markup.link` | `@markup.link.text` |
| Link labels | `@markup.link.label` | `@markup.link.label` |
| Image descriptions | `@markup.italic` | `@markup.image` |
| Block quotes | `@markup.quote` | `@markup.quote` |
| List markers | `@markup.list` | `@markup.list.marker` |
| Task checked | `@constant.builtin` | `@markup.list.checked` |
| Task unchecked | `@markup.list.unchecked` | `@markup.list.unchecked` |
| Comments | `@comment` | `@comment` |
| Thematic break | `@string.special` | `@markup.separator` |
| Smart punctuation | `@string.special` | `@string.special` |
| Escape sequences | `@string.escape` | `@string.escape` |
| Attributes `{}` | various | `@processing.attribute` |
| Class names | `@type` | `@identifier.type` |
| Identifiers `#id` | `@tag` | `@identifier.decorator` |
| Key in key=value | `@property` | `@identifier.property` |
| Value in key=value | `@string` | `@string` |
| Language spec | `@attribute` | `@identifier.decorator` |
| Table headers | `@markup.heading` | `@markup.heading` |
| Table separators | `@punctuation.special` | `@punctuation.separator` |
| Table pipes | `@punctuation.special` | `@punctuation.separator` |
| Brackets | `@punctuation.bracket` | `@punctuation.bracket` |
| Delimiters | `@punctuation.delimiter` | `@punctuation.delimiter` |
| Footnote labels | `@markup.link.label` | `@markup.link.label` |
| Definition terms | `@type.definition` | `@identifier.type` |
| Symbols | `@string.special.symbol` | `@string.special` |
| Highlight/super/sub | `@string.special` | `@string.special` |
| TODO | `@comment.todo` | `@comment.todo` |
| NOTE | `@comment.note` | `@comment` |
| FIXME | `@comment.error` | `@comment.todo` |
| Div markers | `@punctuation.delimiter` | `@punctuation.delimiter` |
| Block quote markers | `@punctuation.special` | `@punctuation.special` |
| Caption markers | `@punctuation.special` | `@punctuation.special` |
| Caption text | `@markup.italic` | `@markup.italic` |
| Frontmatter markers | `@punctuation.delimiter` | `@punctuation.delimiter` |

### Directives Dropped (Not Supported by Nova)

- `#set! conceal` — Nova has no character concealing
- `#set! priority` — Nova resolves by specificity
- `#offset!` — Nova doesn't support offset predicates
- `@spell` / `@nospell` — Nova handles spell checking separately

## Folds Query

Foldable regions:

| Node | Purpose |
|---|---|
| `section` | Heading + all content until next same-or-higher heading |
| `code_block` | Fenced code blocks |
| `raw_block` | Raw format blocks (`::: =html`) |
| `list` | Bullet, ordered, task, and definition lists |
| `div` | Fenced div containers |
| `block_quote` | Block quotations |
| `table` | Tables |
| `footnote` | Footnote definitions |

## Symbols Query

Heading-based navigation in Nova's Symbols sidebar:

- `(heading)` marked as `@symbol` — appears in navigator
- Heading text content captured as `@name` — display text
- `(section)` marked as `@subtree` — creates nesting hierarchy (H2 under H1, H3 under H2, etc.)

## Injections Query

Five language injection cases:

| Construct | Djot Syntax | Grammar Path |
|---|---|---|
| Fenced code | `` ```python `` | `code_block` → `(language)` + `(code)` |
| Raw blocks | `::: =html` | `raw_block` → `raw_block_info(language)` + `(content)` |
| Raw inline | `` `<b>text</b>`{=html} `` | `raw_inline` → `(content)` + `raw_inline_attribute(language)` |
| Math | `$x^2$` | `math(content)` → hardcoded `latex` |
| Frontmatter | `---toml` | `frontmatter` → `(language)` + `(frontmatter_content)` |

## Build Process

### Dependencies

- `tree-sitter` CLI (or `npx tree-sitter` from grammar's devDependencies)
- Xcode Command Line Tools (`cc`, `lipo`)
- Node.js + npm (for `tree-sitter generate`)

### Steps

1. Clone `treeman/tree-sitter-djot`
2. `npm install && npx tree-sitter generate` (produces `src/parser.c` + `src/scanner.c`)
3. Compile for arm64 and x86_64 separately
4. `lipo -create` to produce universal binary
5. Copy to `Syntaxes/libtree-sitter-djot.dylib`

### Makefile

Automates the full clone → generate → compile → lipo pipeline.

### Git Strategy

- Compiled `.dylib` is **committed** (Nova extensions are static bundles)
- `tree-sitter-djot/` clone directory is in `.gitignore`
- Grammar version tracked in `CHANGELOG.md`

## Out of Scope

- Language server / completions / diagnostics
- Live preview
- JavaScript scripts or commands
- Extension settings or configuration
- Djot-to-HTML export

## Dependencies

| Dependency | Version | License | Purpose |
|---|---|---|---|
| `tree-sitter-djot` | latest (v2.0.0) | MIT | Parser grammar |

## Supported Djot Features

### Block-Level (13 types)
Paragraph, Heading (1-6), Thematic Break, Section, Div, Code Block, Raw Block, Block Quote, Ordered List, Bullet List, Task List, Definition List, Table

### Inline-Level (25 types)
Str, Soft/Hard Break, Non-Breaking Space, Emphasis, Strong, Mark/Highlight, Superscript, Subscript, Insert, Delete, Double/Single Quoted, Link, Image, Span, Verbatim (inline code), Raw Inline, Inline/Display Math, URL/Email autolinks, Footnote Reference, Smart Punctuation, Symbol

### Extras (from tree-sitter-djot)
Frontmatter (TOML/YAML), TODO/NOTE/FIXME markers, tight sublists

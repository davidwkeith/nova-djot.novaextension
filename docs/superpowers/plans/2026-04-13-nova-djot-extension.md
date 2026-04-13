# Nova Djot Extension Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Nova editor extension providing Djot markup language support via the tree-sitter-djot grammar.

**Architecture:** Compile the existing tree-sitter-djot grammar into a universal macOS dylib, write Nova-compatible Tree-sitter query files for highlighting/folds/symbols/injections, and wire everything together with a syntax XML and extension manifest. No JavaScript, no language server — pure declarative syntax extension.

**Tech Stack:** Tree-sitter (C parser), Nova extension APIs (XML + .scm queries), macOS toolchain (cc, lipo)

**Spec:** `docs/superpowers/specs/2026-04-13-nova-djot-extension-design.md`

---

## File Map

| Action | File | Responsibility |
|---|---|---|
| Rewrite | `extension.json` | Extension manifest with correct metadata |
| Delete | `Syntaxes/djoit.xml` | Remove scaffolded placeholder |
| Create | `Syntaxes/Djot.xml` | Syntax definition, file detection, tree-sitter config |
| Create | `Syntaxes/libtree-sitter-djot.dylib` | Compiled universal grammar binary |
| Rewrite | `Queries/highlights.scm` | Syntax highlighting captures (Nova conventions) |
| Rewrite | `Queries/folds.scm` | Foldable region definitions |
| Rewrite | `Queries/symbols.scm` | Sidebar heading navigation |
| Create | `Queries/injections.scm` | Language injection for code blocks, raw content, math |
| Create | `Makefile` | Build automation for grammar compilation |
| Update | `.gitignore` | Ignore tree-sitter-djot clone directory |
| Rewrite | `README.md` | User-facing extension documentation |
| Update | `CHANGELOG.md` | Version 1.0.0 release notes |

---

### Task 1: Update Extension Metadata

**Files:**
- Modify: `extension.json`
- Modify: `.gitignore`

- [ ] **Step 1: Rewrite extension.json**

Replace the entire contents of `extension.json` with:

```json
{
    "identifier": "io.dwk.djot",
    "name": "Djot",
    "organization": "dwk",
    "description": "Djot markup language support with syntax highlighting, code folding, and symbol navigation.",
    "version": "1.0.0",
    "categories": ["languages"],
    "license": "MIT"
}
```

- [ ] **Step 2: Update .gitignore**

Replace the contents of `.gitignore` with:

```
.DS_Store
tree-sitter-djot/
```

- [ ] **Step 3: Commit**

```bash
git add extension.json .gitignore
git commit -m "chore: update extension metadata for Djot (io.dwk.djot)"
```

---

### Task 2: Compile Tree-sitter Grammar

**Files:**
- Create: `Makefile`
- Create: `Syntaxes/libtree-sitter-djot.dylib`
- Delete: `Syntaxes/djoit.xml`

- [ ] **Step 1: Delete the placeholder syntax file**

```bash
rm Syntaxes/djoit.xml
```

- [ ] **Step 2: Create the Makefile**

Create `Makefile` with:

```makefile
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
```

- [ ] **Step 3: Run the build**

```bash
make build
```

Expected: `Syntaxes/libtree-sitter-djot.dylib` is created. Verify:

```bash
file Syntaxes/libtree-sitter-djot.dylib
```

Expected output should contain: `Mach-O universal binary with 2 architectures: [x86_64:Mach-O 64-bit dynamically linked shared library x86_64] [arm64]`

- [ ] **Step 4: Commit**

```bash
git add Makefile Syntaxes/libtree-sitter-djot.dylib
git rm Syntaxes/djoit.xml
git commit -m "feat: compile tree-sitter-djot grammar to universal dylib"
```

---

### Task 3: Create Syntax Definition

**Files:**
- Create: `Syntaxes/Djot.xml`

- [ ] **Step 1: Create Djot.xml**

Create `Syntaxes/Djot.xml` with:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<syntax name="djot">
    <meta>
        <name>Djot</name>
        <type>markup</type>
        <preferred-file-extension>dj</preferred-file-extension>
    </meta>

    <detectors>
        <extension priority="1.0">dj</extension>
    </detectors>

    <indentation>
        <increase>
            <expression>^\s*([-*+]|\d+[.)]) </expression>
        </increase>
        <decrease>
            <expression></expression>
        </decrease>
    </indentation>

    <comments>
        <multiline>
            <starts-with>
                <expression>\{%</expression>
            </starts-with>
            <ends-with>
                <expression>%\}</expression>
            </ends-with>
        </multiline>
    </comments>

    <tree-sitter language="djot">
        <highlights path="highlights.scm" />
        <folds path="folds.scm" />
        <symbols path="symbols.scm" />
        <injections path="injections.scm" />
    </tree-sitter>
</syntax>
```

- [ ] **Step 2: Commit**

```bash
git add Syntaxes/Djot.xml
git commit -m "feat: add Djot syntax definition with .dj file detection"
```

---

### Task 4: Write Highlights Query

**Files:**
- Modify: `Queries/highlights.scm`

- [ ] **Step 1: Write the highlights query**

Replace the empty `Queries/highlights.scm` with:

```scheme
; Headings
(heading) @markup.heading

((heading
  (marker) @_heading.marker) @markup.heading.1
  (#eq? @_heading.marker "# "))

((heading
  (marker) @_heading.marker) @markup.heading.2
  (#eq? @_heading.marker "## "))

((heading
  (marker) @_heading.marker) @markup.heading.3
  (#eq? @_heading.marker "### "))

((heading
  (marker) @_heading.marker) @markup.heading.4
  (#eq? @_heading.marker "#### "))

((heading
  (marker) @_heading.marker) @markup.heading.5
  (#eq? @_heading.marker "##### "))

((heading
  (marker) @_heading.marker) @markup.heading.6
  (#eq? @_heading.marker "###### "))

; Thematic break
(thematic_break) @markup.separator

; Div markers
[
  (div_marker_begin)
  (div_marker_end)
] @punctuation.delimiter

; Code blocks and raw blocks
[
  (code_block)
  (raw_block)
  (frontmatter)
] @markup.code

; Code with a language spec — let injection handle content highlighting
(code_block
  .
  (code_block_marker_begin)
  (language)
  (code) @_none)

[
  (code_block_marker_begin)
  (code_block_marker_end)
  (raw_block_marker_begin)
  (raw_block_marker_end)
] @punctuation.delimiter

; Language specifier on code blocks
(language) @identifier.decorator

(language_marker) @punctuation.delimiter

; Block quotes
(block_quote) @markup.quote

(block_quote_marker) @punctuation.special

; Tables
(table_header) @markup.heading

(table_header
  "|" @punctuation.separator)

(table_row
  "|" @punctuation.separator)

(table_separator) @punctuation.separator

(table_caption
  (marker) @punctuation.special)

(table_caption) @markup.italic

; List markers
[
  (list_marker_dash)
  (list_marker_plus)
  (list_marker_star)
  (list_marker_definition)
  (list_marker_decimal_period)
  (list_marker_decimal_paren)
  (list_marker_decimal_parens)
  (list_marker_lower_alpha_period)
  (list_marker_lower_alpha_paren)
  (list_marker_lower_alpha_parens)
  (list_marker_upper_alpha_period)
  (list_marker_upper_alpha_paren)
  (list_marker_upper_alpha_parens)
  (list_marker_lower_roman_period)
  (list_marker_lower_roman_paren)
  (list_marker_lower_roman_parens)
  (list_marker_upper_roman_period)
  (list_marker_upper_roman_paren)
  (list_marker_upper_roman_parens)
] @markup.list.marker

; Task list markers
(list_marker_task
  (unchecked)) @markup.list.unchecked

(list_marker_task
  (checked)) @markup.list.checked

; Definition list terms
(list_item
  (term) @identifier.type)

; Smart punctuation
[
  (ellipsis)
  (en_dash)
  (em_dash)
  (quotation_marks)
] @string.special

; Escape sequences
[
  (hard_line_break)
  (backslash_escape)
] @string.escape

; Frontmatter
(frontmatter_marker) @punctuation.delimiter

; Inline formatting
(emphasis) @markup.italic

(strong) @markup.bold

(symbol) @string.special

(insert) @markup.underline

(delete) @markup.strikethrough

[
  (highlighted)
  (superscript)
  (subscript)
] @string.special

; Inline formatting delimiters
[
  (emphasis_begin)
  (emphasis_end)
  (strong_begin)
  (strong_end)
  (superscript_begin)
  (superscript_end)
  (subscript_begin)
  (subscript_end)
  (highlighted_begin)
  (highlighted_end)
  (insert_begin)
  (insert_end)
  (delete_begin)
  (delete_end)
  (verbatim_marker_begin)
  (verbatim_marker_end)
  (math_marker)
  (math_marker_begin)
  (math_marker_end)
  (raw_inline_attribute)
  (raw_inline_marker_begin)
  (raw_inline_marker_end)
] @punctuation.delimiter

; Math
(math) @value.number

; Inline code
(verbatim) @markup.code

; Raw inline
(raw_inline) @markup.code

; Comments
[
  (comment)
  (inline_comment)
] @comment

; Span brackets
(span
  [
    "["
    "]"
  ] @punctuation.bracket)

; Attributes
(inline_attribute
  [
    "{"
    "}"
  ] @punctuation.bracket)

(block_attribute
  [
    "{"
    "}"
  ] @punctuation.bracket)

[
  (class)
  (class_name)
] @identifier.type

(identifier) @identifier.decorator

(key_value
  "=" @operator)

(key_value
  (key) @identifier.property)

(key_value
  (value) @string)

; Links
(link_text
  [
    "["
    "]"
  ] @punctuation.bracket)

(autolink
  [
    "<"
    ">"
  ] @punctuation.bracket)

(inline_link
  (inline_link_destination) @markup.link)

(link_reference_definition
  ":" @punctuation.special)

(full_reference_link
  (link_text) @markup.link.text)

(full_reference_link
  (link_label) @markup.link.label)

(full_reference_link
  [
    "["
    "]"
  ] @punctuation.bracket)

(collapsed_reference_link
  "[]" @punctuation.bracket)

(collapsed_reference_link
  (link_text) @markup.link.text)

(inline_link
  (link_text) @markup.link.text)

; Images
(full_reference_image
  (link_label) @markup.link.label)

(full_reference_image
  [
    "["
    "]"
  ] @punctuation.bracket)

(collapsed_reference_image
  "[]" @punctuation.bracket)

(image_description
  [
    "!["
    "]"
  ] @punctuation.bracket)

(image_description) @markup.image

; Link reference definitions
(link_reference_definition
  [
    "["
    "]"
  ] @punctuation.bracket)

(link_reference_definition
  (link_label) @markup.link.label)

(inline_link_destination
  [
    "("
    ")"
  ] @punctuation.bracket)

[
  (autolink)
  (inline_link_destination)
  (link_destination)
  (link_reference_definition)
] @markup.link

; Footnotes
(footnote
  (reference_label) @markup.link.label)

(footnote_reference
  (reference_label) @markup.link.label)

[
  (footnote_marker_begin)
  (footnote_marker_end)
] @punctuation.bracket

; TODO/NOTE/FIXME
(todo) @comment.todo
(note) @comment
(fixme) @comment.todo
```

- [ ] **Step 2: Commit**

```bash
git add Queries/highlights.scm
git commit -m "feat: add Djot syntax highlighting query (Nova captures)"
```

---

### Task 5: Write Folds Query

**Files:**
- Modify: `Queries/folds.scm`

- [ ] **Step 1: Write the folds query**

Replace the empty `Queries/folds.scm` with:

```scheme
; Sections fold as heading + all content until next same-or-higher heading
(section) @fold

; Fenced code and raw blocks
(code_block) @fold
(raw_block) @fold

; Lists (bullet, ordered, task, definition)
(list) @fold

; Fenced div containers
(div) @fold

; Block quotations
(block_quote) @fold

; Tables
(table) @fold

; Footnote definitions
(footnote) @fold
```

- [ ] **Step 2: Commit**

```bash
git add Queries/folds.scm
git commit -m "feat: add Djot code folding query"
```

---

### Task 6: Write Symbols Query

**Files:**
- Modify: `Queries/symbols.scm`

- [ ] **Step 1: Write the symbols query**

Replace the empty `Queries/symbols.scm` with:

```scheme
; Headings appear in Nova's Symbols sidebar for document navigation
(heading
    (marker) @context.begin
    (_) @name
) @symbol

; Sections create nesting hierarchy (H2 under H1, H3 under H2, etc.)
(section) @subtree
```

- [ ] **Step 2: Commit**

```bash
git add Queries/symbols.scm
git commit -m "feat: add Djot symbol navigation for headings"
```

---

### Task 7: Write Injections Query

**Files:**
- Create: `Queries/injections.scm`

- [ ] **Step 1: Write the injections query**

Create `Queries/injections.scm` with:

```scheme
; Fenced code blocks with language tag (e.g., ```python)
(code_block
    (language) @injection.language
    (code) @injection.content)

; Raw blocks with format specifier (e.g., ::: =html)
(raw_block
    (raw_block_info
        (language) @injection.language)
    (content) @injection.content)

; Raw inline with format attribute (e.g., `<b>text</b>`{=html})
(raw_inline
    (content) @injection.content
    (raw_inline_attribute
        (language) @injection.language))

; Math expressions injected as LaTeX
(math
    (content) @injection.content
    (#set! injection.language "latex"))

; Frontmatter with language tag (e.g., ---toml)
(frontmatter
    (language) @injection.language
    (frontmatter_content) @injection.content)
```

- [ ] **Step 2: Commit**

```bash
git add Queries/injections.scm
git commit -m "feat: add language injection for code blocks, raw content, math, frontmatter"
```

---

### Task 8: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Rewrite README.md**

Replace the contents of `README.md` with:

```markdown
**Djot** provides syntax highlighting, code folding, and symbol navigation for the [Djot](https://djot.net/) markup language.

## Features

- Full syntax highlighting for all Djot constructs
- Code folding for sections, code blocks, lists, tables, and more
- Heading navigation in the Symbols sidebar
- Language injection in fenced code blocks, raw blocks, math (LaTeX), and frontmatter

## Supported Syntax

### Block-Level
Headings, paragraphs, code blocks, raw blocks, block quotes, ordered lists, bullet lists, task lists, definition lists, tables, divs, thematic breaks, footnotes, frontmatter

### Inline-Level
Emphasis, strong, insert, delete, highlight, superscript, subscript, inline code, math, links, images, autolinks, footnote references, smart punctuation, symbols, raw inline, spans with attributes

## File Extension

Djot files use the `.dj` extension.

## Grammar

This extension uses the [tree-sitter-djot](https://github.com/treeman/tree-sitter-djot) grammar by treeman (MIT license).
```

- [ ] **Step 2: Update CHANGELOG.md**

Replace the contents of `CHANGELOG.md` with:

```markdown
## Version 1.0.0

- Syntax highlighting for all Djot block and inline constructs
- Code folding for sections, code blocks, lists, tables, divs, block quotes, footnotes
- Heading navigation in the Symbols sidebar
- Language injection for fenced code blocks, raw blocks/inline, math (LaTeX), and frontmatter
- Based on tree-sitter-djot grammar v2.0.0
```

- [ ] **Step 3: Commit**

```bash
git add README.md CHANGELOG.md
git commit -m "docs: update README and CHANGELOG for v1.0.0"
```

---

### Task 9: Manual Testing in Nova

**Files:** None (verification only)

- [ ] **Step 1: Create a test Djot file**

Create a file called `test.dj` (outside the extension directory) with representative Djot content:

```djot
---toml
title = "Test Document"
---

# Heading 1

## Heading 2

A paragraph with _emphasis_, *strong*, +insert+, -delete-, {=highlight=}, ^superscript^, ~subscript~, and `inline code`.

A [link](https://example.com) and a ![image description](image.png).

An autolink: <https://djot.net/>

A footnote reference[^note].

[^note]: This is the footnote content.

> A block quote with
> multiple lines.

- Bullet list
- Second item
  - Nested item

1. Ordered list
2. Second item

- [ ] Task unchecked
- [x] Task checked

: Term
: Definition

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |

^ Table caption

$$
x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}
$$

Inline math: $E = mc^2$

```python
def hello():
    print("Hello from Djot!")
```

::: =html
<div class="custom">Raw HTML content</div>
:::

`<b>inline raw</b>`{=html}

{.note #important key="value"}
A paragraph with attributes.

::: warning
A fenced div.
:::

{% This is a Djot comment %}

Smart punctuation: "quotes" and 'single' and ... and -- and ---

A symbol: :smile:

A \*backslash escape\*.

TODO: something
NOTE: something else
FIXME: broken thing
```

- [ ] **Step 2: Activate the extension in Nova**

Open the extension project in Nova, then go to **Extensions > Activate Project as Extension**.

- [ ] **Step 3: Open test.dj and verify**

Open the test file. Check:
- Headings are highlighted and appear in the Symbols sidebar with nesting
- Emphasis, strong, and other inline formatting are visually distinct
- Code blocks show language-specific highlighting (Python)
- Raw blocks (`::: =html`) highlight the HTML content
- Math expressions are highlighted
- Comments are dimmed
- Lists, tables, block quotes are properly styled
- Folding works on sections, code blocks, lists, tables
- Frontmatter is highlighted

- [ ] **Step 4: Fix any issues found during testing**

Adjust query files as needed. Commit fixes.

1# Djot Editor Commands Design Spec

**Date**: 2026-04-29
**Status**: Approved
**Tracking**: GitHub issue #1 (first checkbox — JavaScript scripts or commands)

## Overview

Add eight editor commands to the Nova Djot extension that cover the most common Djot
editing operations: toggling inline formatting, inserting links and templates, and
prepending blockquote markers. All commands are implemented in pure JavaScript on top of
Nova's editor API and require no language-server or Go-side changes.

The example "Fold all sections" command from the issue is intentionally **out of scope**:
Nova's public JavaScript API does not expose programmatic code folding. Tree-sitter
folds.scm already provides gutter-based folding for sections.

## Architecture

All command code lives in `Scripts/main.js`, alongside the existing LSP boot and preview
command, separated by a comment banner. No new files, no `require()` calls, no build
step. This avoids the Nova `main`-field resolution gotcha captured in CLAUDE memory
(`main: "main.js"` resolves relative to `Scripts/`; multi-file `require()` paths are
fragile).

Two private helpers do the shared work:

- `wrapSelection(editor, open, close)` — toggles delimiters around each selected range.
- `prependLines(editor, prefix)` — toggles a line-prefix on each selected line.

Each public command is a thin wrapper that calls one helper or performs a direct
template insertion. Commands register inside the existing `exports.activate` after the
preview command. The extension's `*` activation event already fires on Nova startup, so
no manifest activation changes are needed.

## Components

### The eight commands

All gated by `when: editorSyntax == 'djot'` so they only appear in `.dj` files.

| Command ID | Title | Implementation |
|---|---|---|
| `io.dwk.djot.toggleEmphasis` | Toggle Emphasis | `wrapSelection(ed, "_", "_")` |
| `io.dwk.djot.toggleStrong` | Toggle Strong | `wrapSelection(ed, "*", "*")` |
| `io.dwk.djot.toggleInlineCode` | Toggle Inline Code | `wrapSelection(ed, "` + "`" + `", "` + "`" + `")` |
| `io.dwk.djot.toggleHighlight` | Toggle Highlight | `wrapSelection(ed, "{=", "=}")` |
| `io.dwk.djot.toggleBlockquote` | Toggle Blockquote | `prependLines(ed, "> ")` |
| `io.dwk.djot.insertLink` | Insert Link | direct insert `[text](url)`, cursor on `url` |
| `io.dwk.djot.insertCodeBlock` | Insert Fenced Code Block | direct insert: triple-backtick fence with `language` placeholder |
| `io.dwk.djot.insertTable` | Insert Table | direct insert: 2-column × 3-row starter table |

Each command appears in both `commands.editor` (Editor menu / command palette while a
Djot file is active) and `commands.extensions` (Extensions → Djot menu, regardless of
focus, matching the pattern used by the existing Preview command).

**No default keyboard shortcuts** are declared. Users bind their own via
Extensions → Extension Library → Djot → Settings → Keyboard Shortcuts.

## Data flow and toggle semantics

### `wrapSelection(editor, open, close)`

For each range in `editor.selectedRanges`, processed in reverse document order so
earlier edits do not shift later ranges:

1. **Empty range:** scan the current line for an enclosing `open…close` pair around
   the cursor (i.e., the nearest `open` to the left and the nearest `close` to the
   right, both on the same line). If found, replace the enclosing span with its
   inner text. Otherwise, insert `open + close` at the cursor and place the cursor
   between them by setting `editor.selectedRanges` after the edit resolves.
2. **Non-empty range — inner wrap detection:** if the selected text starts with `open`
   and ends with `close`, strip them and replace the range with the inner text.
3. **Non-empty range — outer wrap detection:** if the characters immediately *outside*
   the selection (i.e., the `open.length` chars before the range start and the
   `close.length` chars after the range end) match `open` and `close`, expand the
   replacement range to include them and replace with the inner text alone.
4. **Otherwise:** replace selection with `open + selectedText + close`.

All sub-edits for a single command invocation happen inside one `editor.edit(cb)`
callback so the user gets a single undo step. Multi-cursor editing is preserved by
iterating selections in reverse sort order.

### `prependLines(editor, prefix)`

1. Compute the union of lines touching any selected range (use document content plus
   the range start/end positions to find line boundaries).
2. **Toggle detection:** if *every non-empty* affected line already starts with
   `prefix`, strip the prefix from each line that has it (and leave blank lines
   alone). Otherwise, add `prefix` to every affected line that does not already start
   with it (do not double-prefix). Blank lines inside the selection get the prefix in
   "add" mode — `> ` on a blank line is valid Djot blockquote continuation.
3. Apply all line edits inside one `editor.edit(cb)` callback (one undo step).

### Insert commands (Link, Code Block, Table)

Single insertion at the first selected range. Selected text, if any, goes into a
sensible slot:

- **Insert Link:** `[<selected-or-text>](url)`. After insertion, set
  `editor.selectedRanges` so `url` is selected.
- **Insert Fenced Code Block:** triple-backtick fence with a `language` placeholder on
  the opening line; selected text becomes the body. After insertion, select the
  `language` placeholder.
- **Insert Table:** 2×3 starter table (header row, separator row, two body rows).
  Selected text goes into the first body cell. After insertion, select the first
  header cell text.

Templates use exact strings — no string-template engine, no escaping concerns.

## Error handling

Nova guarantees `editor` is non-null inside an editor command callback, so no null
check is needed. Empty documents still provide a zero-length range, which all
algorithms handle naturally (empty selection → insert delimiters; line-prefix on empty
line → prepend prefix to the empty line).

The commands are silent on success and silent on no-op. A toggle that flips state
*is* the success path, not an error. There are no `showWarningMessage` calls.

## Testing

The extension has no JS test harness today. Adding one (Jest, mocha) requires stubbing
the `nova.*` global — significant infrastructure for a small surface area. The plan is
**manual smoke testing via `test.dj`**, documented as a checklist in the implementation
plan:

- Toggle Emphasis on selected word → `_word_`
- Toggle Emphasis on `_word_` selected → `word`
- Toggle Emphasis with empty selection inside existing `_word_` → `word`
  (outer-wrap detection)
- Toggle Strong / Toggle Inline Code / Toggle Highlight — same three cases
- Toggle Blockquote on three selected lines → all three get `> `
- Toggle Blockquote on three already-quoted lines → all three lose `> `
- Toggle Blockquote on a mix of quoted and unquoted lines → all become quoted
- Insert Link with no selection → `[](url)` with `url` selected
- Insert Link with selected text → `[selected](url)` with `url` selected
- Insert Fenced Code Block, Insert Table → render correctly in the preview pane
- Multi-cursor Toggle Emphasis → all selections wrap independently
- Undo after any command → exactly one undo step restores original text

If a future regression surfaces, *that* is when we invest in a test harness.

## extension.json changes

Eight new entries in `commands.editor` (each with `when: editorSyntax == 'djot'`, no
`shortcut` field) and eight matching entries in `commands.extensions`. No new
top-level keys. Version bumps to `1.1.0` (additive feature, semver minor).

## Out of scope

- Fold all sections (no Nova JS API for programmatic folding).
- Insert template / new-document scaffolding (deferred; can ship later if requested).
- Heading level increment/decrement (deferred).
- Footnote auto-definition (deferred; LSP completions already cover the discovery
  side).
- Settings UI for command shortcuts (covered by issue #1's second checkbox, separate
  PR).
- Djot-to-HTML export command (issue #1's third checkbox, separate PR).

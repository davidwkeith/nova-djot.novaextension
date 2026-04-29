# Djot Editor Commands Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add eight Djot editor commands (toggle emphasis/strong/inline-code/highlight/blockquote, insert link/code-block/table) to the Nova Djot extension as pure-JavaScript additions to `Scripts/main.js`.

**Architecture:** Two private helpers (`wrapSelection`, `prependLines`) implement the shared toggle algorithms. Eight thin command handlers register inside the existing `exports.activate`. Both `extension.json` command slots (`commands.editor` and `commands.extensions`) gain matching entries. Single file change for the JS, single manifest change. No test harness — manual smoke testing per spec.

**Tech Stack:** Nova editor JavaScript API (CommonJS sandbox in JavaScriptCore), `extension.json` manifest, Nova's `TextEditor.edit(cb)` API for batched single-undo edits.

**Spec:** `docs/superpowers/specs/2026-04-29-djot-editor-commands-design.md`

---

## File Structure

| File | Action | Purpose |
|---|---|---|
| `extension.json` | Modify | Add 8 entries to `commands.editor`, 8 to `commands.extensions`; bump version to `1.1.0` |
| `Scripts/main.js` | Modify | Add `wrapSelection` and `prependLines` helpers; register 8 commands inside `exports.activate` |
| `CHANGELOG.md` | Modify | Add `1.1.0` entry listing the new commands |
| `README.md` | Modify | Add an "Editor Commands" section documenting the eight commands |
| `docs/TODO.md` | Modify | Mark "JavaScript scripts or commands" as in progress (deferred items have separate issues) |

No new files. Everything lives next to existing code.

---

## Task 1: Branch setup and version bump in manifest

**Files:**
- Modify: `extension.json`

- [ ] **Step 1: Create a feature branch**

```bash
git checkout -b feat/1-editor-commands
```

- [ ] **Step 2: Bump version in `extension.json` from `1.0.0` to `1.1.0`**

Open `extension.json` and change:

```json
    "version": "1.0.0",
```

to:

```json
    "version": "1.1.0",
```

- [ ] **Step 3: Verify manifest is still valid JSON**

Run:

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

Expected: no output (valid JSON). Non-zero exit means a syntax error.

- [ ] **Step 4: Commit**

```bash
git add extension.json
git commit -m "chore(#1): bump version to 1.1.0 for editor commands"
```

---

## Task 2: Add `wrapSelection` helper

**Files:**
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add the helper above `exports.activate`**

In `Scripts/main.js`, immediately after the two top-level `var` declarations (`var langClient = null;` and `var previewPort = null;`) and before `exports.activate`, insert:

```javascript
// ---------------------------------------------------------------------------
// Editor command helpers
// ---------------------------------------------------------------------------

// Toggle delimiters around each selected range. Three behaviors:
//   - Empty range: insert open+close, place cursor between them.
//   - Selection already wrapped (inner): strip delimiters from selected text.
//   - Selection's neighbors form delimiters (outer): expand range, strip them.
//   - Otherwise: wrap the selection.
// All edits batched into one undo step. Multi-cursor safe via reverse-order
// processing. Cursor repositioning only applied for the single-empty-range
// case (other multi-cursor combinations rely on Nova's default behavior).
function wrapSelection(editor, open, close) {
    var ranges = editor.selectedRanges.slice().sort(function(a, b) {
        return b.start - a.start;
    });
    var openLen = open.length;
    var closeLen = close.length;
    var docLen = editor.document.length;

    var soleEmpty = (ranges.length === 1 && ranges[0].start === ranges[0].end);
    var soleEmptyStart = soleEmpty ? ranges[0].start : -1;

    editor.edit(function(e) {
        for (var i = 0; i < ranges.length; i++) {
            var r = ranges[i];

            if (r.start === r.end) {
                e.insert(r.start, open + close);
                continue;
            }

            var selectedText = editor.document.getTextInRange(r);

            if (selectedText.length >= openLen + closeLen &&
                selectedText.substring(0, openLen) === open &&
                selectedText.substring(selectedText.length - closeLen) === close) {
                var inner = selectedText.substring(openLen, selectedText.length - closeLen);
                e.replace(r, inner);
                continue;
            }

            if (r.start >= openLen && r.end + closeLen <= docLen) {
                var outerRange = new Range(r.start - openLen, r.end + closeLen);
                var outerText = editor.document.getTextInRange(outerRange);
                if (outerText.substring(0, openLen) === open &&
                    outerText.substring(outerText.length - closeLen) === close) {
                    e.replace(outerRange, selectedText);
                    continue;
                }
            }

            e.replace(r, open + selectedText + close);
        }
    });

    if (soleEmpty) {
        var pos = soleEmptyStart + openLen;
        editor.selectedRanges = [new Range(pos, pos)];
    }
}
```

- [ ] **Step 2: Parse-check the file**

Run:

```bash
node --check Scripts/main.js
```

Expected: no output, exit code 0. Any syntax error halts here.

- [ ] **Step 3: Commit**

```bash
git add Scripts/main.js
git commit -m "feat(#1): add wrapSelection helper for delimiter toggles"
```

---

## Task 3: Register Toggle Emphasis (validates the wrapSelection helper end-to-end)

**Files:**
- Modify: `extension.json`
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add Toggle Emphasis to `extension.json`**

Inside `commands.editor`, after the existing Preview entry, add a comma and:

```json
            {
                "title": "Toggle Emphasis",
                "command": "io.dwk.djot.toggleEmphasis",
                "when": "editorSyntax == 'djot'"
            }
```

Inside `commands.extensions`, after the existing Preview entry, add a comma and:

```json
            {
                "title": "Toggle Emphasis",
                "command": "io.dwk.djot.toggleEmphasis"
            }
```

- [ ] **Step 2: Validate JSON**

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

Expected: no output.

- [ ] **Step 3: Register the command in `Scripts/main.js`**

Inside `exports.activate`, after the existing `nova.commands.register("io.dwk.djot.preview", ...)` block, add:

```javascript
    nova.commands.register("io.dwk.djot.toggleEmphasis", function(editor) {
        wrapSelection(editor, "_", "_");
    });
```

- [ ] **Step 4: Parse-check**

```bash
node --check Scripts/main.js
```

Expected: no output.

- [ ] **Step 5: Smoke-test in Nova**

Reload the extension (Extensions → Activate Project as Extension, or restart Nova if installed normally). Open `test.dj`. Verify all three behaviors:

- Select the word `Apple` somewhere in the file → run **Editor → Toggle Emphasis** → text becomes `_Apple_`. Undo restores it in one step.
- With `_Apple_` still selected, run Toggle Emphasis again → text becomes `Apple`.
- Place cursor inside an existing `_word_` (no selection, between the underscores) → run Toggle Emphasis → underscores are stripped, text becomes `word`.
- Place cursor in empty space, no selection → run Toggle Emphasis → `__` inserted with cursor between them.

If any case fails, fix `wrapSelection` before continuing.

- [ ] **Step 6: Commit**

```bash
git add extension.json Scripts/main.js
git commit -m "feat(#1): add Toggle Emphasis command"
```

---

## Task 4: Register Toggle Strong, Toggle Inline Code, Toggle Highlight

**Files:**
- Modify: `extension.json`
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add three entries to `commands.editor`**

After the Toggle Emphasis entry, add:

```json
            ,
            {
                "title": "Toggle Strong",
                "command": "io.dwk.djot.toggleStrong",
                "when": "editorSyntax == 'djot'"
            },
            {
                "title": "Toggle Inline Code",
                "command": "io.dwk.djot.toggleInlineCode",
                "when": "editorSyntax == 'djot'"
            },
            {
                "title": "Toggle Highlight",
                "command": "io.dwk.djot.toggleHighlight",
                "when": "editorSyntax == 'djot'"
            }
```

(Adjust the leading comma to match existing structure — i.e. a comma after the Toggle Emphasis closing `}` then the three new entries separated by commas, no trailing comma.)

- [ ] **Step 2: Add three matching entries to `commands.extensions`**

```json
            ,
            {
                "title": "Toggle Strong",
                "command": "io.dwk.djot.toggleStrong"
            },
            {
                "title": "Toggle Inline Code",
                "command": "io.dwk.djot.toggleInlineCode"
            },
            {
                "title": "Toggle Highlight",
                "command": "io.dwk.djot.toggleHighlight"
            }
```

- [ ] **Step 3: Validate JSON**

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

Expected: no output.

- [ ] **Step 4: Register the three commands in `Scripts/main.js`**

After the Toggle Emphasis registration, add:

```javascript
    nova.commands.register("io.dwk.djot.toggleStrong", function(editor) {
        wrapSelection(editor, "*", "*");
    });

    nova.commands.register("io.dwk.djot.toggleInlineCode", function(editor) {
        wrapSelection(editor, "`", "`");
    });

    nova.commands.register("io.dwk.djot.toggleHighlight", function(editor) {
        wrapSelection(editor, "{=", "=}");
    });
```

- [ ] **Step 5: Parse-check**

```bash
node --check Scripts/main.js
```

Expected: no output.

- [ ] **Step 6: Smoke-test in Nova**

Reload the extension. In `test.dj`:

- Select a word, run Toggle Strong → wrapped in `*…*`. Run again → unwrapped.
- Select a word, run Toggle Inline Code → wrapped in `` `…` ``. Run again → unwrapped.
- Select a word, run Toggle Highlight → wrapped in `{=…=}`. Run again → unwrapped.
- Verify the **Editor** menu lists all four toggle commands while a `.dj` file is active, and they disappear when switching to a non-Djot file.

- [ ] **Step 7: Commit**

```bash
git add extension.json Scripts/main.js
git commit -m "feat(#1): add Toggle Strong, Inline Code, Highlight commands"
```

---

## Task 5: Add `prependLines` helper and register Toggle Blockquote

**Files:**
- Modify: `extension.json`
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add the `prependLines` helper to `Scripts/main.js`**

Immediately after the `wrapSelection` function definition, add:

```javascript
// Toggle a line prefix on every line touched by any selected range.
// Detection rule: if every NON-EMPTY affected line already starts with prefix,
// strip the prefix from those lines (leave blank lines alone). Otherwise,
// prepend the prefix to every affected line that does not already have it
// (blank lines included — `> ` on a blank line is valid blockquote
// continuation).
function prependLines(editor, prefix) {
    var doc = editor.document;
    var docText = doc.getTextInRange(new Range(0, doc.length));
    var ranges = editor.selectedRanges;

    var lineStartSet = {};
    for (var i = 0; i < ranges.length; i++) {
        var r = ranges[i];
        var firstLineStart = lineStartFor(docText, r.start);
        var pos = firstLineStart;
        while (true) {
            lineStartSet[pos] = true;
            var nl = docText.indexOf("\n", pos);
            if (nl === -1 || nl >= r.end) break;
            pos = nl + 1;
            // If the selection ends exactly at a line start, do not include
            // that next line — match the convention of "lines the user is
            // visibly working on".
            if (pos > r.end) break;
        }
    }

    var lineStarts = Object.keys(lineStartSet)
        .map(Number)
        .sort(function(a, b) { return b - a; });

    if (lineStarts.length === 0) return;

    var allHavePrefix = true;
    var anyNonEmpty = false;
    for (var j = 0; j < lineStarts.length; j++) {
        var ls = lineStarts[j];
        var lineEnd = lineEndFor(docText, ls);
        var lineText = docText.substring(ls, lineEnd);
        if (lineText.length === 0) continue;
        anyNonEmpty = true;
        if (lineText.substring(0, prefix.length) !== prefix) {
            allHavePrefix = false;
        }
    }
    var stripping = anyNonEmpty && allHavePrefix;

    editor.edit(function(e) {
        for (var k = 0; k < lineStarts.length; k++) {
            var ls2 = lineStarts[k];
            var lineEnd2 = lineEndFor(docText, ls2);
            var lineText2 = docText.substring(ls2, lineEnd2);

            if (stripping) {
                if (lineText2.substring(0, prefix.length) === prefix) {
                    e.replace(new Range(ls2, ls2 + prefix.length), "");
                }
            } else {
                if (lineText2.substring(0, prefix.length) !== prefix) {
                    e.insert(ls2, prefix);
                }
            }
        }
    });
}

function lineStartFor(docText, pos) {
    var prevNl = docText.lastIndexOf("\n", pos - 1);
    return prevNl === -1 ? 0 : prevNl + 1;
}

function lineEndFor(docText, lineStart) {
    var nextNl = docText.indexOf("\n", lineStart);
    return nextNl === -1 ? docText.length : nextNl;
}
```

- [ ] **Step 2: Add Toggle Blockquote to `extension.json`**

Add to `commands.editor`:

```json
            ,
            {
                "title": "Toggle Blockquote",
                "command": "io.dwk.djot.toggleBlockquote",
                "when": "editorSyntax == 'djot'"
            }
```

Add to `commands.extensions`:

```json
            ,
            {
                "title": "Toggle Blockquote",
                "command": "io.dwk.djot.toggleBlockquote"
            }
```

- [ ] **Step 3: Validate JSON**

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

Expected: no output.

- [ ] **Step 4: Register the command**

In `Scripts/main.js`, after the Toggle Highlight registration:

```javascript
    nova.commands.register("io.dwk.djot.toggleBlockquote", function(editor) {
        prependLines(editor, "> ");
    });
```

- [ ] **Step 5: Parse-check**

```bash
node --check Scripts/main.js
```

- [ ] **Step 6: Smoke-test in Nova**

In `test.dj`:

- Select three consecutive paragraph lines, run **Editor → Toggle Blockquote** → all three lines get `> ` prepended.
- With those same three lines (now `> `-prefixed) selected, run again → prefix is stripped from all three.
- Select a mix of `> ` lines and unprefixed lines, run → all lines become prefixed (toggle to "add" mode).
- Select a region that includes a blank line between two paragraphs, run → all lines including the blank one get `> `.
- Verify undo restores in one step.

- [ ] **Step 7: Commit**

```bash
git add extension.json Scripts/main.js
git commit -m "feat(#1): add prependLines helper and Toggle Blockquote command"
```

---

## Task 6: Register Insert Link

**Files:**
- Modify: `extension.json`
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add Insert Link to `extension.json`**

Add to `commands.editor`:

```json
            ,
            {
                "title": "Insert Link",
                "command": "io.dwk.djot.insertLink",
                "when": "editorSyntax == 'djot'"
            }
```

Add to `commands.extensions`:

```json
            ,
            {
                "title": "Insert Link",
                "command": "io.dwk.djot.insertLink"
            }
```

- [ ] **Step 2: Validate JSON**

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

- [ ] **Step 3: Register the command in `Scripts/main.js`**

After the Toggle Blockquote registration:

```javascript
    nova.commands.register("io.dwk.djot.insertLink", function(editor) {
        var firstRange = editor.selectedRanges[0];
        var selectedText = editor.document.getTextInRange(firstRange);
        var insertText = "[" + selectedText + "](url)";

        editor.edit(function(e) {
            e.replace(firstRange, insertText);
        });

        var urlStart = firstRange.start + 1 + selectedText.length + 2;
        var urlEnd = urlStart + 3;
        editor.selectedRanges = [new Range(urlStart, urlEnd)];
    });
```

- [ ] **Step 4: Parse-check**

```bash
node --check Scripts/main.js
```

- [ ] **Step 5: Smoke-test in Nova**

In `test.dj`:

- Place cursor in empty space, run Insert Link → `[](url)` inserted with `url` selected (so the user can immediately type).
- Select the word `Apple`, run Insert Link → `[Apple](url)` with `url` selected.
- Verify undo restores in one step.

- [ ] **Step 6: Commit**

```bash
git add extension.json Scripts/main.js
git commit -m "feat(#1): add Insert Link command"
```

---

## Task 7: Register Insert Fenced Code Block

**Files:**
- Modify: `extension.json`
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add Insert Fenced Code Block to `extension.json`**

Add to `commands.editor`:

```json
            ,
            {
                "title": "Insert Fenced Code Block",
                "command": "io.dwk.djot.insertCodeBlock",
                "when": "editorSyntax == 'djot'"
            }
```

Add to `commands.extensions`:

```json
            ,
            {
                "title": "Insert Fenced Code Block",
                "command": "io.dwk.djot.insertCodeBlock"
            }
```

- [ ] **Step 2: Validate JSON**

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

- [ ] **Step 3: Register the command in `Scripts/main.js`**

```javascript
    nova.commands.register("io.dwk.djot.insertCodeBlock", function(editor) {
        var firstRange = editor.selectedRanges[0];
        var selectedText = editor.document.getTextInRange(firstRange);
        var insertText = "```language\n" + selectedText + "\n```\n";

        editor.edit(function(e) {
            e.replace(firstRange, insertText);
        });

        var langStart = firstRange.start + 3;
        var langEnd = langStart + 8;
        editor.selectedRanges = [new Range(langStart, langEnd)];
    });
```

- [ ] **Step 4: Parse-check**

```bash
node --check Scripts/main.js
```

- [ ] **Step 5: Smoke-test in Nova**

In `test.dj`:

- Place cursor on an empty line, run Insert Fenced Code Block → fence inserted with `language` selected.
- Select a chunk of plain text, run Insert Fenced Code Block → text moves into the fence body, `language` selected.
- Open the live preview pane and confirm the fence renders as a code block with the language tag (or as plain code if `language` isn't a known one).

- [ ] **Step 6: Commit**

```bash
git add extension.json Scripts/main.js
git commit -m "feat(#1): add Insert Fenced Code Block command"
```

---

## Task 8: Register Insert Table

**Files:**
- Modify: `extension.json`
- Modify: `Scripts/main.js`

- [ ] **Step 1: Add Insert Table to `extension.json`**

Add to `commands.editor`:

```json
            ,
            {
                "title": "Insert Table",
                "command": "io.dwk.djot.insertTable",
                "when": "editorSyntax == 'djot'"
            }
```

Add to `commands.extensions`:

```json
            ,
            {
                "title": "Insert Table",
                "command": "io.dwk.djot.insertTable"
            }
```

- [ ] **Step 2: Validate JSON**

```bash
python3 -c "import json; json.load(open('extension.json'))"
```

- [ ] **Step 3: Register the command in `Scripts/main.js`**

```javascript
    nova.commands.register("io.dwk.djot.insertTable", function(editor) {
        var firstRange = editor.selectedRanges[0];
        var selectedText = editor.document.getTextInRange(firstRange);
        var firstCell = selectedText || "Header 1";
        var insertText =
            "| " + firstCell + " | Header 2 |\n" +
            "|---|---|\n" +
            "| Cell 1 | Cell 2 |\n" +
            "| Cell 3 | Cell 4 |\n";

        editor.edit(function(e) {
            e.replace(firstRange, insertText);
        });

        var cellStart = firstRange.start + 2;
        var cellEnd = cellStart + firstCell.length;
        editor.selectedRanges = [new Range(cellStart, cellEnd)];
    });
```

- [ ] **Step 4: Parse-check**

```bash
node --check Scripts/main.js
```

- [ ] **Step 5: Smoke-test in Nova**

In `test.dj`:

- Place cursor on a blank line, run Insert Table → 2×3 table inserted, `Header 1` selected.
- Select the word `Apple`, run Insert Table → `Apple` becomes the first header cell text, selection lands on `Apple`.
- Confirm in preview that the table renders correctly with header row and two body rows.

- [ ] **Step 6: Commit**

```bash
git add extension.json Scripts/main.js
git commit -m "feat(#1): add Insert Table command"
```

---

## Task 9: Documentation updates

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/TODO.md`

- [ ] **Step 1: Add an "Editor Commands" section to `README.md`**

Insert this section after the "Live Preview" section and before "Building from Source":

```markdown
## Editor Commands

Eight editor commands are available under the **Editor** menu (and via Nova's command palette) when a `.dj` file is open. They also appear under **Extensions → Djot**. None claim default keyboard shortcuts — bind your own via Extensions → Extension Library → Djot → Settings → Keyboard Shortcuts.

| Command | Effect |
|---|---|
| Toggle Emphasis | Wraps selection with `_…_`, or unwraps if already wrapped |
| Toggle Strong | Wraps selection with `*…*` |
| Toggle Inline Code | Wraps selection with `` `…` `` |
| Toggle Highlight | Wraps selection with `{=…=}` |
| Toggle Blockquote | Prepends `> ` to selected lines, or strips it if all lines are quoted |
| Insert Link | Inserts `[text](url)`, leaves the cursor on `url` |
| Insert Fenced Code Block | Inserts a triple-backtick fence with a `language` placeholder |
| Insert Table | Inserts a 2-column × 3-row starter table |

All commands batch their edits into a single undo step. Toggle commands also detect when the cursor is *inside* an already-wrapped span (no selection) and unwrap it.
```

- [ ] **Step 2: Add a `1.1.0` entry to `CHANGELOG.md`**

Insert at the top of the changelog (just under any heading, before existing entries):

```markdown
## 1.1.0

- Added eight editor commands: Toggle Emphasis, Toggle Strong, Toggle Inline Code, Toggle Highlight, Toggle Blockquote, Insert Link, Insert Fenced Code Block, Insert Table.
```

- [ ] **Step 3: Update `docs/TODO.md`**

Change:

```markdown
- [ ] JavaScript scripts or commands
```

to:

```markdown
- [x] JavaScript scripts or commands (initial set; deferred items tracked in #6, #7, #8, #9)
```

- [ ] **Step 4: Commit**

```bash
git add README.md CHANGELOG.md docs/TODO.md
git commit -m "docs(#1): document editor commands and update changelog"
```

---

## Task 10: Full smoke-test pass and PR

**Files:** none modified — verification only

- [ ] **Step 1: Reload extension and run the spec's full smoke-test checklist**

In Nova with `test.dj` open, walk through every item from the spec's Testing section:

- Toggle Emphasis on selected word → `_word_`
- Toggle Emphasis on `_word_` selected → `word`
- Toggle Emphasis with empty selection inside existing `_word_` → `word`
- Toggle Strong / Toggle Inline Code / Toggle Highlight — same three cases each
- Toggle Blockquote on three selected lines → all three get `> `
- Toggle Blockquote on three already-quoted lines → all three lose `> `
- Toggle Blockquote on a mix of quoted and unquoted lines → all become quoted
- Insert Link with no selection → `[](url)` with `url` selected
- Insert Link with selected text → `[selected](url)` with `url` selected
- Insert Fenced Code Block, Insert Table → render correctly in preview
- Multi-cursor Toggle Emphasis → all selections wrap independently
- Undo after any command → exactly one undo step restores original text

If any case fails, do not proceed — fix the bug, commit the fix, and rerun the checklist.

- [ ] **Step 2: Push the branch**

```bash
git push -u origin feat/1-editor-commands
```

- [ ] **Step 3: Open a PR**

```bash
gh pr create --title "feat(#1): add eight Djot editor commands" --body "$(cat <<'EOF'
## Summary

- Adds eight editor commands implementing the design in `docs/superpowers/specs/2026-04-29-djot-editor-commands-design.md`.
- Toggle Emphasis / Strong / Inline Code / Highlight wrap and unwrap inline delimiters.
- Toggle Blockquote prepends or strips `> ` on selected lines.
- Insert Link / Fenced Code Block / Table insert templates with the cursor placed on the most-likely-edited slot.
- Bumps version to 1.1.0.

Closes the first checkbox of #1. Deferred items tracked in #6 (templates), #7 (heading levels), #8 (footnote auto-definition), and #9 (fold all sections — blocked on Nova API).

## Test plan

- [ ] Reload extension, walk through the smoke-test checklist in the spec.
- [ ] Confirm commands are gated to `.dj` files (do not appear under Editor menu when a non-Djot file is focused).
- [ ] Confirm undo restores in a single step for each command.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 4: Return the PR URL** so the user can review.

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

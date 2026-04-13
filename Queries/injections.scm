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

# ADR 0002: Licensing

Status: accepted, 2026-10-03.

## Decision
- Code (Go sources, scripts, build files): MIT (`LICENSE`).
- Own documentation and printables (Markdown, templates, generated HTML and PDF): CC BY 4.0 (`LICENSE-DOCS`, unmodified legal code
  from creativecommons.org, SHA-256 `9ba9550ad48438d0836ddab3da480b3b69ffa0aac7b7878b5a0039e7ab429411`).
- Third-party data keeps its license: every word list is stored with its upstream reference, license text, attribution and
  SHA-256 next to it.
- Third-party material whose redistribution is not clearly permitted (for example vendor PDFs) is linked, never copied.

## Reasons
MIT is short, permissive and widely understood. CC BY 4.0 lets anyone print, translate and redistribute the
printables while keeping attribution, which matters for material that users will pass on as paper.

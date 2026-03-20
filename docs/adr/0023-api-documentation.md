# ADR 0023: API Documentation with GoDoc/TSDoc and GitHub Pages

## Status

Accepted

## Date

2026-03-20

## Context

Source code documentation and browsable API references were disconnected. Developers and AI assistants had no centralized, auto-generated documentation for Go packages or TypeScript modules. Manual documentation maintenance does not scale with 16+ game implementations across backend and frontend.

We needed:
1. Inline documentation comments on all exported symbols (GoDoc for Go, TSDoc for TypeScript)
2. Automated generation of browsable HTML documentation
3. Deployment to a publicly accessible site on every release

## Decision

We adopt the following documentation stack:

- **GoDoc comments** (`// SymbolName description`) on all exported Go symbols under `internal/`
- **TSDoc comments** (`/** description */`) on all exported TypeScript symbols under `frontend/src/`
- **gomarkdoc** to generate Markdown from Go packages, converted to HTML with **pandoc**
- **TypeDoc** to generate HTML documentation from TypeScript sources
- **GitHub Pages** deployment via GitHub Actions on push to `master`, unified with the existing repomix deployment

The generated site structure:
```
_site/
  index.html          # Landing page with links
  repomix-output.txt  # Compressed repo snapshot
  go/                  # Go API docs (HTML per package)
  ts/                  # TypeScript API docs (HTML)
```

## Consequences

### Positive
- All exported symbols have inline documentation, improving IDE hover information and code readability
- Auto-generated docs stay in sync with code — no manual HTML maintenance
- Single workflow replaces the previous repomix-only deployment
- Developers and AI assistants can browse API docs at the GitHub Pages URL

### Negative
- Adding GoDoc/TSDoc to ~400 files is a large initial investment
- CI build time increases slightly (gomarkdoc + pandoc + TypeDoc generation)
- TypeDoc is added as a dev dependency, increasing `node_modules` size marginally

### Neutral
- Documentation quality depends on comment quality — automated generation does not validate accuracy
- The `deploy-repomix.yml` workflow is removed in favor of the unified `deploy-pages.yml`

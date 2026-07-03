# Papiro Project Development Plan

## 1. Executive Summary
Papiro is a multi-component ecosystem designed to index, convert, and serve markdown-based documentation as interactive wiki websites. The project consists of three primary modules:
- **Papiro Indexer**: A Go CLI tool that recursively scans directories, converts markdown/HTML files, and generates a structured `index.json` alongside static assets.
- **Papiro Web UI**: A frontend interface to render the indexed HTML files as a navigable wiki.
- **Papiro Deploy**: A GitHub Action that automates the indexing and deployment process for any repository, publishing it to GitHub Pages.

This document outlines the current architecture, existing features, identified gaps, and a phased development roadmap to ensure stability, scalability, and production readiness.

---

## 2. System Architecture

### 2.1 Papiro Indexer (Go CLI)
- **Entry Point**: `main.go` parses CLI flags (`-path`, `-destination`, `-obsidian`) and initializes the configuration.
- **Configuration Model**: `model/configuration.go` holds runtime settings.
- **Indexing Engine**: `service/index.go` handles recursive directory traversal via `IndexPath`. It categorizes files by extension (`.md`, `.html`, or others) and delegates processing:
  - `.md`: Parsed using `blackfriday` (`service/markdown.go`) and converted to HTML.
  - `.html`: Processed directly.
  - Others: Copied as static assets.
- **Path & URL Handling**: `service/paths.go` manages URL encoding (lowercase, hyphenated spaces), extension manipulation, and file I/O operations (`OpenFile`, `CopyFile`, `WriteFile`).
- **Output Generation**: `WriteIndex` serializes the directory tree into a JSON structure (`model/index.go`) containing `IndexItem` nodes (Type: Directory/File, Path, Files map) and writes it to `index.json`.
- **Logging**: `logger/logger.go` provides simple console logging with emoji prefixes for `Log`, `Error`, and `Warning`.

### 2.2 Papiro Web UI
- Designed to consume the generated `index.json` and static HTML files.
- Provides wiki-style navigation, search, and cross-linking capabilities.
- Currently conceptual; requires frontend implementation aligned with the indexer's output structure.

### 2.3 Papiro Deploy
- GitHub Action workflow that triggers on push/PR events.
- Runs the Papiro Indexer against the repository.
- Publishes the generated static site to GitHub Pages.
- Enables zero-maintenance documentation hosting for any markdown-heavy repo.

---

## 3. Current Feature Set
| Component | Features |
|-----------|----------|
| **Indexer** | Recursive directory scanning, Markdown-to-HTML conversion, HTML passthrough, Static asset copying, URL-safe path generation, JSON tree output (`index.json`), CLI flags for path/destination/obsidian mode |
| **Paths/IO** | File stats validation, safe file creation/copying, extension manipulation, URL encoding/decoding helpers |
| **Logging** | Console output with severity markers (❌ Error, ⚠️ Warning) |
| **Config** | Struct-based configuration passing through the pipeline |

---

## 4. Identified Gaps & Missing Features
Based on `docs/Features.md` and architectural review:

### 🔴 Critical / High Priority
1. **Numbered Indexing**: Support for `<number>_filename` sorting to ensure deterministic file ordering in the output directory.
2. **Obsidian Wikilinks Support**: Parse and resolve `[[WikiLink]]` syntax, convert them to relative HTML anchors or external links, and handle missing target warnings gracefully.
3. **Error Handling & Validation**: Current error handling relies on basic `if nil != error` checks. Needs structured errors, context wrapping, and graceful degradation for permission issues or large directories.

### 🟡 Medium Priority
4. **Concurrency**: Indexing large repositories is sequential. Implement worker pools for file processing to improve performance.
5. **Configuration File Support**: Allow `.papiro.yaml` or `.json` config instead of only CLI flags.
6. **Asset Pipeline Enhancements**: Handle CSS/JS bundling, image optimization, and relative path correction in generated HTML.

### 🟢 Low Priority / Nice-to-Have
7. **Incremental Indexing**: Skip unchanged files using checksums or modification times.
8. **Plugin System**: Allow custom processors for other file types (`.adoc`, `.rst`, `.pdf`).
9. **Web UI Search & Filtering**: Client-side search, tag filtering, and dark mode.

---

## 5. Development Roadmap

### Phase 1: Core Indexer Hardening & Obsidian Compatibility
- [ ] Implement numbered file sorting (`1_name.md`, `2_name.md`)
- [ ] Add Obsidian wikilink parser & resolver
- [ ] Replace basic error checks with structured errors (`fmt.Errorf` + `%w`)
- [ ] Add context-aware logging (trace/debug levels)
- [ ] Write unit tests for `service/paths.go` and `service/markdown.go`

### Phase 2: Performance & Scalability
- [ ] Introduce concurrent file processing using `errgroup` or worker pools
- [ ] Implement incremental indexing via `.papiro/cache`
- [ ] Add progress bar / ETA output during long runs
- [ ] Optimize JSON serialization for large trees (streaming or chunked writes)

### Phase 3: Web UI & Frontend Integration
- [ ] Scaffold minimal Go-based static server for local testing
- [ ] Build lightweight SPA/Vue/React wiki viewer consuming `index.json`
- [ ] Implement client-side search, breadcrumb navigation, and backlinks
- [ ] Add theme support (light/dark) and responsive design

### Phase 4: GitHub Action & CI/CD Pipeline
- [ ] Create `.github/workflows/papiro-deploy.yml`
- [ ] Package indexer as a reusable Go binary or Docker image
- [ ] Add workflow inputs for custom paths, destinations, and obsidian mode
- [ ] Implement deployment validation (check `index.json` exists, verify HTML output)

### Phase 5: Documentation & Release
- [ ] Write comprehensive CLI docs (`papiro --help`, examples, troubleshooting)
- [ ] Create README with architecture diagrams and quickstart guide
- [ ] Set up Go module versioning, tagging, and release automation
- [ ] Add integration tests against sample markdown repositories

---

## 6. Technical Considerations & Best Practices
- **Dependency Management**: Pin `blackfriday` or migrate to a maintained fork (e.g., `goldmark`) for security updates.
- **Path Traversal Safety**: Validate all input paths to prevent directory traversal attacks (`../`).
- **Memory Usage**: Stream large files instead of loading entirely into memory; use buffered I/O.
- **Cross-Platform Compatibility**: Ensure path separators and permissions work on Windows, macOS, and Linux.
- **Observability**: Add metrics (indexing time, file count, error rate) for production monitoring.

---

## 7. Next Steps
1. Review and approve this development plan
2. Prioritize Phase 1 tasks in the issue tracker
3. Set up branch protection and CI checks for Go code
4. Begin implementation of numbered indexing & wikilink support

*This plan will be updated iteratively as new requirements or architectural decisions emerge.*

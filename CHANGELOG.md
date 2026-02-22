# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-02-22

### Added
- Initial release
- Web scraping engine with Colly (depth control, same-domain)
- Link transformation for offline portability
- HTML/JS filtering system (pattern-based)
- REST API (scrape, status, export)
- Async scraping with real-time status tracking
- ZIP export (full project archive)
- PDF export (consolidated single document)
- Web UI (responsive, minimalist)
- Docker containerization (multi-stage build)
- Comprehensive error handling
- Structured logging
- Full documentation (README, AGENTS, per-agent guides)

### Security
- Non-root user in Docker
- Input validation (URL, depth, filters)
- Graceful error handling (no sensitive data leakage)

### Performance
- Concurrent asset downloading
- Streaming export (ZIP, PDF)
- Efficient memory usage (file-based storage)

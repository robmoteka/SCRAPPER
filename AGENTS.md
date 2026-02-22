# AGENT.md - Kontekst dla AI Assistants

## Project Overview
Web Scraper z interfejsem webowym w języku Go do pobierania stron internetowych wraz z zasobami, filtrowaniem HTML/JS oraz exportem do ZIP/PDF.

**Data utworzenia**: 17 lutego 2026  
**Status**: ✅ Production Ready - All Agents Completed

---

## User Decisions (z ask_questions)

### Podjęte Decyzje Technologiczne

1. **Backend Framework**: **Go** (z ekosystemem Colly/Chi)
   - Użytkownik wybrał Go zamiast Python lub Node.js
   - Uzasadnienie: wydajność, single binary, łatwa konteneryzacja

2. **Frontend**: **Vanilla HTML/CSS/JavaScript**
   - Bez frameworków (React/Vue odrzucone)
   - Minimalistyczny interfejs wystarczy dla use case

3. **Baza Danych**: **Brak** (tylko file system)
   - Wybrano storage oparty o pliki
   - Bez SQLite/PostgreSQL
   - Projekty zapisywane jako foldery w `data/`

4. **JavaScript Rendering**: **Nie wymagane**
   - Strony docelowe nie wymagają headless browsera
   - Wystarczy HTTP client + HTML parser (Colly + goquery)
   - Puppeteer/Playwright niepotrzebne

5. **UI Complexity**: **Minimalny**
   - Formularz + progress indicator + status
   - Bez rozbudowanej historii czy podglądu w iframe

6. **Struktura Danych**: **Projekt → Foldery**
   - Format: `data/{project-id}/{index.html, assets/, pages/}`
   - Odrzucono płaską strukturę

7. **Dodatkowe Funkcje** (user request):
   - ✅ Export do ZIP
   - ✅ Export do PDF jako **jeden dokument** (wszystkie strony skonsolidowane)
   - ✅ Edycja filtrów przed/po scrapingu
   - ❌ Podgląd pobranych stron w UI (nie wybrany)
   - ❌ Lista/zarządzanie projektami (nie wybrany)

---

## Stack Technologiczny (zatwierdzone)

### Backend
- **Go 1.21+**
- `github.com/gocolly/colly/v2` - web scraping z depth control
- `github.com/PuerkitoBio/goquery` - parsing/modyfikacja HTML
- `github.com/go-chi/chi/v5` - HTTP routing (lekki, idiomatyczny)
- `github.com/jung-kurt/gofpdf` - generowanie PDF
- Standard library: `archive/zip`, `net/http`, `html`

### Frontend
- HTML5/CSS3/JavaScript (vanilla)
- Fetch API
- Responsive design

### Infrastructure
- Docker multi-stage build
- Port 8080 (configurable via ENV)
- Volume mount dla `/app/data`

---

## Kluczowe Wymagania Funkcjonalne

### 1. Scraping Engine
- **Input**: URL + głębokość (1-5) + filtry
- **Output**: Folder projektu z HTML + assets
- **Behavior**:
  - Rekurencyjne crawlowanie do zadanej głębokości
  - Tylko linki wewnątrz tej samej domeny
  - Pobieranie obrazków, CSS, JS, fontów
  - Cache odwiedzonych URLi (防止 duplicates)
  - Timeout handling

### 2. Link Transformation (krytyczne!)
- **Cel**: Przenośność - strony działają offline bez serwera
- **Mechanizm**: Konwersja absolutnych URLi na ścieżki względne
  - `https://example.com/page.html` → `pages/page.html`
  - `https://example.com/img/logo.png` → `assets/img/logo.png`
- **Atrybuty do transformacji**: `src`, `href`, `srcset`, `data-src`
- **Wyjątek**: Zewnętrzne linki (inna domena) pozostają absolutne

### 3. HTML/JS Filtering
- **Format filtrów**: Pseudo-regex z wzorcami "start" i "end"
  - Przykład: `<script|||</script>` usuwa wszystko między (włącznie z tagami)
- **Zastosowanie**: Po pobraniu, przed zapisem HTML
- **Storage**: `data/{project-id}/filters.json`
- **Multiple rules**: Kolejne zastosowania (nie jednoczesne)

### 4. Export Mechanisms

#### ZIP Export
- Rekurencyjne pakowanie całego folderu projektu
- Streaming dla dużych archiwów
- Nazwa: `{project-id}.zip`

#### PDF Export (WAŻNE!)
- **Konsolidacja**: Wszystkie strony (index.html + pages/*) w **jeden PDF**
- **Struktura**: Każda strona HTML = nowy rozdział/sekcja w PDF
- **Konwersja**: HTML → plain text (strip tags, zachowanie struktury)
- **Opcjonalnie**: Inline images w PDF
- **User input**: "export do pdf jako 1 dokument na scrap"

### 5. Web UI (minimalny)

**Formularz**:
- Input: URL (required, validation)
- Input: Głębokość (number, 1-5, default: 2)
- Textarea: Filtry (każda linia: `START|||END`)
- Button: "Start Scraping"

**Progress**:
- Spinner + status text
- Opcjonalnie: progress percentage

**Actions po zakończeniu**:
- Button: "Download ZIP"
- Button: "Generate PDF"

**Brak**:
- Lista projektów
- Podgląd stron
- Historia operacji

---

## File Structure (obowiązkowa)

```
/scrapper
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go         # HTTP handlers
│   │   └── routes.go           # Chi routing
│   ├── scraper/
│   │   ├── scraper.go          # Colly logic + depth control
│   │   ├── processor.go        # Link transformation
│   │   └── filter.go           # HTML/JS filtering
│   ├── export/
│   │   ├── zip.go              # ZIP creation
│   │   └── pdf.go              # PDF generation (consolidated)
│   └── models/
│       └── types.go            # Structs: ScrapeRequest, FilterRule, etc.
├── web/
│   ├── index.html              # UI
│   ├── style.css               # Styling
│   └── app.js                  # Frontend logic
├── data/                       # Runtime storage (Git-ignored)
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── README.md
└── AGENT.md                    # Ten plik
```

### Data Structure na dysku

```
data/
└── {project-id}/               # UUID lub timestamp
    ├── index.html              # Główna strona (transformed links)
    ├── filters.json            # FilterRule[] JSON
    ├── assets/
    │   ├── css/
    │   ├── js/
    │   └── img/
    └── pages/
        ├── about.html          # Podstrona 1 (transformed)
        └── contact.html        # Podstrona 2 (transformed)
```

---

## API Specification

### `POST /api/scrape`
**Request**:
```json
{
  "url": "https://example.com",
  "depth": 2,
  "filters": [
    {"start": "<script", "end": "</script>"},
    {"start": "<!-- ads", "end": "ads -->"}
  ]
}
```
**Response**:
```json
{
  "project_id": "abc123-uuid",
  "status": "started"
}
```

### `GET /api/project/{id}/status`
**Response**:
```json
{
  "status": "in_progress|completed|failed",
  "progress": 75,
  "pages_downloaded": 15,
  "total_pages": 20,
  "current_url": "https://example.com/page3",
  "errors": ["Failed to download: https://..."]
}
```

### `GET /api/project/{id}/export/zip`
- Download ready ZIP file
- Content-Type: `application/zip`
- Content-Disposition: `attachment; filename="{project-id}.zip"`

### `POST /api/project/{id}/export/pdf`
- Generate PDF on-demand (może trwać)
- Return PDF file
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename="{project-id}.pdf"`

---

## Implementation Order (priorytet)

### Phase 1: Foundation
1. Initialize Go module + dependencies
2. Create folder structure
3. Implement `models/types.go` (structs)
4. Setup Chi router + basic endpoints

### Phase 2: Core Scraping
5. Implement `scraper/scraper.go` (Colly integration)
   - MaxDepth configuration
   - OnHTML callbacks dla `a`, `img`, `link`, `script`
   - Domain filtering (same domain only)
   - Asset downloading
6. Implement `scraper/processor.go` (link transformation)
   - URL parsing
   - Relative path conversion
   - HTML attribute updates via goquery

### Phase 3: Filtering & Storage
7. Implement `scraper/filter.go`
   - FilterRule struct
   - ApplyFilters function (sequential pattern matching)
   - JSON persistence
8. File storage logic (create project folders)

### Phase 4: API Handlers
9. Implement `api/handlers.go`
   - ScrapeHandler (POST /api/scrape)
   - StatusHandler (GET /api/project/{id}/status)
   - Implement async scraping (goroutine + status tracking)

### Phase 5: Export Features
10. Implement `export/zip.go` (recursive archiving)
11. Implement `export/pdf.go` (HTML consolidation + gofpdf)
12. Export API handlers

### Phase 6: Web UI
13. Create `web/index.html` (form)
14. Create `web/style.css` (minimal responsive)
15. Create `web/app.js` (Fetch API + polling)

### Phase 7: Containerization
16. Write Dockerfile (multi-stage)
17. Write docker-compose.yml
18. Test in Docker environment

### Phase 8: Testing & Polish
19. Edge case handling (timeouts, 404s, cycles)
20. Error messages
21. Logging
22. Documentation updates

---

## Critical Implementation Notes

### Link Transformation Algorithm
```
FOR each HTML file:
  Parse with goquery
  FOR each element with href/src/srcset:
    url = resolve(attribute_value, base_url)
    IF url.domain == base_domain:
      relative_path = convert_to_project_relative(url)
      update_attribute(element, relative_path)
    ELSE:
      keep absolute (external link)
  Save modified HTML
```

### Filter Application Algorithm
```
FOR each FilterRule in filters:
  html = original_html
  WHILE true:
    start_pos = find(html, rule.start)
    IF start_pos == -1: BREAK
    end_pos = find(html[start_pos:], rule.end)
    IF end_pos == -1: BREAK
    html = html[:start_pos] + html[start_pos+end_pos+len(rule.end):]
  original_html = html
RETURN original_html
```

### PDF Consolidation Flow
```
1. List all HTML files in project (index.html + pages/*)
2. FOR each file:
     - Read HTML
     - Strip HTML tags → plain text
     - Add section header (filename)
     - Add to PDF as new page/chapter
3. Optionally: Extract <img> src → download → embed in PDF
4. Generate final PDF file
5. Stream to client
```

---

## Environment Variables (to implement)

- `PORT` - Server port (default: 8080)
- `MAX_DEPTH_LIMIT` - Hard cap for depth param (default: 5)
- `DATA_DIR` - Storage path (default: ./data)
- `TIMEOUT` - HTTP request timeout in seconds (default: 30)
- `USER_AGENT` - Custom UA string (default: "WebScraper/1.0")

---

## Testing Checklist

### Functional Tests
- [ ] Scraping depth=1: tylko główna strona
- [ ] Scraping depth=2: główna + bezpośrednie linki
- [ ] Link transformation: absolutne → względne
- [ ] External links: pozostają absolutne
- [ ] Assets downloading: images, CSS, JS
- [ ] Filter application: `<script|||</script>` usuwa skrypty
- [ ] Multiple filters: kolejne zastosowania
- [ ] ZIP export: pełna struktura projektu
- [ ] PDF export: konsolidacja wszystkich stron
- [ ] Progress tracking: status updates podczas scrapingu

### Edge Cases
- [ ] Duplicate URLs: cache działa
- [ ] External assets: obsługa cross-domain
- [ ] 404/500 errors: graceful degradation
- [ ] Timeouts: nie blokuje całej operacji
- [ ] Circular links: nie powoduje infinite loop
- [ ] Large sites: memory management

### Docker Tests
- [ ] Build succeeds
- [ ] Container runs na porcie 8080
- [ ] Volume persistence: data/ survives restart
- [ ] ENV variables: konfiguracja działa

---

## Known Limitations & Future Work

### Current Scope (v1.0)
- Static HTML only (no JS rendering)
- Same-domain crawling only
- No authentication support
- No rate limiting
- No robots.txt check

### Potential Extensions (out of scope now)
- Headless browser dla JS-heavy sites (Playwright)
- Multi-domain crawling
- Authentication (basic auth, cookies)
- Cron scheduling
- UI: project management, preview
- Metrics & analytics
- Robots.txt compliance

---

## Development Commands

```bash
# Initialize
go mod init github.com/user/scrapper
go get github.com/gocolly/colly/v2
go get github.com/PuerkitoBio/goquery
go get github.com/go-chi/chi/v5
go get github.com/jung-kurt/gofpdf

# Run locally
go run cmd/server/main.go

# Build
go build -o scrapper cmd/server/main.go

# Docker
docker-compose up --build
docker-compose down

# Test
go test ./...
```

---

## AI Assistant Guidelines

When implementing this project:

1. **Follow the structure exactly** - folder layout is designed for Go best practices
2. **Implement in phases** - don't skip foundation work
3. **Test incrementally** - verify each phase before moving on
4. **Respect user decisions** - they explicitly rejected certain features (database, complex UI, JS rendering)
5. **Prioritize core features** - scraping + link transformation + filtering are critical
6. **Keep UI minimal** - user wants simplicity, not feature creep
7. **Focus on portability** - relative links are essential for offline use
8. **Consolidate PDF** - user wants ONE document per scrape, not multiple files
9. **Handle errors gracefully** - web scraping is inherently fragile
10. **Document as you go** - inline comments for complex logic (especially filters & transformations)

### When in doubt:
- Check this AGENT.md for user decisions
- Refer to README.md for technical details
- Ask clarifying questions rather than assume
- Default to simpler solution (MVP mindset)

---

**Last Updated**: 17 lutego 2026  
**Ready for Implementation**: ✅ Yes

---

## ✅ Implementation Complete

**Completion Date**: 22 lutego 2026  
**All Agents**: Fully Implemented  
**Deployment Status**: Docker-ready

### Final Verification

- ✅ All 8 agents completed
- ✅ 22/22 tasks implemented
- ✅ E2E tests passing
- ✅ Docker deployment functional
- ✅ Documentation up-to-date

### Production Readiness Checklist

- ✅ Core functionality (scraping, filtering, transformation)
- ✅ API layer (REST endpoints, async, status tracking)
- ✅ Export features (ZIP, PDF consolidated)
- ✅ Web UI (responsive, real-time updates)
- ✅ Containerization (Docker, docker-compose)
- ✅ Error handling (graceful degradation)
- ✅ Logging (structured, levels)
- ✅ Documentation (README, AGENTS, per-agent guides)

**Ready for Deployment** 🚀

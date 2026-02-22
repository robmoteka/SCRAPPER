# Web Scraper z Interfejsem Webowym

Aplikacja do pobierania stron internetowych wraz z obrazkami i zasobami, z możliwością filtrowania HTML/JS oraz exportu do ZIP/PDF.

## Stack Technologiczny

### Backend (Go 1.21+)
- `github.com/gocolly/colly/v2` - web scraping engine z kontrolą głębokości
- `github.com/PuerkitoBio/goquery` - parsing i modyfikacja HTML
- `github.com/go-chi/chi/v5` - routing HTTP (lekki, idiomatyczny)
- `github.com/jung-kurt/gofpdf` - generowanie PDF
- Standard library: `archive/zip`, `net/http`, `html`

### Frontend
- Vanilla HTML5/CSS3/JavaScript
- Fetch API do komunikacji z backendem
- Responsywny formularz + progress indicator

### Infrastruktura
- Docker multi-stage build (build + runtime)
- File system storage (bez bazy danych)
- Port 8080 (konfigurowalny przez ENV)

## Struktura Projektu

```
/scrapper
├── cmd/
│   └── server/
│       └── main.go                    # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go                # HTTP handlers
│   │   └── routes.go                  # Routing setup
│   ├── scraper/
│   │   ├── scraper.go                 # Colly scraping logic
│   │   ├── processor.go               # Link transformation
│   │   └── filter.go                  # HTML/JS filtering
│   ├── export/
│   │   ├── zip.go                     # ZIP export
│   │   └── pdf.go                     # PDF generation
│   └── models/
│       └── types.go                   # Data structures
├── web/
│   ├── index.html                     # UI formularz
│   ├── style.css                      # Styling
│   └── app.js                         # Frontend logic
├── data/                              # Runtime projects storage
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### Struktura Danych

Zapisane projekty organizowane są w folderach:

```
data/
└── {project-id}/
    ├── index.html                     # Główna strona
    ├── filters.json                   # Reguły filtrowania
    ├── assets/                        # Obrazki, CSS, JS
    │   ├── css/
    │   ├── js/
    │   └── img/
    └── pages/                         # Podstrony
        ├── page1.html
        └── page2.html
```

## Funkcjonalności

### Scraping
- **Parametr głębokości**: kontrola rekurencyjnego pobierania stron (1-5 poziomów)
- **Zakres URL przez prefix**: opcjonalne ograniczenie crawl i pobierania assetów do URL-i zaczynających się od `url_prefix`
- **Pobieranie zasobów**: obrazki, CSS, JS, fonty
- **Transformacja linków**: konwersja na ścieżki względne dla przenośności
- **Crawling wewnątrz domeny**: podążanie za linkami tylko w obrębie tej samej domeny

### Filtrowanie HTML/JS
- **Wzorce tekstowe**: usuwanie fragmentów kodu między "start pattern" a "end pattern"
- **Multiple rules**: możliwość zastosowania wielu filtrów naraz
- **Przykłady**:
  - `<script|||</script>` - usuwa wszystkie skrypty
  - `<div id="ads"|||</div>` - usuwa div z reklamami
  - `<!-- comment-start|||comment-end -->` - usuwa komentarze

### Export
- **ZIP**: pakowanie całego projektu (HTML + assets) do archiwum
- **PDF**: konsolidacja wszystkich stron w jeden dokument PDF

### Interfejs Webowy
- Minimalistyczny formularz z polami:
  - URL strony do pobrania
  - Głębokość crawlingu (1-5)
  - Prefix URL (opcjonalny, ogranicza scraping do wskazanego zakresu)
  - Filtry HTML/JS (format: `START|||END`, każdy filtr w nowej linii)
- Progress indicator podczas scrapingu
- Przyciski do exportu ZIP i PDF po zakończeniu

## Plan Implementacji

### 1. Inicjalizacja projektu Go
```bash
go mod init github.com/user/scrapper
go get github.com/gocolly/colly/v2
go get github.com/PuerkitoBio/goquery
go get github.com/go-chi/chi/v5
go get github.com/jung-kurt/gofpdf
```

### 2. Core Scraper (`internal/scraper/scraper.go`)
- Konfiguracja Colly z `MaxDepth` parametrem
- Callback dla `OnHTML` do zbierania linków (`a[href]`, `img[src]`, `link[href]`, `script[src]`)
- Pobieranie i zapisywanie assets do `data/{project-id}/assets/`
- Rekurencyjne podążanie za linkami wewnątrz domeny
- Cache odwiedzonych URLi dla uniknięcia duplikatów

### 3. Transformacja Linków (`internal/scraper/processor.go`)
- Parsing HTML przez goquery
- Przekształcanie absolutnych URLi na względne ścieżki:
  - `https://example.com/page.html` → `pages/page.html`
  - `https://example.com/img/photo.jpg` → `assets/img/photo.jpg`
- Updatowanie atrybutów: `src`, `href`, `srcset`, `data-src`
- Zachowanie zewnętrznych linków jako absolutne

### 4. System Filtrowania (`internal/scraper/filter.go`)
- Struktura `FilterRule` z polami `StartPattern` i `EndPattern`
- Funkcja `ApplyFilters(html string, rules []FilterRule) string`
- Algorytm: znajdź `StartPattern`, znajdź następujący `EndPattern`, usuń zawartość między nimi
- Wsparcie dla multiple rules
- Zapisywanie rules w JSON: `data/{project-id}/filters.json`

### 5. REST API (`internal/api/`)

**Endpoints**:
- `POST /api/scrape` - rozpoczęcie scrapingu
  ```json
  {
    "url": "https://example.com",
    "url_prefix": "https://example.com/docs/ABC",
    "depth": 2,
    "filters": [
      {"start": "<script", "end": "</script>"},
      {"start": "<!-- ads-start", "end": "ads-end -->"}
    ]
  }
  ```
  Response: `{"project_id": "abc123", "status": "started"}`

- `GET /api/project/{id}/status` - status scrapingu
  ```json
  {
    "status": "completed|in_progress|failed",
    "progress": 75,
    "pages_downloaded": 15,
    "errors": []
  }
  ```

- `GET /api/project/{id}/export/zip` - download ZIP
- `POST /api/project/{id}/export/pdf` - generowanie i download PDF

### 6. Web UI (`web/`)
- Formularz HTML z walidacją
- JavaScript do obsługi:
  - Wysyłanie formularza (`POST /api/scrape`)
  - Polling statusu (`GET /api/project/{id}/status`)
  - Aktualizacja progress bar
  - Aktywacja przycisków export po zakończeniu
- CSS dla responsywności i stylu

### 7. Export do ZIP (`internal/export/zip.go`)
- Rekurencyjne pakowanie folderu projektu
- Streaming response dla dużych archiwów
- Nazwa pliku: `{project-id}.zip`

### 8. Export do PDF (`internal/export/pdf.go`)
- Iteracja przez wszystkie HTML pliki
- Konwersja HTML do text (strip tags, zachowanie struktury)
- Każda strona jako nowy rozdział w PDF
- Opcjonalnie: osadzanie obrazków inline
- Nazwa pliku: `{project-id}.pdf`

### 9. Konteneryzacja

**Dockerfile** (multi-stage build):
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o scrapper ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/scrapper .
COPY --from=builder /build/web ./web
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./scrapper"]
```

**docker-compose.yml**:
```yaml
version: '3.8'
services:
  scrapper:
    build: .
    ports:
      - "8900:8080"
    volumes:
      - ./data:/app/data
    environment:
      - PORT=8080
      - MAX_DEPTH_LIMIT=5
      - DATA_DIR=/app/data
```

### 10. Main Entry Point (`cmd/server/main.go`)
- Inicjalizacja Chi routera
- Mount static files z `/web` na `/`
- Mount API routes na `/api`
- Graceful shutdown
- Logging middleware

## Uruchomienie

### Lokalnie
```bash
# Instalacja dependencies
go mod download

# Uruchomienie
go run cmd/server/main.go

# Dostęp
# http://localhost:8900
```

### Docker
```bash
# Build i uruchomienie
docker-compose up --build

# Dostęp
# http://localhost:8900

# Zatrzymanie
docker-compose down
```

## Testowanie

### Flow testowy:
1. Otwórz `http://localhost:8900`
2. Wprowadź URL testowy (np. `https://example.com`)
3. Ustaw głębokość na 2
4. Opcjonalnie dodaj filtry:
   ```
   <script|||</script>
   <style|||</style>
   ```
5. Kliknij "Start Scraping"
6. Obserwuj progress
7. Po zakończeniu:
   - Sprawdź pliki w `data/{project-id}/`
   - Zweryfikuj transformację linków w HTML
   - Pobierz ZIP
   - Wygeneruj PDF

### Edge Cases:
- ✅ Zewnętrzne linki (pozostawić absolutne)
- ✅ Duplikaty URLi (cache odwiedzonych)
- ✅ Assets z innych domen (pobrać lub oznaczyć)
- ✅ Timeout dla wolnych stron
- ✅ Strony z błędami 404/500
- ✅ Cykliczne odnośniki (infinite loops)

## Konfiguracja

### Environment Variables
- `PORT` - port serwera (default: 8080)
- `MAX_DEPTH_LIMIT` - maksymalna głębokość (default: 5)
- `DATA_DIR` - katalog na dane (default: ./data)
- `TIMEOUT` - timeout dla requestów w sekundach (default: 30)
- `USER_AGENT` - custom User-Agent string

## Decisions Technologiczne

- **Go zamiast Python**: Wydajność, łatwa konteneryzacja (single binary), native concurrency
- **Colly**: Mature library z built-in depth control, lepsze od raw HTTP clienta
- **Vanilla JS**: Minimalny UI nie wymaga frameworka, mniej dependencies
- **File system storage**: Bez bazy danych, prostsze dla portable deployments
- **Struktura projektu**: Jeden projekt = jeden folder z assets i pages dla czytelności
- **PDF format**: Konsolidacja wszystkich stron w jeden dokument
- **Filters JSON**: Osobny plik dla łatwej edycji przed/po scrapingu
- **Relative links**: Przenośność - scrapowane strony działają offline bez serwera

## TODO / Przyszłe Rozszerzenia

- [ ] Wsparcie dla JavaScript-heavy stron (Playwright/Chromium headless)
- [ ] Planowanie zadań (cron-like scheduling)
- [ ] Lista i zarządzanie projektami w UI
- [ ] Podgląd pobranych stron w embedded iframe
- [ ] Import/export konfiguracji filtrów
- [ ] Logowanie szczegółowe z timestampami
- [ ] Autentykacja użytkowników (basic auth)
- [ ] Rate limiting dla scrapingu
- [ ] Robots.txt compliance check
- [ ] Sitemap.xml support
- [ ] Metryki i statystyki (liczba stron, rozmiar, czas)

## Production Deployment

### Quick Start with Docker

```bash
git clone <repo-url>
cd SCRAPPER
docker-compose up -d
```

Aplikacja będzie dostępna pod adresem `http://localhost:8900`.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port w kontenerze |
| `DATA_DIR` | `/app/data` | Project storage directory |
| `MAX_DEPTH_LIMIT` | `5` | Maximum crawling depth allowed |
| `TIMEOUT` | `30` | HTTP request timeout (seconds) |
| `USER_AGENT` | `WebScraper/1.0` | Custom User-Agent string |

### Troubleshooting

- Scraping fails immediately: sprawdź URL i logi `docker-compose logs scrapper`.
- Permission denied na `data/`: ustaw uprawnienia `chmod 755 data`.
- Export fails: upewnij się, że status projektu to `completed`.

### Monitoring

```bash
docker-compose logs -f scrapper
docker ps
docker stats web-scraper
ls -lah data/
```

### Backup & Restore

```bash
tar -czf scrapper-backup-$(date +%Y%m%d).tar.gz data/
tar -xzf scrapper-backup-DATE.tar.gz
```

## Licencja

MIT

## Autor

Created: 17 lutego 2026

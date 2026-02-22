# SCRAPPER

Web Scraper z interfejsem webowym napisany w Go. Aplikacja pobiera stronę i zasoby, zapisuje projekt na dysku, pozwala filtrować HTML/JS i eksportować wynik do ZIP lub jednego PDF.

## Status Aplikacji

- **Status**: ✅ Production Ready
- **Wersja**: 1.0
- **Data aktualizacji**: 22 lutego 2026
- **Zakres v1**: scraping statycznego HTML, transformacja linków, filtry, export ZIP/PDF, UI web, Docker

## Najważniejsze Funkcje

- Scraping stron z kontrolą głębokości (1-5)
- Pobieranie assetów (CSS, JS, obrazy, fonty)
- Transformacja linków do ścieżek względnych (offline portability)
- Filtry treści w formacie `START|||END`
- Status joba i progress przez API
- Export projektu do ZIP
- Export wszystkich stron do **jednego** pliku PDF
- Minimalny interfejs webowy (formularz + progress + export)

## Architektura i Struktura

- Backend: Go + Chi + Colly + goquery + gofpdf
- Frontend: Vanilla HTML/CSS/JS
- Storage: file system (`data/`)
- Deployment: Docker + docker-compose

Kluczowe katalogi:

- `cmd/server/` – entry point serwera
- `internal/api/` – routing, handlery, status
- `internal/scraper/` – scraping, transformacja linków, filtry, storage
- `internal/export/` – ZIP i PDF
- `web/` – UI
- `data/` – projekty runtime
- `ARCH/` – archiwum dokumentacji etapowej (agent files + poprzednie README/ORCHESTRATOR)

## Wymagania

- Go 1.21+
- Docker + Docker Compose (do uruchamiania kontenerowego)

## Uruchomienie Lokalnie

```bash
go mod download
go run cmd/server/main.go
```

Aplikacja: `http://localhost:8900`

## Uruchomienie w Docker

```bash
docker-compose up --build -d
```

Aplikacja: `http://localhost:8900`

Zatrzymanie:

```bash
docker-compose down
```

## Jak Używać (UI)

1. Otwórz `http://localhost:8900`
2. Podaj URL startowy
3. Ustaw depth (1-5)
4. Opcjonalnie dodaj filtry (`START|||END`, każdy w nowej linii)
5. Kliknij **Start Scraping**
6. Po zakończeniu pobierz ZIP lub wygeneruj PDF

## Jak Używać (API)

### Start scrape

`POST /api/scrape`

Przykład:

```bash
curl -X POST http://localhost:8900/api/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "depth": 2,
    "filters": [
      {"start": "<script", "end": "</script>"}
    ]
  }'
```

### Status projektu

`GET /api/project/{id}/status`

### Export ZIP

`GET /api/project/{id}/export/zip`

### Export PDF

`POST /api/project/{id}/export/pdf`

## Konfiguracja (ENV)

- `PORT` (default: `8080`, mapowany na host `8900` w compose)
- `DATA_DIR` (default: `/app/data` w kontenerze)
- `MAX_DEPTH_LIMIT` (default: `5`)
- `TIMEOUT` (default: `30`)
- `USER_AGENT` (default: `WebScraper/1.0`)

## Monitoring i Diagnostyka

```bash
docker-compose logs -f scrapper
docker ps
docker stats web-scraper
```

## Znane Ograniczenia (v1)

- Brak renderowania JavaScript (brak headless browsera)
- Crawling ograniczony do tej samej domeny
- Brak auth/cookies dla stron chronionych
- Brak robots.txt compliance
- Single-user design

## Dalszy Rozwój

Priorytetowy roadmap:

1. **Stabilność i QA**
   - automatyczne testy E2E
   - testy regresji eksportu PDF/ZIP
2. **Wydajność**
   - lepsze limity i throttling requestów
   - optymalizacja dużych projektów
3. **Funkcje scrapingu**
   - opcjonalny headless mode (Playwright)
   - rozszerzone reguły filtrowania
4. **UX i zarządzanie projektami**
   - lista projektów
   - podgląd rezultatów
   - re-run projektu na tych samych ustawieniach
5. **Operacyjność**
   - metryki i health telemetry
   - backup/restore workflows

## Dokumentacja Uzupełniająca

- `AGENTS.md` – pełny kontekst i decyzje projektowe
- `CHANGELOG.md` – historia zmian
- `ARCH/` – archiwalne materiały wdrożeniowe

## Licencja

MIT

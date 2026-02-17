# Web Scraper z Interfejsem Webowym

Aplikacja do pobierania stron internetowych wraz z obrazkami i zasobami, z możliwością filtrowania HTML/JS oraz exportu do ZIP/PDF.

## 🚀 Quick Start

### Lokalne uruchomienie

```bash
# Klonuj repozytorium
git clone https://github.com/robmoteka/SCRAPPER.git
cd SCRAPPER

# Zbuduj aplikację
go build -o scrapper ./cmd/server

# Uruchom serwer
./scrapper

# Otwórz w przeglądarce
# http://localhost:8080
```

### Docker

```bash
# Build i uruchomienie
docker-compose up --build

# Dostęp do UI
# http://localhost:8080

# Zatrzymanie
docker-compose down
```

## ✨ Funkcje

- ✅ Scraping stron z kontrolą głębokości (1-5 poziomów)
- ✅ Pobieranie wszystkich zasobów (obrazki, CSS, JS)
- ✅ Transformacja linków na ścieżki względne (offline-ready)
- ✅ Filtrowanie HTML/JS (usuwanie skryptów, reklam, etc.)
- ✅ Export do ZIP (pełna struktura projektu)
- ✅ Export do PDF (konsolidacja wszystkich stron w jeden dokument)
- ✅ Progress tracking w czasie rzeczywistym
- ✅ Prosty, responsywny interfejs webowy

## 📖 Jak używać

### Przez interfejs webowy

1. Otwórz http://localhost:8080 w przeglądarce
2. Wpisz URL strony do pobrania (np. `https://example.com`)
3. Ustaw głębokość crawlingu (1-5):
   - **1** = tylko główna strona
   - **2** = główna strona + bezpośrednie linki
   - **3-5** = głębsze poziomy
4. (Opcjonalnie) Dodaj filtry HTML/JS:
   ```
   <script|||</script>
   <!-- ads-start|||ads-end -->
   <div id="tracking"|||</div>
   ```
5. Kliknij **"Start Scraping"**
6. Obserwuj postęp w czasie rzeczywistym
7. Po zakończeniu:
   - **Pobierz ZIP** - kompletny offline-ready archiwum
   - **Generuj PDF** - wszystkie strony w jednym dokumencie

### Przez API

#### Rozpocznij scraping
```bash
curl -X POST http://localhost:8080/api/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "depth": 2,
    "filters": [
      {"start": "<script", "end": "</script>"},
      {"start": "<!-- ads", "end": "ads -->"}
    ]
  }'
```

Odpowiedź:
```json
{
  "project_id": "abc123-uuid",
  "status": "started"
}
```

#### Sprawdź status
```bash
curl http://localhost:8080/api/project/{project_id}/status
```

#### Pobierz ZIP
```bash
curl http://localhost:8080/api/project/{project_id}/export/zip -o project.zip
```

#### Generuj PDF
```bash
curl -X POST http://localhost:8080/api/project/{project_id}/export/pdf -o project.pdf
```

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
  - Filtry HTML/JS (format: `START|||END`, każdy filtr w nowej linii)
- Progress indicator podczas scrapingu
- Przyciski do exportu ZIP i PDF po zakończeniu

## ⚙️ Konfiguracja

### Environment Variables
- `PORT` - port serwera (default: 8080)
- `MAX_DEPTH_LIMIT` - maksymalna głębokość (default: 5)
- `DATA_DIR` - katalog na dane (default: ./data)
- `TIMEOUT` - timeout dla requestów w sekundach (default: 30)
- `USER_AGENT` - custom User-Agent string (default: WebScraper/1.0)

## 🔍 Testowanie

### Flow testowy:
1. Otwórz `http://localhost:8080`
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

## 🏗️ Architektura

### Decisions Technologiczne

- **Go zamiast Python**: Wydajność, łatwa konteneryzacja (single binary), native concurrency
- **Colly**: Mature library z built-in depth control, lepsze od raw HTTP clienta
- **Vanilla JS**: Minimalny UI nie wymaga frameworka, mniej dependencies
- **File system storage**: Bez bazy danych, prostsze dla portable deployments
- **Struktura projektu**: Jeden projekt = jeden folder z assets i pages dla czytelności
- **PDF format**: Konsolidacja wszystkich stron w jeden dokument
- **Filters JSON**: Osobny plik dla łatwej edycji przed/po scrapingu
- **Relative links**: Przenośność - scrapowane strony działają offline bez serwera

## 📝 Known Limitations

### Current Scope (v1.0)
- Static HTML only (no JavaScript rendering)
- Same-domain crawling only
- No authentication support
- No rate limiting
- No robots.txt compliance check

## Licencja

MIT

## Autor

Created: 17 lutego 2026

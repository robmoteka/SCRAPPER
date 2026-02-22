# AGENT 6: Web UI Frontend

**Phase**: Frontend  
**Zadania**: 14-16  
**Dependencies**: Agent 4 (API działające)  
**Estimated Time**: 40-50 minut

---

## Cel Agenta

Stworzenie minimalnego, responsywnego interfejsu webowego z formularzem, progress tracking i export buttons.

---

## Prerequisites

Przed rozpoczęciem sprawdź:
- [x] Agent 4 ukończony (API działa)
- [x] Endpoints: POST /api/scrape, GET /api/status, GET /api/export/zip, POST /api/export/pdf
- [x] Folder `web/` istnieje

---

## Zadania do Wykonania

### ✅ Zadanie 14: Create web/index.html

**Cel**: Minimalny HTML formularz z UI elements.

**Plik**: `web/index.html`

```html
<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Web Scraper</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>🕷️ Web Scraper</h1>
            <p>Pobierz stronę internetową wraz z zasobami</p>
        </header>

        <main>
            <!-- Scraping Form -->
            <div id="scrapeForm" class="card">
                <h2>Konfiguracja</h2>
                
                <form id="form">
                    <div class="form-group">
                        <label for="url">URL strony *</label>
                        <input 
                            type="url" 
                            id="url" 
                            name="url" 
                            placeholder="https://example.com"
                            required
                        >
                    </div>

                    <div class="form-group">
                        <label for="depth">Głębokość crawlingu (1-5)</label>
                        <input 
                            type="number" 
                            id="depth" 
                            name="depth" 
                            min="1" 
                            max="5" 
                            value="2"
                            required
                        >
                        <small>Liczba poziomów podstron do pobrania</small>
                    </div>

                    <div class="form-group">
                        <label for="filters">Filtry HTML/JS (opcjonalne)</label>
                        <textarea 
                            id="filters" 
                            name="filters" 
                            rows="5"
                            placeholder="<script|||</script>&#10;<style|||</style>&#10;<!-- ads-start|||ads-end -->"
                        ></textarea>
                        <small>Format: START|||END (każdy filtr w nowej linii)</small>
                    </div>

                    <button type="submit" class="btn btn-primary" id="startBtn">
                        🚀 Rozpocznij Scraping
                    </button>
                </form>
            </div>

            <!-- Progress Display -->
            <div id="progressCard" class="card hidden">
                <h2>Postęp</h2>
                
                <div class="progress-bar">
                    <div id="progressFill" class="progress-fill"></div>
                </div>
                
                <p id="progressText" class="progress-text">Inicjalizacja...</p>
                
                <div id="progressStats" class="stats">
                    <div class="stat">
                        <span class="stat-label">Status:</span>
                        <span id="statusText" class="stat-value">-</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">Pobrane strony:</span>
                        <span id="downloadedText" class="stat-value">0</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">Łącznie:</span>
                        <span id="totalText" class="stat-value">0</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">Aktualny URL:</span>
                        <span id="currentUrlText" class="stat-value url-text">-</span>
                    </div>
                </div>

                <div id="errors" class="errors hidden">
                    <h3>⚠️ Błędy:</h3>
                    <ul id="errorsList"></ul>
                </div>

                <button id="cancelBtn" class="btn btn-secondary hidden">
                    ❌ Anuluj
                </button>
            </div>

            <!-- Export Options -->
            <div id="exportCard" class="card hidden">
                <h2>✅ Zakończono!</h2>
                <p>Projekt gotowy do eksportu</p>
                
                <div class="export-buttons">
                    <button id="exportZipBtn" class="btn btn-primary">
                        📦 Pobierz ZIP
                    </button>
                    <button id="exportPdfBtn" class="btn btn-primary">
                        📄 Generuj PDF
                    </button>
                </div>

                <button id="newScrapeBtn" class="btn btn-secondary">
                    🔄 Nowy Scraping
                    </button>
            </div>
        </main>

        <footer>
            <p>Web Scraper v1.0 | Powered by Go + Colly</p>
        </footer>
    </div>

    <script src="app.js"></script>
</body>
</html>
```

**Verification**: Otwórz `web/index.html` w przeglądarce (static file, bez serwera).

---

### ✅ Zadanie 15: Create web/style.css

**Cel**: Minimal, responsive styling.

**Plik**: `web/style.css`

```css
/* Reset & Base */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #333;
    line-height: 1.6;
    min-height: 100vh;
    padding: 20px;
}

.container {
    max-width: 800px;
    margin: 0 auto;
}

/* Header */
header {
    text-align: center;
    color: white;
    margin-bottom: 40px;
}

header h1 {
    font-size: 2.5rem;
    margin-bottom: 10px;
}

header p {
    font-size: 1.1rem;
    opacity: 0.9;
}

/* Card */
.card {
    background: white;
    border-radius: 12px;
    padding: 30px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
    margin-bottom: 20px;
    transition: transform 0.2s;
}

.card:hover {
    transform: translateY(-2px);
}

.card h2 {
    margin-bottom: 20px;
    color: #667eea;
}

/* Form */
.form-group {
    margin-bottom: 20px;
}

.form-group label {
    display: block;
    margin-bottom: 8px;
    font-weight: 600;
    color: #555;
}

.form-group input,
.form-group textarea {
    width: 100%;
    padding: 12px;
    border: 2px solid #e0e0e0;
    border-radius: 8px;
    font-size: 1rem;
    transition: border-color 0.3s;
}

.form-group input:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #667eea;
}

.form-group textarea {
    resize: vertical;
    font-family: monospace;
}

.form-group small {
    display: block;
    margin-top: 5px;
    color: #999;
    font-size: 0.875rem;
}

/* Buttons */
.btn {
    padding: 12px 24px;
    font-size: 1rem;
    font-weight: 600;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s;
    display: inline-block;
    text-align: center;
}

.btn-primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
}

.btn-secondary {
    background: #f0f0f0;
    color: #333;
}

.btn-secondary:hover {
    background: #e0e0e0;
}

.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

/* Progress Bar */
.progress-bar {
    width: 100%;
    height: 30px;
    background: #e0e0e0;
    border-radius: 15px;
    overflow: hidden;
    margin-bottom: 15px;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
    width: 0%;
    transition: width 0.5s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: 600;
}

.progress-text {
    text-align: center;
    font-size: 1.1rem;
    color: #666;
    margin-bottom: 20px;
}

/* Stats */
.stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 15px;
    margin-bottom: 20px;
}

.stat {
    background: #f8f9fa;
    padding: 12px;
    border-radius: 8px;
}

.stat-label {
    display: block;
    font-size: 0.875rem;
    color: #999;
    margin-bottom: 5px;
}

.stat-value {
    display: block;
    font-size: 1.1rem;
    font-weight: 600;
    color: #333;
}

.url-text {
    font-size: 0.875rem;
    word-break: break-all;
    color: #667eea;
}

/* Errors */
.errors {
    background: #fff3cd;
    border: 2px solid #ffc107;
    border-radius: 8px;
    padding: 15px;
    margin-bottom: 20px;
}

.errors h3 {
    margin-bottom: 10px;
    color: #856404;
}

.errors ul {
    list-style-position: inside;
    color: #856404;
}

/* Export Buttons */
.export-buttons {
    display: flex;
    gap: 15px;
    margin-bottom: 15px;
}

.export-buttons .btn {
    flex: 1;
}

/* Utility */
.hidden {
    display: none !important;
}

/* Footer */
footer {
    text-align: center;
    color: white;
    margin-top: 40px;
    opacity: 0.8;
}

/* Responsive */
@media (max-width: 600px) {
    header h1 {
        font-size: 2rem;
    }

    .card {
        padding: 20px;
    }

    .export-buttons {
        flex-direction: column;
    }

    .stats {
        grid-template-columns: 1fr;
    }
}
```

**Verification**: Sprawdź styling w przeglądarce.

---

### ✅ Zadanie 16: Create web/app.js

**Cel**: Frontend logic z Fetch API, polling, event handling.

**Plik**: `web/app.js`

```javascript
// Global state
let currentProjectId = null;
let pollingInterval = null;

// DOM Elements
const form = document.getElementById('form');
const scrapeCard = document.getElementById('scrapeForm');
const progressCard = document.getElementById('progressCard');
const exportCard = document.getElementById('exportCard');

const progressFill = document.getElementById('progressFill');
const progressText = document.getElementById('progressText');
const statusText = document.getElementById('statusText');
const downloadedText = document.getElementById('downloadedText');
const totalText = document.getElementById('totalText');
const currentUrlText = document.getElementById('currentUrlText');
const errorsDiv = document.getElementById('errors');
const errorsList = document.getElementById('errorsList');

const startBtn = document.getElementById('startBtn');
const cancelBtn = document.getElementById('cancelBtn');
const exportZipBtn = document.getElementById('exportZipBtn');
const exportPdfBtn = document.getElementById('exportPdfBtn');
const newScrapeBtn = document.getElementById('newScrapeBtn');

// Event Listeners
form.addEventListener('submit', handleFormSubmit);
cancelBtn.addEventListener('click', handleCancel);
exportZipBtn.addEventListener('click', handleExportZip);
exportPdfBtn.addEventListener('click', handleExportPdf);
newScrapeBtn.addEventListener('click', handleNewScrape);

// Handle form submission
async function handleFormSubmit(e) {
    e.preventDefault();

    // Get form data
    const url = document.getElementById('url').value;
    const depth = parseInt(document.getElementById('depth').value);
    const filtersText = document.getElementById('filters').value;

    // Parse filters
    const filters = parseFilters(filtersText);

    // Prepare request
    const requestData = {
        url: url,
        depth: depth,
        filters: filters
    };

    try {
        // Start scraping
        const response = await fetch('/api/scrape', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to start scraping');
        }

        const data = await response.json();
        currentProjectId = data.project_id;

        // Show progress card
        showProgress();

        // Start polling
        startPolling();

    } catch (error) {
        alert('Error: ' + error.message);
    }
}

// Parse filters from textarea
function parseFilters(text) {
    if (!text.trim()) return [];

    const lines = text.split('\n');
    const filters = [];

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        const parts = trimmed.split('|||');
        if (parts.length === 2) {
            filters.push({
                start: parts[0].trim(),
                end: parts[1].trim()
            });
        }
    }

    return filters;
}

// Show progress card
function showProgress() {
    scrapeCard.classList.add('hidden');
    progressCard.classList.remove('hidden');
    exportCard.classList.add('hidden');
    cancelBtn.classList.remove('hidden');
}

// Show export card
function showExport() {
    scrapeCard.classList.add('hidden');
    progressCard.classList.add('hidden');
    exportCard.classList.remove('hidden');
}

// Start status polling
function startPolling() {
    // Initial poll
    pollStatus();

    // Poll every 2 seconds
    pollingInterval = setInterval(pollStatus, 2000);
}

// Stop status polling
function stopPolling() {
    if (pollingInterval) {
        clearInterval(pollingInterval);
        pollingInterval = null;
    }
}

// Poll status endpoint
async function pollStatus() {
    if (!currentProjectId) return;

    try {
        const response = await fetch(`/api/project/${currentProjectId}/status`);
        
        if (!response.ok) {
            throw new Error('Failed to get status');
        }

        const data = await response.json();
        updateProgress(data);

        // Check if completed
        if (data.status === 'completed') {
            stopPolling();
            showExport();
        } else if (data.status === 'failed') {
            stopPolling();
            progressText.textContent = '❌ Scraping nie powiódł się';
            cancelBtn.textContent = 'Powrót';
        }

    } catch (error) {
        console.error('Polling error:', error);
    }
}

// Update progress UI
function updateProgress(data) {
    // Progress bar
    const progress = data.progress || 0;
    progressFill.style.width = progress + '%';
    progressFill.textContent = progress + '%';

    // Status
    statusText.textContent = formatStatus(data.status);
    downloadedText.textContent = data.pages_downloaded || 0;
    totalText.textContent = data.total_pages || 0;
    currentUrlText.textContent = data.current_url || '-';

    // Progress text
    if (data.status === 'in_progress') {
        progressText.textContent = `Pobieranie w toku... (${data.pages_downloaded}/${data.total_pages})`;
    } else if (data.status === 'completed') {
        progressText.textContent = '✅ Scraping zakończony!';
    } else if (data.status === 'failed') {
        progressText.textContent = '❌ Scraping nie powiódł się';
    }

    // Errors
    if (data.errors && data.errors.length > 0) {
        errorsDiv.classList.remove('hidden');
        errorsList.innerHTML = '';
        data.errors.forEach(error => {
            const li = document.createElement('li');
            li.textContent = error;
            errorsList.appendChild(li);
        });
    } else {
        errorsDiv.classList.add('hidden');
    }
}

// Format status for display
function formatStatus(status) {
    const statusMap = {
        'started': '🔄 Rozpoczęty',
        'in_progress': '⏳ W toku',
        'completed': '✅ Zakończony',
        'failed': '❌ Błąd'
    };
    return statusMap[status] || status;
}

// Handle cancel
function handleCancel() {
    stopPolling();
    handleNewScrape();
}

// Handle new scrape
function handleNewScrape() {
    currentProjectId = null;
    stopPolling();

    // Reset form
    form.reset();

    // Reset UI
    scrapeCard.classList.remove('hidden');
    progressCard.classList.add('hidden');
    exportCard.classList.add('hidden');

    progressFill.style.width = '0%';
    errorsDiv.classList.add('hidden');
}

// Handle ZIP export
async function handleExportZip() {
    if (!currentProjectId) return;

    try {
        window.location.href = `/api/project/${currentProjectId}/export/zip`;
    } catch (error) {
        alert('Export error: ' + error.message);
    }
}

// Handle PDF export
async function handleExportPdf() {
    if (!currentProjectId) return;

    try {
        // Show loading state
        exportPdfBtn.disabled = true;
        exportPdfBtn.textContent = '⏳ Generowanie PDF...';

        const response = await fetch(`/api/project/${currentProjectId}/export/pdf`, {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('PDF generation failed');
        }

        // Download PDF
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${currentProjectId}.pdf`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);

    } catch (error) {
        alert('PDF export error: ' + error.message);
    } finally {
        // Reset button state
        exportPdfBtn.disabled = false;
        exportPdfBtn.textContent = '📄 Generuj PDF';
    }
}
```

**Verification**:
```bash
# Test w przeglądarce (wymaga działającego serwera)
go run cmd/server/main.go
# Otwórz http://localhost:8080
```

---

## Expected Output Files

Po ukończeniu Agenta 6:

```
✅ web/index.html
✅ web/style.css
✅ web/app.js
✅ Frontend działa z backend API
✅ Responsive design
```

---

## Verification Checklist

- [ ] Frontend ładuje się poprawnie
- [ ] Formularz waliduje input (URL, depth 1-5)
- [ ] Submit wywołuje POST /api/scrape
- [ ] Progress card pokazuje real-time updates
- [ ] Progress bar działa (0-100%)
- [ ] Export buttons działają (ZIP immediate, PDF z loading)
- [ ] "Nowy Scraping" resetuje UI
- [ ] Responsywny na mobile

---

## Manual E2E Test

1. Uruchom serwer: `go run cmd/server/main.go`
2. Otwórz `http://localhost:8080`
3. Wprowadź URL: `https://example.com`
4. Głębokość: `2`
5. Filtry: `<script|||</script>`
6. Kliknij "Rozpocznij Scraping"
7. Obserwuj progress (2s polling)
8. Po zakończeniu kliknij "Pobierz ZIP"
9. Kliknij "Generuj PDF" (loading state)
10. Sprawdź oba pliki
11. Kliknij "Nowy Scraping" - formularz reset

---

## Next Agent

Po ukończeniu **Agent 6**, przejdź do:
👉 **AGENT_07_DOCKER.md** (Konteneryzacja)

**Prerequisites verified**:
- ✅ Frontend kompletny
- ✅ Integracja z API działa
- ✅ Wszystkie features UX gotowe

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026

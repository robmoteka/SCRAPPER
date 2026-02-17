# ORCHESTRATOR.md - Orkiestracja Projektu Web Scraper

**Data**: 17 lutego 2026  
**Projekt**: Web Scraper z interfejsem webowym w Go  
**Status**: Ready for orchestrated implementation

---

## Przegląd Orkiestracji

Projekt podzielony jest na **8 agentów specjalistycznych**, każdy odpowiedzialny za konkretną fazę implementacji. Agenci działają sekwencyjnie, gdzie każdy kolejny agent buduje na fundamencie poprzedniego.

### Zależności między Agentami

```
Agent 1 (Foundation) 
    ↓
Agent 2 (Scraping) ← wymaga Agent 1 (modele, struktura)
    ↓
Agent 3 (Filtering) ← wymaga Agent 2 (scraper logic)
    ↓
Agent 4 (API) ← wymaga Agent 1-3 (wszystkie core features)
    ↓
Agent 5 (Export) ← wymaga Agent 4 (API endpoints)
    ↓
Agent 6 (Frontend) ← wymaga Agent 4 (API działające)
    ↓
Agent 7 (Docker) ← wymaga Agent 1-6 (cała aplikacja)
    ↓
Agent 8 (QA) ← wymaga Agent 7 (deployment gotowy)
```

---

## Agenci Specjalistyczni

### 🏗️ Agent 1: Foundation & Bootstrap
**Plik**: `AGENT_01_FOUNDATION.md`  
**Zadania**: 1-4 (Initialize, Structure, Models, Router)  
**Dependency**: Żadne  
**Output**: Podstawowy szkielet aplikacji Go + struktury danych

---

### 🕷️ Agent 2: Core Scraping Engine
**Plik**: `AGENT_02_SCRAPING.md`  
**Zadania**: 5-6 (Colly integration, Link transformation)  
**Dependency**: Agent 1 (modele, folder structure)  
**Output**: Działający silnik scrapingu z depth control + transformacja linków

---

### 🔧 Agent 3: Filtering & Storage
**Plik**: `AGENT_03_FILTERING.md`  
**Zadania**: 7-8 (HTML filtering, File storage)  
**Dependency**: Agent 2 (scraper logic)  
**Output**: System filtrowania HTML/JS + persystencja projektów

---

### 🌐 Agent 4: API Layer
**Plik**: `AGENT_04_API.md`  
**Zadania**: 9-10 (Handlers, Async scraping)  
**Dependency**: Agent 1-3 (wszystkie core features)  
**Output**: REST API endpoints + status tracking

---

### 📦 Agent 5: Export Features
**Plik**: `AGENT_05_EXPORT.md`  
**Zadania**: 11-13 (ZIP export, PDF export, API handlers)  
**Dependency**: Agent 4 (API layer)  
**Output**: Export ZIP/PDF + odpowiednie endpointy

---

### 💻 Agent 6: Web UI Frontend
**Plik**: `AGENT_06_FRONTEND.md`  
**Zadania**: 14-16 (HTML, CSS, JavaScript)  
**Dependency**: Agent 4 (API działające)  
**Output**: Responsywny interfejs webowy

---

### 🐳 Agent 7: Containerization
**Plik**: `AGENT_07_DOCKER.md`  
**Zadania**: 17-19 (Dockerfile, docker-compose, testing)  
**Dependency**: Agent 1-6 (cała aplikacja)  
**Output**: Konteneryzacja + deployment ready

---

### ✅ Agent 8: Polish & QA
**Plik**: `AGENT_08_QA.md`  
**Zadania**: 20-22 (Edge cases, Logging, Documentation)  
**Dependency**: Agent 7 (deployment gotowy)  
**Output**: Production-ready application

---

## Workflow dla GitHub Copilot

### Strategia Implementacji

#### Krok 1: Inicjalizacja
```bash
# GitHub Copilot: Rozpocznij od Agenta 1
@workspace /agent AGENT_01_FOUNDATION.md
```

#### Krok 2: Sekwencyjna Implementacja
Dla każdego agenta (1-8):
1. **Otwórz plik agenta**: `AGENT_XX_NAME.md`
2. **Wywołaj Copilot**:
   ```bash
   @workspace Zaimplementuj wszystkie zadania z tego pliku agenta
   ```
3. **Weryfikacja przed przejściem dalej**:
   - [ ] Wszystkie pliki utworzone
   - [ ] Kod kompiluje się (`go build`)
   - [ ] Testy jednostkowe pass (jeśli applicable)
   - [ ] Manualny smoke test funkcjonalności

#### Krok 3: Checkpoint po każdym agencie
```bash
# Commit postępu
git add .
git commit -m "✅ Agent X completed: [nazwa fazy]"
```

---

## Kompatybilność z GitHub Copilot Workspace

### Użycie w Copilot Chat
```
# Strategia 1: Full Orchestration
"Implementuj projekt zgodnie z ORCHESTRATOR.md, zaczynając od Agent 1"

# Strategia 2: Agent-by-Agent
"Załącz AGENT_01_FOUNDATION.md i zaimplementuj wszystkie zadania"
"Po ukończeniu przejdź do AGENT_02_SCRAPING.md"

# Strategia 3: Task-Specific
"Z AGENT_04_API.md zaimplementuj zadanie 9: API handlers"
```

### Prompt Templates

#### Rozpoczęcie Agenta
```
@workspace #file:AGENT_XX_NAME.md

Przeczytaj plik agenta i zaimplementuj wszystkie zadania sekwencyjnie. 
Przed rozpoczęciem pokaż plan implementacji.
```

#### Kontynuacja po błędzie
```
@workspace #file:AGENT_XX_NAME.md

Zadanie [N] nie powiodło się. Przeanalizuj błąd i zaproponuj fix.
Kontynuuj pozostałe zadania po naprawie.
```

#### Weryfikacja
```
@workspace

Zweryfikuj implementację Agent X:
- Czy wszystkie pliki z "Expected Files" istnieją?
- Czy kod kompiluje się bez błędów?
- Czy API endpoints działają zgodnie z spec?
```

---

## Status Tracking Matrix

| Agent | Status | Zadania | Pliki | Testy | Notes |
|-------|--------|---------|-------|-------|-------|
| 1: Foundation | ⏳ TODO | 1-4 | 0/5 | ⬜ | Inicjalizacja projektu |
| 2: Scraping | ⏳ TODO | 5-6 | 0/2 | ⬜ | Wymaga Agent 1 |
| 3: Filtering | ⏳ TODO | 7-8 | 0/2 | ⬜ | Wymaga Agent 2 |
| 4: API | ⏳ TODO | 9-10 | 0/2 | ⬜ | Wymaga Agent 1-3 |
| 5: Export | ⏳ TODO | 11-13 | 0/3 | ⬜ | Wymaga Agent 4 |
| 6: Frontend | ⏳ TODO | 14-16 | 0/3 | ⬜ | Wymaga Agent 4 |
| 7: Docker | ⏳ TODO | 17-19 | 0/2 | ⬜ | Wymaga Agent 1-6 |
| 8: QA | ⏳ TODO | 20-22 | 0/1 | ⬜ | Wymaga Agent 7 |

**Legenda**:
- ⏳ TODO - Oczekuje na implementację
- 🟡 IN PROGRESS - W trakcie realizacji
- ✅ DONE - Ukończone
- ❌ BLOCKED - Zablokowane przez dependency
- ⬜ - Test nie uruchomiony
- ✅ - Test passed
- ❌ - Test failed

---

## Verification Checklist

### Po każdym agencie:

```bash
# Kompilacja
go build ./...

# Formatting
go fmt ./...

# Linting (optional)
# go vet ./...

# Run server (smoke test)
go run cmd/server/main.go &
curl http://localhost:8080

# Kill server
pkill -f "cmd/server/main.go"
```

### Final Integration Test (po Agent 6):
1. Uruchom serwer lokalnie
2. Otwórz `http://localhost:8080`
3. Wprowadź URL testowy: `https://example.com`
4. Głębokość: 2
5. Dodaj filtr: `<script|||</script>`
6. Rozpocznij scraping
7. Zweryfikuj:
   - [ ] Progress tracking działa
   - [ ] Pliki zapisują się w `data/`
   - [ ] ZIP download działa
   - [ ] PDF download działa

### Docker Test (po Agent 7):
```bash
docker-compose up --build
# Test jak wyżej na http://localhost:8080
docker-compose down
```

---

## Rollback Strategy

### Jeśli agent nie powiedzie się:

1. **Diagnoza**:
   ```bash
   git diff # Zobacz zmiany
   go build ./... # Sprawdź błędy kompilacji
   ```

2. **Rollback**:
   ```bash
   git checkout -- . # Cofnij wszystkie zmiany
   git clean -fd # Usuń nowe pliki
   ```

3. **Analiza**:
   - Przeczytaj ponownie plik agenta
   - Sprawdź dependency (czy poprzedni agent ukończony?)
   - Sprawdź błędy w AGENTS.md context

4. **Retry** z adjusted prompt:
   ```
   @workspace #file:AGENT_XX_NAME.md
   
   Poprzednia implementacja nie powiodła się z powodu: [error].
   Zaimplementuj ponownie, uwzględniając: [fix strategy]
   ```

---

## Communication Protocol

### Format raportowania postępu:

```markdown
## Agent X: [NAZWA] - Status Update

**Started**: [timestamp]
**Completed**: [timestamp]
**Duration**: [minutes]

### Implemented:
- ✅ Zadanie N: [opis]
- ✅ Zadanie N+1: [opis]

### Created Files:
- `path/to/file1.go` - [purpose]
- `path/to/file2.go` - [purpose]

### Tests:
- ✅ Kompilacja: Success
- ✅ Smoke test: Success
- ⬜ Unit tests: N/A (to be added in Agent 8)

### Next Agent:
Agent X+1 ready to start (dependencies satisfied)
```

---

## Critical Success Factors

### Dla GitHub Copilot:

1. **Context Window Management**:
   - Zawsze załącz relevantny plik agenta
   - Referencuj AGENTS.md dla decisions
   - Używaj README.md dla technical specs

2. **Incremental Verification**:
   - Po każdym pliku: compile check
   - Po każdym module: smoke test
   - Po każdym agencie: integration check

3. **Dependency Awareness**:
   - Nie zaczynaj Agent N jeśli Agent N-1 incomplete
   - Sprawdź "Expected Files" z poprzednich agentów
   - Verify imports resolution

4. **Error Recovery**:
   - Capture error messages verbatim
   - Reference exact line numbers
   - Provide full stack trace if available

---

## Quick Command Reference

```bash
# Start fresh agent work
@workspace #file:AGENT_0X_NAME.md Implement all tasks

# Check status
go build ./... && echo "✅ Build OK"

# Run server
go run cmd/server/main.go

# Test API endpoint
curl -X POST http://localhost:8080/api/scrape \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","depth":2}'

# Docker quick test
docker-compose up --build -d && sleep 5 && curl http://localhost:8080

# Cleanup
docker-compose down && rm -rf data/*
```

---

## Final Checklist (pre-delivery)

- [ ] Wszystkie 8 agentów completed (status matrix ✅)
- [ ] Aplikacja build bez warnings
- [ ] Docker container runs successfully
- [ ] Manual E2E test passed (scraping → export)
- [ ] README.md updated z deployment instructions
- [ ] AGENTS.md zaktualizowany (status: Production Ready)
- [ ] data/ folder w .gitignore
- [ ] No hardcoded credentials/secrets
- [ ] Logging działa properly
- [ ] Error handling graceful

---

**Ready to begin**: Invoke Agent 1 with:
```bash
@workspace #file:AGENT_01_FOUNDATION.md Rozpocznij implementację Agenta 1
```

**Last Updated**: 17 lutego 2026  
**Orchestrator Version**: 1.0

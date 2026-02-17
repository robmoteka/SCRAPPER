# AGENT 7: Containerization

**Phase**: Containerization  
**Zadania**: 17-19  
**Dependencies**: Agent 1-6 (cała aplikacja)  
**Estimated Time**: 30-40 minut

---

## Cel Agenta

Konteneryzacja aplikacji z multi-stage Docker build, docker-compose setup oraz deployment testing.

---

## Prerequisites

Przed rozpoczęciem sprawdź:
- [x] Agent 1-6 ukończone (full application ready)
- [x] `go build ./cmd/server` działa lokalnie
- [x] Frontend + API + export działają E2E
- [x] Docker zainstalowany

---

## Zadania do Wykonania

### ✅ Zadanie 17: Write Dockerfile (multi-stage)

**Cel**: Efektywny multi-stage build (builder + runtime).

**Plik**: `Dockerfile`

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install dependencies for Go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o scrapper ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/scrapper .

# Copy web assets
COPY --from=builder /build/web ./web

# Create data directory with proper permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Set environment variables
ENV PORT=8080 \
    DATA_DIR=/app/data \
    MAX_DEPTH_LIMIT=5 \
    TIMEOUT=30

# Run application
CMD ["./scrapper"]
```

**Verification**:
```bash
docker build -t web-scraper:latest .
docker images | grep web-scraper
```

---

### ✅ Zadanie 18: Write docker-compose.yml

**Cel**: Orchestration z volume persistence i environment configuration.

**Plik**: `docker-compose.yml`

```yaml
version: '3.8'

services:
  scrapper:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: web-scraper
    ports:
      - "8080:8080"
    volumes:
      # Persist scraped data
      - ./data:/app/data
    environment:
      # Server configuration
      - PORT=8080
      - DATA_DIR=/app/data
      
      # Scraping limits
      - MAX_DEPTH_LIMIT=5
      - TIMEOUT=30
      
      # User-Agent (optional)
      - USER_AGENT=WebScraper/1.0
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
    networks:
      - scrapper-network

networks:
  scrapper-network:
    driver: bridge
```

**Verification**:
```bash
docker-compose config  # Validate syntax
```

---

### ✅ Zadanie 19: Test in Docker environment

**Cel**: Kompletne testowanie deploymentu w Docker.

**Test Script**: `test_docker.sh`

```bash
#!/bin/bash

echo "🧪 Testing Web Scraper Docker Deployment"
echo "=========================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Build image
echo -e "\n${YELLOW}Step 1: Building Docker image...${NC}"
docker-compose build
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# Step 2: Start container
echo -e "\n${YELLOW}Step 2: Starting container...${NC}"
docker-compose up -d
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Container start failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Container started${NC}"

# Step 3: Wait for healthcheck
echo -e "\n${YELLOW}Step 3: Waiting for health check...${NC}"
sleep 10

HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' web-scraper)
echo "Health status: $HEALTH_STATUS"

# Step 4: Test API endpoints
echo -e "\n${YELLOW}Step 4: Testing API endpoints...${NC}"

# Test root (should serve HTML)
echo "Testing GET /"
curl -f -s http://localhost:8080/ > /dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Frontend accessible${NC}"
else
    echo -e "${RED}❌ Frontend not accessible${NC}"
fi

# Test scrape endpoint
echo "Testing POST /api/scrape"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/scrape \
    -H "Content-Type: application/json" \
    -d '{
        "url": "https://example.com",
        "depth": 1,
        "filters": []
    }')

PROJECT_ID=$(echo $RESPONSE | grep -o '"project_id":"[^"]*' | cut -d'"' -f4)

if [ -n "$PROJECT_ID" ]; then
    echo -e "${GREEN}✅ Scrape endpoint works${NC}"
    echo "Project ID: $PROJECT_ID"
else
    echo -e "${RED}❌ Scrape endpoint failed${NC}"
    echo "Response: $RESPONSE"
fi

# Wait for scraping to complete
echo -e "\n${YELLOW}Waiting 30s for scraping to complete...${NC}"
sleep 30

# Test status endpoint
if [ -n "$PROJECT_ID" ]; then
    echo "Testing GET /api/project/$PROJECT_ID/status"
    STATUS_RESPONSE=$(curl -s http://localhost:8080/api/project/$PROJECT_ID/status)
    echo "Status: $STATUS_RESPONSE"
    
    if echo "$STATUS_RESPONSE" | grep -q "completed"; then
        echo -e "${GREEN}✅ Status endpoint works${NC}"
    else
        echo -e "${YELLOW}⚠️  Scraping not completed yet${NC}"
    fi
fi

# Step 5: Check data persistence
echo -e "\n${YELLOW}Step 5: Checking data persistence...${NC}"
if [ -d "./data/$PROJECT_ID" ]; then
    echo -e "${GREEN}✅ Data directory created${NC}"
    echo "Files:"
    ls -lh ./data/$PROJECT_ID/
else
    echo -e "${RED}❌ Data directory not created${NC}"
fi

# Step 6: Check logs
echo -e "\n${YELLOW}Step 6: Container logs (last 20 lines):${NC}"
docker-compose logs --tail=20 scrapper

# Step 7: Cleanup option
echo -e "\n${YELLOW}Test complete!${NC}"
echo "To stop container: docker-compose down"
echo "To view logs: docker-compose logs -f scrapper"
echo "To restart: docker-compose restart"

echo -e "\n${GREEN}=========================================="
echo "✅ Docker deployment test finished"
echo "==========================================${NC}"
```

**Make executable**:
```bash
chmod +x test_docker.sh
```

**Run test**:
```bash
./test_docker.sh
```

---

## Additional: .dockerignore

**Plik**: `.dockerignore`

```
# Git
.git
.gitignore

# Data and temp files
data/
*.zip
*.pdf

# Documentation
README.md
AGENTS.md
ORCHESTRATOR.md
AGENT_*.md

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db

# Build artifacts
scrapper
*.exe

# Test files
test_*.go
*_test.go
```

---

## Expected Output Files

Po ukończeniu Agenta 7:

```
✅ Dockerfile
✅ docker-compose.yml
✅ .dockerignore
✅ test_docker.sh (executable)
✅ Docker image builds successfully
✅ Container runs and passes health checks
✅ E2E test w Docker passes
```

---

## Verification Checklist

### Build & Start
- [ ] `docker-compose build` kompiluje bez błędów
- [ ] Image size reasonable (<50MB dla Alpine)
- [ ] `docker-compose up -d` uruchamia container
- [ ] Container healthcheck passes (green status)

### Functionality
- [ ] Frontend dostępny na http://localhost:8080
- [ ] API endpoints działają (POST /scrape, GET /status)
- [ ] Scraping wykonuje się w kontenerze
- [ ] Pliki zapisują się do mounted volume (`./data/`)

### Persistence
- [ ] Data survives container restart
- [ ] `docker-compose down && docker-compose up -d` zachowuje dane

### Resource Usage
- [ ] Memory usage reasonable (<500MB idle)
- [ ] CPU usage reasonable (peaks during scraping only)
- [ ] Logs accessible: `docker-compose logs -f`

---

## Manual Docker Test Flow

```bash
# 1. Clean start
docker-compose down -v
rm -rf data/*

# 2. Build and start
docker-compose up --build -d

# 3. Check health
docker ps
docker inspect web-scraper | grep Health -A 5

# 4. Test via browser
open http://localhost:8080

# 5. Run scraping job
# (Use browser or curl)

# 6. Check data persistence
ls -la data/

# 7. Test restart
docker-compose restart
# Wait 10s
curl http://localhost:8080/

# 8. Check logs
docker-compose logs --tail=50 scrapper

# 9. Cleanup
docker-compose down
```

---

## Common Issues & Solutions

### Issue 1: Build fails - "cannot find package"
**Symptom**: `go: module not found`  
**Solution**: Ensure `go.mod` and `go.sum` properly committed; run `go mod tidy` locally first

### Issue 2: Permission denied on /app/data
**Symptom**: `permission denied` in logs  
**Solution**: Check Dockerfile - ensure proper chown for data directory

### Issue 3: Container exits immediately
**Symptom**: `docker ps` shows no running container  
**Solution**: Check logs `docker-compose logs scrapper`; verify binary exists and is executable

### Issue 4: Health check fails
**Symptom**: Container status "unhealthy"  
**Solution**: Test health check command manually: `docker exec web-scraper wget --spider http://localhost:8080/`

### Issue 5: Volume mount issues on Windows
**Symptom**: Data not persisting  
**Solution**: Use absolute path or Docker Desktop settings for file sharing

---

## Production Considerations (future)

### Security
- [ ] Run as non-root user ✅ (already implemented)
- [ ] Use multi-stage build ✅ (already implemented)
- [ ] Scan image for vulnerabilities: `docker scan web-scraper:latest`
- [ ] Use specific base image tags (not `latest`)

### Performance
- [ ] Configure resource limits in docker-compose
  ```yaml
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 1G
  ```

### Monitoring
- [ ] Add structured logging (JSON format)
- [ ] Integrate with logging system (ELK, Loki)
- [ ] Add metrics endpoint (Prometheus format)

### Deployment
- [ ] Use Docker secrets for sensitive config
- [ ] Implement rolling updates strategy
- [ ] Add backup strategy for data volume

---

## Next Agent

Po ukończeniu **Agent 7**, przejdź do:
👉 **AGENT_08_QA.md** (Final polish, error handling, documentation)

**Prerequisites verified**:
- ✅ Docker image builds and runs
- ✅ Container healthy and functional
- ✅ E2E test w Docker passes
- ✅ Data persistence działa

---

**Agent Status**: ⏳ TODO  
**Last Updated**: 17 lutego 2026

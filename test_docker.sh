#!/bin/bash

echo "🧪 Testing Web Scraper Docker Deployment"
echo "=========================================="

HOST_PORT=8900

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Build image
echo -e "\n${YELLOW}Step 1: Building Docker image...${NC}"
docker compose build
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# Step 2: Start container
echo -e "\n${YELLOW}Step 2: Starting container...${NC}"
docker compose up -d
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Container start failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Container started${NC}"

# Step 3: Wait for healthcheck
echo -e "\n${YELLOW}Step 3: Waiting for health check...${NC}"
sleep 10

HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' web-scraper 2>/dev/null)
echo "Health status: ${HEALTH_STATUS:-unknown}"

# Step 4: Test API endpoints
echo -e "\n${YELLOW}Step 4: Testing API endpoints...${NC}"

# Test root (should serve HTML)
echo "Testing GET /"
curl -f -s http://localhost:${HOST_PORT}/ > /dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Frontend accessible${NC}"
else
    echo -e "${RED}❌ Frontend not accessible${NC}"
fi

# Test scrape endpoint
echo "Testing POST /api/scrape"
RESPONSE=$(curl -s -X POST http://localhost:${HOST_PORT}/api/scrape \
    -H "Content-Type: application/json" \
    -d '{
        "url": "https://example.com",
        "depth": 1,
        "filters": []
    }')

PROJECT_ID=$(echo "$RESPONSE" | grep -o '"project_id":"[^"]*' | cut -d'"' -f4)

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
    STATUS_RESPONSE=$(curl -s http://localhost:${HOST_PORT}/api/project/$PROJECT_ID/status)
    echo "Status: $STATUS_RESPONSE"

    if echo "$STATUS_RESPONSE" | grep -q "completed"; then
        echo -e "${GREEN}✅ Status endpoint works${NC}"
    else
        echo -e "${YELLOW}⚠️  Scraping not completed yet${NC}"
    fi
fi

# Step 5: Check data persistence
echo -e "\n${YELLOW}Step 5: Checking data persistence...${NC}"
if [ -n "$PROJECT_ID" ] && [ -d "./data/$PROJECT_ID" ]; then
    echo -e "${GREEN}✅ Data directory created${NC}"
    echo "Files:"
    ls -lh ./data/$PROJECT_ID/
else
    echo -e "${RED}❌ Data directory not created${NC}"
fi

# Step 6: Check logs
echo -e "\n${YELLOW}Step 6: Container logs (last 20 lines):${NC}"
docker compose logs --tail=20 scrapper

# Step 7: Cleanup option
echo -e "\n${YELLOW}Test complete!${NC}"
echo "To stop container: docker compose down"
echo "To view logs: docker compose logs -f scrapper"
echo "To restart: docker compose restart"
echo "App URL: http://localhost:${HOST_PORT}"

echo -e "\n${GREEN}=========================================="
echo "✅ Docker deployment test finished"
echo "==========================================${NC}"

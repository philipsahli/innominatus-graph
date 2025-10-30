#!/bin/bash

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Innominatus Graph SDK - Setup${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check prerequisites
echo -e "${BLUE}📋 Checking prerequisites...${NC}\n"

# Check Go version
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo -e "   Install Go 1.24+ from https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.24"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo -e "${RED}❌ Go version $GO_VERSION is too old${NC}"
    echo -e "   Required: Go $REQUIRED_VERSION or higher"
    exit 1
fi

echo -e "${GREEN}✅ Go $GO_VERSION installed${NC}"

# Check GraphViz (optional but recommended)
if ! command -v dot &> /dev/null; then
    echo -e "${YELLOW}⚠️  GraphViz not installed (optional for SVG/PNG export)${NC}"
    echo -e "   Install: brew install graphviz (macOS) or apt-get install graphviz (Linux)"
else
    echo -e "${GREEN}✅ GraphViz installed${NC}"
fi

# Check PostgreSQL (optional)
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  PostgreSQL not installed (optional - SQLite works for dev)${NC}"
else
    POSTGRES_VERSION=$(psql --version | awk '{print $3}')
    echo -e "${GREEN}✅ PostgreSQL $POSTGRES_VERSION installed${NC}"
fi

echo ""

# Install dependencies
echo -e "${BLUE}📦 Installing Go dependencies...${NC}\n"
go mod download
echo -e "${GREEN}✅ Dependencies installed${NC}\n"

# Verify dependencies
echo -e "${BLUE}🔍 Verifying dependencies...${NC}\n"
go mod verify
echo -e "${GREEN}✅ Dependencies verified${NC}\n"

# Run tests
echo -e "${BLUE}🧪 Running tests...${NC}\n"
if go test ./... -short; then
    echo -e "\n${GREEN}✅ All tests passed${NC}\n"
else
    echo -e "\n${RED}❌ Tests failed${NC}"
    exit 1
fi

# Check test coverage
echo -e "${BLUE}📊 Checking test coverage...${NC}\n"
COVERAGE=$(go test -cover ./... 2>&1 | grep -E "coverage:" | awk '{sum+=$5; count++} END {print sum/count}' | sed 's/%//')
if [ ! -z "$COVERAGE" ]; then
    COVERAGE_INT=$(printf "%.0f" "$COVERAGE")
    if [ "$COVERAGE_INT" -ge 80 ]; then
        echo -e "${GREEN}✅ Test coverage: ${COVERAGE}% (target: >80%)${NC}\n"
    else
        echo -e "${YELLOW}⚠️  Test coverage: ${COVERAGE}% (target: >80%)${NC}\n"
    fi
fi

# Optional: Check SQLite connectivity
echo -e "${BLUE}💾 Testing SQLite connectivity...${NC}\n"
TEST_DB="/tmp/innominatus-graph-test.db"
rm -f "$TEST_DB"

cat > /tmp/test-sqlite.go <<'EOF'
package main
import (
    "fmt"
    "github.com/philipsahli/innominatus-graph/pkg/storage"
)
func main() {
    db, err := storage.NewSQLiteConnection("/tmp/innominatus-graph-test.db")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    storage.AutoMigrate(db)
    fmt.Println("SUCCESS")
}
EOF

if go run /tmp/test-sqlite.go 2>&1 | grep -q "SUCCESS"; then
    echo -e "${GREEN}✅ SQLite connection successful${NC}\n"
else
    echo -e "${YELLOW}⚠️  SQLite connection test failed${NC}\n"
fi
rm -f /tmp/test-sqlite.go "$TEST_DB"

# Optional: Test PostgreSQL connectivity (if configured)
if [ ! -z "$DB_PASSWORD" ]; then
    echo -e "${BLUE}💾 Testing PostgreSQL connectivity...${NC}\n"

    DB_HOST=${DB_HOST:-localhost}
    DB_USER=${DB_USER:-postgres}
    DB_NAME=${DB_NAME:-idp_orchestrator}
    DB_PORT=${DB_PORT:-5432}

    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✅ PostgreSQL connection successful${NC}\n"
    else
        echo -e "${YELLOW}⚠️  PostgreSQL connection failed${NC}"
        echo -e "   Check DB_PASSWORD and connection settings\n"
    fi
fi

# Run demo example
echo -e "${BLUE}🚀 Running demo example...${NC}\n"
cd examples/demo
if go run main.go 2>&1 | grep -q "Demo completed successfully"; then
    echo -e "\n${GREEN}✅ Demo example executed successfully${NC}"
else
    echo -e "\n${YELLOW}⚠️  Demo example had issues (check output above)${NC}"
fi
cd ../..

# Success dashboard
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✨ Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}📖 Next Steps:${NC}"
echo -e "  1. Review ${YELLOW}CLAUDE.md${NC} for architecture rules and principles"
echo -e "  2. Review ${YELLOW}DIGEST.md${NC} for quick context"
echo -e "  3. Run tests: ${YELLOW}go test ./...${NC}"
echo -e "  4. Run demo: ${YELLOW}cd examples/demo && go run main.go${NC}"
echo -e "  5. Create verification scripts in ${YELLOW}verification/${NC}\n"

echo -e "${BLUE}💡 Development Commands:${NC}"
echo -e "  • Run tests:        ${YELLOW}go test ./...${NC}"
echo -e "  • Coverage report:  ${YELLOW}go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out${NC}"
echo -e "  • Format code:      ${YELLOW}gofmt -w .${NC}"
echo -e "  • Lint code:        ${YELLOW}go vet ./...${NC}"
echo -e "  • Run demo:         ${YELLOW}cd examples/demo && go run main.go${NC}\n"

echo -e "${BLUE}🗄️  Database Options:${NC}"
echo -e "  • SQLite (dev):     ${YELLOW}Just run - no setup needed${NC}"
echo -e "  • PostgreSQL:       ${YELLOW}DB_PASSWORD=secret go run examples/demo/main.go${NC}\n"

if [ ! -z "$DB_PASSWORD" ]; then
    echo -e "${GREEN}✅ PostgreSQL configured and tested${NC}"
else
    echo -e "${YELLOW}ℹ️  Using SQLite (set DB_PASSWORD for PostgreSQL)${NC}"
fi

echo ""
echo -e "${BLUE}📚 Documentation:${NC}"
echo -e "  • CLAUDE.md:    Architecture rules (SOLID, KISS, YAGNI)"
echo -e "  • DIGEST.md:    Quick project context"
echo -e "  • README.md:    Minimal quick start"
echo -e "  • verification/: Verification-first development guide\n"

echo -e "${GREEN}Happy coding! 🎉${NC}\n"

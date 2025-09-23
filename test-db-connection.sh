#!/bin/bash

# Database Connection Test Script
# This script verifies your database setup is correct

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database credentials
DB_USER="barbershop_user"
DB_PASSWORD="secure_password_123"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="barbershop"
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}🔍 Database Connection Test${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Test 1: Check if PostgreSQL is accessible
echo -e "${YELLOW}Test 1: Checking PostgreSQL availability...${NC}"
if pg_isready -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is running and accessible${NC}\n"
else
    echo -e "${RED}✗ PostgreSQL is not accessible${NC}"
    echo -e "${YELLOW}Fix: Make sure PostgreSQL is running on ${DB_HOST}:${DB_PORT}${NC}\n"
    exit 1
fi

# Test 2: Check database connection with password
echo -e "${YELLOW}Test 2: Testing database connection...${NC}"
if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Successfully connected to database '${DB_NAME}'${NC}\n"
else
    echo -e "${RED}✗ Failed to connect to database${NC}"
    echo -e "${YELLOW}Fix: Verify credentials - User: ${DB_USER}, Database: ${DB_NAME}${NC}\n"
    exit 1
fi

# Test 3: Check PostgreSQL version
echo -e "${YELLOW}Test 3: Checking PostgreSQL version...${NC}"
PG_VERSION=$(PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT version();" | head -n1)
echo -e "${GREEN}✓ PostgreSQL version: ${PG_VERSION}${NC}\n"

# Test 4: Check if tables exist
echo -e "${YELLOW}Test 4: Checking database schema...${NC}"
TABLE_COUNT=$(PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | xargs)

if [ "$TABLE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Found ${TABLE_COUNT} tables in the database${NC}"
    
    # List tables
    echo -e "${BLUE}Tables:${NC}"
    PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "\dt" 2>/dev/null || echo "  (Use psql to view tables)"
    echo ""
else
    echo -e "${YELLOW}⚠ No tables found - you may need to run migrations${NC}"
    echo -e "${YELLOW}Run: make db-migrate${NC}\n"
fi

# Test 5: Check if barbers table exists and has data
echo -e "${YELLOW}Test 5: Checking barbers table...${NC}"
if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1 FROM barbers LIMIT 1;" > /dev/null 2>&1; then
    BARBER_COUNT=$(PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM barbers;" | xargs)
    
    if [ "$BARBER_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓ Barbers table exists with ${BARBER_COUNT} records${NC}"
        
        # Show sample data
        echo -e "${BLUE}Sample barbers:${NC}"
        PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "SELECT id, shop_name, city, status FROM barbers LIMIT 3;"
        echo ""
    else
        echo -e "${YELLOW}⚠ Barbers table exists but is empty${NC}"
        echo -e "${YELLOW}Run: make db-seed${NC}\n"
    fi
else
    echo -e "${YELLOW}⚠ Barbers table not found${NC}"
    echo -e "${YELLOW}Run: make db-migrate && make db-seed${NC}\n"
fi

# Test 6: Test connection string format
echo -e "${YELLOW}Test 6: Validating connection string...${NC}"
echo -e "${GREEN}✓ Your DATABASE_URL is:${NC}"
echo -e "${BLUE}${DB_URL}${NC}\n"

# Test 7: Check user permissions
echo -e "${YELLOW}Test 7: Checking user permissions...${NC}"
if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "CREATE TEMP TABLE test_table (id INT); DROP TABLE test_table;" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ User '${DB_USER}' has sufficient permissions (CREATE/DROP)${NC}\n"
else
    echo -e "${RED}✗ User '${DB_USER}' lacks necessary permissions${NC}"
    echo -e "${YELLOW}Fix: Grant permissions with:${NC}"
    echo -e "${BLUE}GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};${NC}\n"
fi

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Database Connection Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Host:${NC}     ${DB_HOST}:${DB_PORT}"
echo -e "${GREEN}Database:${NC} ${DB_NAME}"
echo -e "${GREEN}User:${NC}     ${DB_USER}"
echo -e "${GREEN}Tables:${NC}   ${TABLE_COUNT}"
echo -e "${GREEN}Status:${NC}   ${GREEN}Connected ✓${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Next steps
echo -e "${BLUE}📋 Next Steps:${NC}"
echo -e "1. ${GREEN}Copy your .env file:${NC}     cp .env.example .env"
echo -e "2. ${GREEN}Update DATABASE_URL in .env${NC}"
echo -e "3. ${GREEN}Run migrations:${NC}         make db-migrate"
echo -e "4. ${GREEN}Seed database:${NC}          make db-seed"
echo -e "5. ${GREEN}Start server:${NC}           make run"
echo -e "6. ${GREEN}Test API:${NC}               curl http://localhost:8080/health\n"

echo -e "${GREEN}🎉 Database setup verified successfully!${NC}\n"
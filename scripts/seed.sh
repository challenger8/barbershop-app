#!/bin/bash

# Barbershop Database Seeding Script with Table Creation
# Usage: ./scripts/seed.sh [options]

set -e  # Exit on any error

# Default values
DATABASE_URL=${DATABASE_URL:-"postgres://barbershop_user:secure_password_123@localhost:5432/barbershop?sslmode=disable"}
FORCE_SEED=false
VERBOSE=false
SEEDS_PATH="./scripts/seeds"
MIGRATIONS_PATH="./pkg/database/migrations"
ENV_FILE="../.env"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to show help
show_help() {
    cat << EOF
Barbershop Database Seeding Script

USAGE:
    ./scripts/seed.sh [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -f, --force             Force re-seeding (drops existing data)
    -v, --verbose           Verbose output
    -d, --database URL      Database connection URL
    -p, --path PATH         Path to seed files directory
    -m, --migrations PATH   Path to migration files directory
    -e, --env FILE          Environment file path
    --sql-only              Run only SQL seeds (skip Go seed runner)
    --test-data             Generate additional test data
    --cleanup               Clean up test data
    --backup                Create backup before seeding
    --restore FILE          Restore from backup file
    --skip-migrations       Skip table creation (migrations)

EXAMPLES:
    # Basic seeding (creates tables + seeds data)
    ./scripts/seed.sh

    # Force re-seed with verbose output
    ./scripts/seed.sh --force --verbose

    # Seed with custom database URL
    ./scripts/seed.sh -d "postgres://user:pass@localhost/mydb"

    # Generate additional test data
    ./scripts/seed.sh --test-data

    # Create backup before seeding
    ./scripts/seed.sh --backup --force

ENVIRONMENT VARIABLES:
    DATABASE_URL            Database connection string
    BARBERSHOP_ENV          Environment (development, staging, production)
    SEED_BACKUP_PATH        Path for database backups

EOF
}

# Parse command line arguments
SKIP_MIGRATIONS=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -f|--force)
            FORCE_SEED=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -d|--database)
            DATABASE_URL="$2"
            shift 2
            ;;
        -p|--path)
            SEEDS_PATH="$2"
            shift 2
            ;;
        -m|--migrations)
            MIGRATIONS_PATH="$2"
            shift 2
            ;;
        -e|--env)
            ENV_FILE="$2"
            shift 2
            ;;
        --sql-only)
            SQL_ONLY=true
            shift
            ;;
        --test-data)
            TEST_DATA=true
            shift
            ;;
        --cleanup)
            CLEANUP=true
            shift
            ;;
        --backup)
            CREATE_BACKUP=true
            shift
            ;;
        --restore)
            RESTORE_FILE="$2"
            shift 2
            ;;
        --skip-migrations)
            SKIP_MIGRATIONS=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Load environment file if it exists
if [[ -f "$ENV_FILE" ]]; then
    print_status "Loading environment from $ENV_FILE"
    set -a  # automatically export all variables
    source "$ENV_FILE"
    set +a
else
    print_warning "Environment file $ENV_FILE not found, using defaults"
fi

# Validate database connection
validate_database() {
    print_status "Validating database connection..."
    
    if ! command -v psql &> /dev/null; then
        print_error "psql is not installed. Please install PostgreSQL client tools."
        exit 1
    fi

    # Test connection
    if ! psql "$DATABASE_URL" -c "SELECT 1;" &> /dev/null; then
        print_error "Cannot connect to database. Please check your DATABASE_URL."
        print_error "Current URL: $DATABASE_URL"
        exit 1
    fi

    print_success "Database connection validated"
}

# Create database tables (migrations)
create_tables() {
    if [[ "$SKIP_MIGRATIONS" == "true" ]]; then
        print_status "Skipping table creation (--skip-migrations flag)"
        return 0
    fi

    print_status "Creating database tables..."
    
    # Check if migrations directory exists
    if [[ -d "$MIGRATIONS_PATH" ]]; then
        print_status "Running migrations from $MIGRATIONS_PATH"
        
        # Find all SQL migration files and sort them
        MIGRATION_FILES=($(find "$MIGRATIONS_PATH" -name "*.sql" | sort))
        
        if [[ ${#MIGRATION_FILES[@]} -gt 0 ]]; then
            for migration_file in "${MIGRATION_FILES[@]}"; do
                filename=$(basename "$migration_file")
                print_status "  Running migration: $filename"
                
                if [[ "$VERBOSE" == "true" ]]; then
                    psql "$DATABASE_URL" -f "$migration_file"
                else
                    psql "$DATABASE_URL" -f "$migration_file" > /dev/null 2>&1
                fi
                
                if [[ $? -eq 0 ]]; then
                    print_success "  ✓ $filename completed"
                else
                    print_error "  ✗ $filename failed"
                    exit 1
                fi
            done
        else
            print_warning "No migration files found in $MIGRATIONS_PATH"
            create_tables_inline
        fi
    else
        print_warning "Migrations directory not found: $MIGRATIONS_PATH"
        print_status "Creating tables with inline SQL..."
        create_tables_inline
    fi
    
    print_success "Database tables created successfully"
}

# Create tables inline (if no migration files exist)
create_tables_inline() {
    print_status "Creating tables from inline SQL definitions..."
    
    # Create users table
    psql "$DATABASE_URL" << 'EOSQL'
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    user_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    email_verified BOOLEAN DEFAULT false,
    phone_verified BOOLEAN DEFAULT false,
    two_factor_enabled BOOLEAN DEFAULT false,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    date_of_birth TIMESTAMP,
    gender VARCHAR(50),
    profile_picture_url TEXT,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    preferences JSONB,
    notification_settings JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    created_by INTEGER,
    deleted_at TIMESTAMP
);
EOSQL
    print_success "  ✓ users table created"

    # Create barbers table
    psql "$DATABASE_URL" << 'EOSQL'
CREATE TABLE IF NOT EXISTS barbers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    uuid VARCHAR(255) UNIQUE NOT NULL,
    shop_name VARCHAR(255) NOT NULL,
    business_name VARCHAR(255),
    business_registration_number VARCHAR(255),
    tax_id VARCHAR(255),
    address TEXT NOT NULL,
    address_line_2 TEXT,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    phone VARCHAR(50),
    business_email VARCHAR(255),
    website_url VARCHAR(255),
    description TEXT,
    years_experience INTEGER,
    specialties JSONB,
    certifications JSONB,
    languages_spoken JSONB,
    profile_image_url TEXT,
    cover_image_url TEXT,
    gallery_images JSONB,
    working_hours JSONB,
    rating DECIMAL(3,2) DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    total_bookings INTEGER DEFAULT 0,
    response_time_minutes INTEGER DEFAULT 0,
    acceptance_rate DECIMAL(5,2) DEFAULT 0,
    cancellation_rate DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    is_verified BOOLEAN DEFAULT false,
    verification_date TIMESTAMP,
    verification_notes TEXT,
    advance_booking_days INTEGER DEFAULT 30,
    min_booking_notice_hours INTEGER DEFAULT 2,
    auto_accept_bookings BOOLEAN DEFAULT false,
    instant_booking_enabled BOOLEAN DEFAULT false,
    commission_rate DECIMAL(5,2) DEFAULT 15.00,
    payout_method VARCHAR(100),
    payout_details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_barbers_user_id ON barbers(user_id);
CREATE INDEX IF NOT EXISTS idx_barbers_city ON barbers(city);
CREATE INDEX IF NOT EXISTS idx_barbers_state ON barbers(state);
CREATE INDEX IF NOT EXISTS idx_barbers_status ON barbers(status);
CREATE INDEX IF NOT EXISTS idx_barbers_rating ON barbers(rating DESC);
CREATE INDEX IF NOT EXISTS idx_barbers_deleted_at ON barbers(deleted_at);
EOSQL
    print_success "  ✓ barbers table created with indexes"

    # Create additional tables if they don't exist
    psql "$DATABASE_URL" << 'EOSQL'
-- Service Categories
CREATE TABLE IF NOT EXISTS service_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES service_categories(id),
    level INTEGER DEFAULT 1,
    category_path VARCHAR(500),
    icon_url TEXT,
    color_hex VARCHAR(7),
    image_url TEXT,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    meta_title VARCHAR(255),
    meta_description TEXT,
    keywords JSONB,
    service_count INTEGER DEFAULT 0,
    barber_count INTEGER DEFAULT 0,
    average_price DECIMAL(10,2) DEFAULT 0,
    popularity_score DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Services
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    short_description TEXT,
    detailed_description TEXT,
    category_id INTEGER REFERENCES service_categories(id),
    service_type VARCHAR(100),
    complexity INTEGER DEFAULT 1,
    skill_level_required VARCHAR(50),
    default_duration_min INTEGER,
    default_duration_max INTEGER,
    suggested_price_min DECIMAL(10,2),
    suggested_price_max DECIMAL(10,2),
    currency VARCHAR(10) DEFAULT 'USD',
    target_gender VARCHAR(50),
    target_age_min INTEGER,
    target_age_max INTEGER,
    hair_types JSONB,
    requires_consultation BOOLEAN DEFAULT false,
    required_tools JSONB,
    required_products JSONB,
    required_certifications JSONB,
    allergen_warnings JSONB,
    health_precautions JSONB,
    requires_health_check BOOLEAN DEFAULT false,
    image_url TEXT,
    gallery_images JSONB,
    video_url TEXT,
    tags JSONB,
    search_keywords JSONB,
    meta_description TEXT,
    has_variations BOOLEAN DEFAULT false,
    allows_add_ons BOOLEAN DEFAULT false,
    global_popularity_score DECIMAL(5,2) DEFAULT 0,
    total_global_bookings INTEGER DEFAULT 0,
    average_global_rating DECIMAL(3,2) DEFAULT 0,
    total_global_reviews INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    is_approved BOOLEAN DEFAULT true,
    approval_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    last_modified_by INTEGER,
    version INTEGER DEFAULT 1,
    change_log JSONB
);

-- Barber Services
CREATE TABLE IF NOT EXISTS barber_services (
    id SERIAL PRIMARY KEY,
    barber_id INTEGER REFERENCES barbers(id) ON DELETE CASCADE,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    custom_name VARCHAR(255),
    custom_description TEXT,
    price DECIMAL(10,2) NOT NULL,
    max_price DECIMAL(10,2),
    currency VARCHAR(10) DEFAULT 'USD',
    discount_price DECIMAL(10,2),
    discount_valid_until TIMESTAMP,
    estimated_duration_min INTEGER NOT NULL,
    estimated_duration_max INTEGER,
    buffer_time_minutes INTEGER DEFAULT 0,
    advance_notice_hours INTEGER DEFAULT 2,
    max_advance_booking_days INTEGER,
    available_days JSONB,
    available_time_slots JSONB,
    requires_consultation BOOLEAN,
    consultation_duration INTEGER,
    pre_service_instructions TEXT,
    post_service_care TEXT,
    min_customer_age INTEGER,
    max_customer_age INTEGER,
    is_seasonal BOOLEAN DEFAULT false,
    seasonal_start_month INTEGER,
    seasonal_end_month INTEGER,
    portfolio_images JSONB,
    before_after_images JSONB,
    total_bookings INTEGER DEFAULT 0,
    total_revenue DECIMAL(12,2) DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    cancellation_rate DECIMAL(5,2) DEFAULT 0,
    customer_satisfaction DECIMAL(5,2) DEFAULT 0,
    repeat_customer_rate DECIMAL(5,2) DEFAULT 0,
    bookings_last_30_days INTEGER DEFAULT 0,
    revenue_last_30_days DECIMAL(10,2) DEFAULT 0,
    popularity_score DECIMAL(5,2) DEFAULT 0,
    demand_level DECIMAL(3,2) DEFAULT 0,
    is_promotional BOOLEAN DEFAULT false,
    promotional_text TEXT,
    promotion_start_date TIMESTAMP,
    promotion_end_date TIMESTAMP,
    is_featured BOOLEAN DEFAULT false,
    display_order INTEGER DEFAULT 0,
    service_note TEXT,
    is_active BOOLEAN DEFAULT true,
    paused_reason TEXT,
    paused_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Time Slots
CREATE TABLE IF NOT EXISTS time_slots (
    id SERIAL PRIMARY KEY,
    barber_id INTEGER REFERENCES barbers(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration_minutes INTEGER NOT NULL,
    is_available BOOLEAN DEFAULT true,
    slot_type VARCHAR(50) DEFAULT 'regular',
    base_price DECIMAL(10,2),
    dynamic_price DECIMAL(10,2),
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    service_id INTEGER REFERENCES barber_services(id),
    max_customers INTEGER DEFAULT 1,
    min_advance_notice_hours INTEGER DEFAULT 2,
    notes TEXT,
    special_requirements JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER
);

-- Bookings
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    booking_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id INTEGER REFERENCES users(id),
    barber_id INTEGER REFERENCES barbers(id) ON DELETE CASCADE,
    time_slot_id INTEGER REFERENCES time_slots(id),
    service_name VARCHAR(255) NOT NULL,
    service_category VARCHAR(100),
    estimated_duration_minutes INTEGER NOT NULL,
    customer_name VARCHAR(255),
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    service_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    tip_amount DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'USD',
    payment_status VARCHAR(50) NOT NULL,
    payment_method VARCHAR(100),
    payment_reference VARCHAR(255),
    paid_at TIMESTAMP,
    notes TEXT,
    special_requests TEXT,
    internal_notes TEXT,
    confirmation_method VARCHAR(50),
    confirmation_sent_at TIMESTAMP,
    reminder_sent_at TIMESTAMP,
    scheduled_start_time TIMESTAMP NOT NULL,
    scheduled_end_time TIMESTAMP NOT NULL,
    actual_start_time TIMESTAMP,
    actual_end_time TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancelled_by INTEGER,
    cancellation_reason TEXT,
    cancellation_fee DECIMAL(10,2) DEFAULT 0,
    booking_source VARCHAR(100),
    referral_source VARCHAR(255),
    utm_campaign VARCHAR(255),
    ml_prediction_score DECIMAL(5,4),
    customer_segment VARCHAR(100),
    booking_value_score DECIMAL(5,2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Reviews
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE CASCADE,
    customer_id INTEGER REFERENCES users(id),
    barber_id INTEGER REFERENCES barbers(id) ON DELETE CASCADE,
    overall_rating INTEGER NOT NULL CHECK (overall_rating >= 1 AND overall_rating <= 5),
    service_quality_rating INTEGER CHECK (service_quality_rating >= 1 AND service_quality_rating <= 5),
    punctuality_rating INTEGER CHECK (punctuality_rating >= 1 AND punctuality_rating <= 5),
    cleanliness_rating INTEGER CHECK (cleanliness_rating >= 1 AND cleanliness_rating <= 5),
    value_for_money_rating INTEGER CHECK (value_for_money_rating >= 1 AND value_for_money_rating <= 5),
    professionalism_rating INTEGER CHECK (professionalism_rating >= 1 AND professionalism_rating <= 5),
    title VARCHAR(255),
    comment TEXT,
    pros TEXT,
    cons TEXT,
    would_recommend BOOLEAN,
    would_book_again BOOLEAN,
    service_as_expected BOOLEAN,
    duration_accurate BOOLEAN,
    images JSONB,
    is_verified BOOLEAN DEFAULT false,
    is_published BOOLEAN DEFAULT false,
    moderation_status VARCHAR(50) DEFAULT 'pending',
    moderation_notes TEXT,
    moderated_by INTEGER,
    moderated_at TIMESTAMP,
    helpful_votes INTEGER DEFAULT 0,
    total_votes INTEGER DEFAULT 0,
    barber_response TEXT,
    barber_response_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(100) NOT NULL,
    channels JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    related_entity_type VARCHAR(100),
    related_entity_id INTEGER,
    data JSONB,
    priority VARCHAR(50) DEFAULT 'normal',
    scheduled_for TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Barber Availability
CREATE TABLE IF NOT EXISTS barber_availability (
    id SERIAL PRIMARY KEY,
    barber_id INTEGER REFERENCES barbers(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    day_of_week INTEGER NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    availability_type VARCHAR(50) NOT NULL,
    is_recurring BOOLEAN DEFAULT false,
    recurring_pattern VARCHAR(50),
    recurring_end_date DATE,
    notes TEXT,
    blocked_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create additional indexes
CREATE INDEX IF NOT EXISTS idx_bookings_customer_id ON bookings(customer_id);
CREATE INDEX IF NOT EXISTS idx_bookings_barber_id ON bookings(barber_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_scheduled_start_time ON bookings(scheduled_start_time);
CREATE INDEX IF NOT EXISTS idx_reviews_barber_id ON reviews(barber_id);
CREATE INDEX IF NOT EXISTS idx_reviews_customer_id ON reviews(customer_id);
CREATE INDEX IF NOT EXISTS idx_reviews_is_published ON reviews(is_published);
CREATE INDEX IF NOT EXISTS idx_barber_services_barber_id ON barber_services(barber_id);
CREATE INDEX IF NOT EXISTS idx_barber_services_is_active ON barber_services(is_active);
CREATE INDEX IF NOT EXISTS idx_time_slots_barber_id ON time_slots(barber_id);
CREATE INDEX IF NOT EXISTS idx_time_slots_start_time ON time_slots(start_time);
CREATE INDEX IF NOT EXISTS idx_time_slots_is_available ON time_slots(is_available);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);

EOSQL
    print_success "  ✓ All additional tables and indexes created"
}

# Create backup
create_backup() {
    if [[ "$CREATE_BACKUP" == "true" ]]; then
        print_status "Creating database backup..."
        
        BACKUP_DIR="${SEED_BACKUP_PATH:-./backups}"
        mkdir -p "$BACKUP_DIR"
        
        BACKUP_FILE="$BACKUP_DIR/barbershop_backup_$(date +%Y%m%d_%H%M%S).sql"
        
        if pg_dump "$DATABASE_URL" > "$BACKUP_FILE"; then
            print_success "Backup created: $BACKUP_FILE"
        else
            print_error "Failed to create backup"
            exit 1
        fi
    fi
}

# Restore from backup
restore_backup() {
    if [[ -n "$RESTORE_FILE" ]]; then
        print_status "Restoring from backup: $RESTORE_FILE"
        
        if [[ ! -f "$RESTORE_FILE" ]]; then
            print_error "Backup file not found: $RESTORE_FILE"
            exit 1
        fi

        # Drop and recreate database
        DB_NAME=$(echo "$DATABASE_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')
        BASE_URL=$(echo "$DATABASE_URL" | sed 's/\/[^\/]*$/\/postgres/')
        
        print_status "Dropping database $DB_NAME..."
        psql "$BASE_URL" -c "DROP DATABASE IF EXISTS $DB_NAME;"
        
        print_status "Creating database $DB_NAME..."
        psql "$BASE_URL" -c "CREATE DATABASE $DB_NAME;"
        
        print_status "Restoring data..."
        if psql "$DATABASE_URL" < "$RESTORE_FILE"; then
            print_success "Database restored successfully"
            exit 0
        else
            print_error "Failed to restore database"
            exit 1
        fi
    fi
}

# Run SQL seeds directly
run_sql_seeds() {
    print_status "Running SQL seed files..."
    
    if [[ ! -d "$SEEDS_PATH" ]]; then
        print_error "Seeds directory not found: $SEEDS_PATH"
        exit 1
    fi

    # Find all SQL files and sort them
    SQL_FILES=($(find "$SEEDS_PATH" -name "*.sql" | sort))
    
    if [[ ${#SQL_FILES[@]} -eq 0 ]]; then
        print_warning "No SQL files found in $SEEDS_PATH"
        return 0
    fi

    for sql_file in "${SQL_FILES[@]}"; do
        filename=$(basename "$sql_file")
        print_status "Executing $filename..."
        
        if [[ "$VERBOSE" == "true" ]]; then
            psql "$DATABASE_URL" -f "$sql_file"
        else
            psql "$DATABASE_URL" -f "$sql_file" > /dev/null 2>&1
        fi
        
        if [[ $? -eq 0 ]]; then
            print_success "✓ $filename completed"
        else
            print_error "✗ $filename failed"
            exit 1
        fi
    done
}

# Generate test data
generate_test_data() {
    if [[ "$TEST_DATA" == "true" ]]; then
        print_status "Generating additional test data..."
        
        # Generate more users, bookings, etc.
        psql "$DATABASE_URL" << EOF
-- Generate additional customers
INSERT INTO users (uuid, email, password_hash, name, phone, user_type, status, email_verified, created_at, updated_at)
SELECT 
    gen_random_uuid()::text,
    'customer' || generate_series || '@test.example.com',
    '\$2a\$12\$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeENZW3D7gVrK5ZK.',
    'Test Customer ' || generate_series,
    '+1-555-' || LPAD(generate_series::text, 4, '0'),
    'customer',
    'active',
    true,
    NOW() - (random() * interval '30 days'),
    NOW()
FROM generate_series(1, 10);

-- Generate additional bookings (only if barbers and time_slots exist)
INSERT INTO bookings (uuid, booking_number, customer_id, barber_id, time_slot_id, service_name, estimated_duration_minutes, status, service_price, total_price, currency, payment_status, scheduled_start_time, scheduled_end_time, created_at, updated_at)
SELECT 
    gen_random_uuid()::text,
    'TEST' || LPAD(generate_series::text, 6, '0'),
    (SELECT id FROM users WHERE user_type = 'customer' ORDER BY random() LIMIT 1),
    (SELECT id FROM barbers ORDER BY random() LIMIT 1),
    1,
    'Test Service',
    30,
    CASE WHEN random() > 0.7 THEN 'completed' ELSE 'confirmed' END,
    25.00 + (random() * 50),
    30.00 + (random() * 60),
    'USD',
    'paid',
    NOW() + (random() * interval '30 days'),
    NOW() + (random() * interval '30 days') + interval '30 minutes',
    NOW() - (random() * interval '15 days'),
    NOW()
FROM generate_series(1, 20)
WHERE EXISTS (SELECT 1 FROM barbers LIMIT 1);

SELECT 'Test data generated successfully' as result;
EOF
        
        print_success "Test data generated"
    fi
}

# Cleanup test data
cleanup_test_data() {
    if [[ "$CLEANUP" == "true" ]]; then
        print_status "Cleaning up test data..."
        
        psql "$DATABASE_URL" << EOF
-- Remove test users and related data
DELETE FROM bookings WHERE booking_number LIKE 'TEST%';
DELETE FROM users WHERE email LIKE '%test.example.com';

SELECT 'Test data cleaned up successfully' as result;
EOF
        
        print_success "Test data cleaned up"
    fi
}

# Main execution function
main() {
    print_status "🌱 Barbershop Database Setup & Seeding"
    print_status "======================================="
    
    # Handle restore operation first
    restore_backup
    
    # Validate environment
    validate_database
    
    # Create backup if requested
    create_backup
    
    # Handle cleanup operation
    cleanup_test_data
    
    # Step 1: Create tables (migrations)
    print_status ""
    print_status "Step 1: Creating database schema..."
    create_tables
    
    # Check if seeding is needed (unless forced)
    if [[ "$FORCE_SEED" != "true" ]]; then
        NEEDS_SEEDING=$(psql "$DATABASE_URL" -t -c "
            SELECT NOT EXISTS (
                SELECT 1 FROM information_schema.tables 
                WHERE table_name = 'users' 
                AND EXISTS (SELECT 1 FROM users LIMIT 1)
            );
        " 2>/dev/null | xargs)
        
        if [[ "$NEEDS_SEEDING" == "f" ]]; then
            print_success "Database already contains data. Use --force to re-seed."
            exit 0
        fi
    fi
    
    # Step 2: Run seeding
    print_status ""
    print_status "Step 2: Seeding database with data..."
    run_sql_seeds
    
    # Generate additional test data if requested
    generate_test_data
    
    # Display summary
    print_status ""
    print_status "📊 Database Summary"
    print_status "=================="
    
    psql "$DATABASE_URL" -c "
        SELECT 
            'Users' as entity, 
            COUNT(*)::text as count 
        FROM users
        UNION ALL
        SELECT 
            'Active Barbers' as entity, 
            COUNT(*)::text as count 
        FROM barbers WHERE status = 'active'
        UNION ALL
        SELECT 
            'Services' as entity, 
            COUNT(*)::text as count 
        FROM services WHERE is_active = true
        UNION ALL
        SELECT 
            'Total Bookings' as entity, 
            COUNT(*)::text as count 
        FROM bookings
        UNION ALL
        SELECT 
            'Published Reviews' as entity, 
            COUNT(*)::text as count 
        FROM reviews WHERE is_published = true;
    "
    
    print_success ""
    print_success "🎉 Database setup completed successfully!"
    print_status ""
    print_status "🔑 Test Credentials:"
    print_status "  Admin:    admin@barbershop.com / password123"
    print_status "  Customer: john.doe@email.com / password123"
    print_status "  Barber:   tony.soprano@barbershop.com / password123"
    print_status ""
    print_status "🌐 API Server:"
    print_status "  Start with: go run cmd/server/main.go cmd/server/routes.go"
    print_status "  Visit: http://localhost:8080/health"
    print_status ""
    print_status "📝 Next Steps:"
    print_status "  1. Start the API server"
    print_status "  2. Test endpoints: curl http://localhost:8080/api/v1/barbers"
    print_status "  3. Check the API documentation in API_TESTING.md"
    print_status ""
}

# Error handling
trap 'print_error "Script interrupted"; exit 1' INT TERM

# Run main function
main "$@"
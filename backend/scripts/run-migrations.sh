#!/bin/bash
# Run database migrations for RAG Dashboard

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Please set DATABASE_URL in .env file or as environment variable"
    exit 1
fi

# Parse DATABASE_URL to get connection details
# Format: postgresql://user:pass@host:port/dbname?params
DB_URL_REGEX="postgresql://([^:]+):([^@]+)@([^:]+):([^/]+)/([^?]+)"
if [[ $DATABASE_URL =~ $DB_URL_REGEX ]]; then
    DB_USER="${BASH_REMATCH[1]}"
    DB_PASS="${BASH_REMATCH[2]}"
    DB_HOST="${BASH_REMATCH[3]}"
    DB_PORT="${BASH_REMATCH[4]}"
    DB_NAME="${BASH_REMATCH[5]}"
else
    echo "Error: Could not parse DATABASE_URL"
    exit 1
fi

MIGRATIONS_DIR="./migrations"

echo "Running migrations..."
echo "Database: $DB_NAME on $DB_HOST:$DB_PORT"

# Set PGPASSWORD for psql
export PGPASSWORD="$DB_PASS"

# Run each migration file
for migration in $(ls $MIGRATIONS_DIR/*.sql | grep -v '.down.sql' | sort); do
    echo "Applying migration: $(basename $migration)"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration"
    if [ $? -eq 0 ]; then
        echo "✓ Migration applied successfully"
    else
        echo "✗ Migration failed"
        exit 1
    fi
done

echo "All migrations completed successfully!"

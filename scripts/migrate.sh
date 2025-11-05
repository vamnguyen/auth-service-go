#!/bin/bash
# Database migration script

set -e

echo "🗄️  Running database migrations..."

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "❌ ERROR: DATABASE_URL is not set"
    echo "Please set DATABASE_URL environment variable or add it to .env file"
    exit 1
fi

echo "✅ DATABASE_URL is set"

# Run migrations via Go
go run ./cmd/server/main.go migrate 2>&1 | grep -q "migrated successfully" && {
    echo "✅ Migrations completed successfully"
} || {
    echo "⚠️  Check if database is accessible"
}

echo "Done!"

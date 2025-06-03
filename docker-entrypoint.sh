#!/bin/sh
set -e

# 1. Apply Drizzle migrations (runs npm run db:push)
echo "⏳ Running migrations..."
npm run db:push --force

# 2. Then exec whatever CMD was given (i.e. “node build/index.js”)
echo "✅ Migrations complete. Starting server."
exec "$@"


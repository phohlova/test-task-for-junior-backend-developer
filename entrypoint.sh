set -e

echo "Waiting for PostgreSQL..."
until PGPASSWORD=postgres psql -h postgres -U postgres -d taskservice -c "SELECT 1" > /dev/null 2>&1; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready!"

echo "Applying migration 001..."
PGPASSWORD=postgres psql -h postgres -U postgres -d taskservice -f /app/migrations/0001_create_tasks.up.sql

echo "Applying migration 002..."
PGPASSWORD=postgres psql -h postgres -U postgres -d taskservice -f /app/migrations/0002_add_recurrence_tasks.up.sql

echo "Migrations applied successfully!"
echo "Starting application..."
exec /app/taskservice
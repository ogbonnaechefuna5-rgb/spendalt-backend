# Quick Start Guide

Get Spendalt backend running in 5 minutes.

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Docker (optional, for easy setup)

## Option 1: Docker (Recommended)

```bash
cd backend

# Start PostgreSQL and Redis
docker-compose up -d

# Wait 5 seconds for DB to initialize
sleep 5

# Copy environment file
cp .env.example .env

# Run migrations
make migrate

# Start server
make run
```

Server running at `http://localhost:8080` 🚀

## Option 2: Manual Setup

### 1. Install PostgreSQL

**macOS:**
```bash
brew install postgresql@15
brew services start postgresql@15
```

**Ubuntu:**
```bash
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

### 2. Create Database

```bash
psql -U postgres
CREATE DATABASE spendalt;
CREATE USER spendalt WITH PASSWORD 'spendalt123';
GRANT ALL PRIVILEGES ON DATABASE spendalt TO spendalt;
\q
```

### 3. Configure Environment

```bash
cd backend
cp .env.example .env
```

Edit `.env`:
```
PORT=8080
DATABASE_URL=postgres://spendalt:spendalt123@localhost:5432/spendalt?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key-change-in-production
ENVIRONMENT=development
```

### 4. Run Migrations

```bash
psql -U spendalt -d spendalt -f migrations/001_init.sql
```

### 5. Install Dependencies

```bash
go mod download
```

### 6. Start Server

```bash
go run cmd/api/main.go
```

## Test the API

### 1. Create Account

```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+2348012345678",
    "password": "test123",
    "full_name": "Test User"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "uuid-here"
}
```

### 2. Add Transaction

```bash
# Replace YOUR_TOKEN with token from signup
curl -X POST http://localhost:8080/api/v1/transactions/ingest/manual \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000,
    "transaction_type": "debit",
    "description": "Lunch at KFC",
    "category": "Food",
    "transaction_date": "2024-01-15T12:30:00Z"
  }'
```

### 3. Get Transactions

```bash
curl http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. Create Budget

```bash
curl -X POST http://localhost:8080/api/v1/budgets \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Food",
    "amount": 50000,
    "period": "monthly",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31"
  }'
```

### 5. Get Financial Health Score

```bash
curl http://localhost:8080/api/v1/health/score \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Using Postman

1. Import `Spendalt.postman_collection.json`
2. Run "Signup" request
3. Copy token from response
4. Set `{{token}}` variable in Postman environment
5. Test all endpoints

## Troubleshooting

### Database Connection Failed

**Error:** `connection refused`

**Fix:**
```bash
# Check if PostgreSQL is running
pg_isready

# Start PostgreSQL
brew services start postgresql@15  # macOS
sudo systemctl start postgresql    # Linux
```

### Port Already in Use

**Error:** `bind: address already in use`

**Fix:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in .env
PORT=8081
```

### Migration Failed

**Error:** `relation already exists`

**Fix:**
```bash
# Drop and recreate database
psql -U postgres
DROP DATABASE spendalt;
CREATE DATABASE spendalt;
\q

# Run migrations again
make migrate
```

## Development Workflow

### Hot Reload (Optional)

Install Air for hot reload:
```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml`:
```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/api"
  bin = "tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor"]
```

Run with hot reload:
```bash
air
```

### Run Tests

```bash
make test
```

### Build Binary

```bash
make build
./bin/api
```

## Next Steps

1. ✅ Backend running
2. 📱 Connect React Native app
3. 🧪 Test SMS ingestion (Android only)
4. 📊 View analytics in app
5. 🚀 Deploy to production

## Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for:
- Railway deployment
- Render deployment
- AWS deployment
- Environment configuration
- SSL setup
- Monitoring

## Need Help?

- 📖 [API Documentation](API.md)
- 🏗️ [Architecture Guide](ARCHITECTURE.md)
- 💻 [Implementation Details](IMPLEMENTATION.md)
- 🐛 [GitHub Issues](https://github.com/spendalt/backend/issues)

---

**Happy coding! 🎉**

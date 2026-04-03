# Spendalt Backend - Go Fiber

Production-ready fintech backend for Nigerian personal finance analytics.

## Architecture Overview

```
┌─────────────────┐
│  Mobile App     │
│ (React Native)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  API Gateway    │
│   (Go Fiber)    │
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    ▼         ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│ Auth   │ │Ingest  │ │Process │ │Analytics│
│Service │ │Service │ │Service │ │Service  │
└────────┘ └────────┘ └────────┘ └────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────┐
│Postgres│ │ Redis  │
└────────┘ └────────┘
```

## Key Design Decisions

### 1. DATA INGESTION STRATEGY

**Hybrid approach to minimize API costs:**

- **SMS Parsing (Primary)**: Android users send SMS to backend
- **Manual Entry (Fallback)**: iOS users or when SMS fails
- **Statement Upload**: Batch processing (future)
- **API Sync**: Only on explicit user request (not automatic)

**Deduplication via fingerprinting:**
```
fingerprint = SHA256(user_id + amount + date + merchant)
```

### 2. DATABASE DESIGN

**Two-tier storage:**

1. **raw_transactions**: Unprocessed data from all sources
2. **transactions**: Cleaned, categorized, deduplicated

**Why?**
- Allows reprocessing if categorization improves
- Audit trail for debugging
- Can train ML models later

### 3. CATEGORIZATION ENGINE

**Rule-based (MVP):**
- Keyword matching against category database
- Merchant normalization (remove special chars, lowercase)
- Fallback to "Other" category

**Future: ML-based:**
- Train on user corrections
- Collaborative filtering across users
- Merchant database grows organically

### 4. COST OPTIMIZATION

**API call minimization:**
- No automatic syncing
- User triggers sync manually
- Cache merchant data (Redis)
- Batch operations where possible

**Estimated costs at scale:**
- 10k users × 50 SMS/month = 500k transactions
- 0 API calls (SMS-based)
- Only storage + compute costs

### 5. SCALABILITY PLAN

**0 → 10k users:**
- Single server (2 vCPU, 4GB RAM)
- PostgreSQL on same server
- Cost: ~$20/month

**10k → 100k users:**
- Separate DB server
- Add Redis for caching
- Horizontal scaling (2-3 API servers)
- Cost: ~$100/month

**100k → 1M users:**
- Database read replicas
- Message queue (RabbitMQ/Kafka)
- Microservices split
- CDN for static assets
- Cost: ~$500-1000/month

### 6. SECURITY

**Data protection:**
- Passwords: bcrypt (cost 14)
- JWT tokens: 30-day expiry
- HTTPS only in production
- SQL injection prevention (parameterized queries)

**Compliance (Nigeria):**
- NDPR (Nigeria Data Protection Regulation)
- No PII in logs
- User data deletion on request
- Encrypted backups

## Setup

### Prerequisites
```bash
go 1.21+
PostgreSQL 14+
Redis 7+ (optional for MVP)
```

### Installation
```bash
cd backend
cp .env.example .env
# Edit .env with your database credentials

# Install dependencies
go mod download

# Run migrations
psql -U postgres -d spendalt -f migrations/001_init.sql

# Run server
go run cmd/api/main.go
```

### Environment Variables
```
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/spendalt?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key-change-in-production
ENVIRONMENT=development
```

## API Endpoints

### Authentication
```
POST /api/v1/auth/signup
POST /api/v1/auth/login
```

### User Profile
```
GET  /api/v1/user/profile
PUT  /api/v1/user/profile
POST /api/v1/user/change-password
DEL  /api/v1/user/account
```

### Transactions
```
POST /api/v1/transactions/ingest/sms      # SMS ingestion
POST /api/v1/transactions/ingest/manual   # Manual entry
GET  /api/v1/transactions                 # List transactions
```

### Categories
```
GET /api/v1/categories                    # All categories
GET /api/v1/categories/breakdown          # Spending by category
```

### Budgets
```
POST /api/v1/budgets                      # Create budget
GET  /api/v1/budgets                      # List budgets with spending
PUT  /api/v1/budgets/:id                  # Update budget
DEL  /api/v1/budgets/:id                  # Delete budget
```

### Savings Goals
```
POST /api/v1/savings                      # Create goal
GET  /api/v1/savings                      # List goals
PUT  /api/v1/savings/:id/progress         # Update progress
DEL  /api/v1/savings/:id                  # Delete goal
```

### Analytics
```
GET /api/v1/analytics/insights            # Monthly summary
GET /api/v1/analytics/weekly-trend        # 7-day spending
```

### Financial Health
```
GET /api/v1/health/score                  # Health score & insights
```

See [API.md](API.md) for detailed documentation.

## Data Flow

### SMS Ingestion Flow
```
1. Mobile app receives SMS
2. App sends raw SMS to /ingest/sms
3. Backend extracts amount, type, merchant
4. Saves to raw_transactions
5. Async processing:
   - Categorize using keywords
   - Generate fingerprint
   - Check for duplicates
   - Save to transactions table
6. Return success to app
```

### Manual Entry Flow
```
1. User enters transaction details
2. App sends to /ingest/manual
3. Backend generates fingerprint
4. Saves directly to transactions
5. Return success
```

## MVP Build Plan (4 weeks)

### Week 1: Foundation
- [x] Database schema
- [x] Auth system (signup/login)
- [x] Basic API structure
- [ ] Deploy to staging

### Week 2: Ingestion
- [x] SMS parsing logic
- [x] Manual entry endpoint
- [ ] Deduplication testing
- [ ] Mobile app integration

### Week 3: Analytics
- [x] Category breakdown
- [x] Weekly trends
- [ ] Budget tracking
- [ ] Financial health score

### Week 4: Polish
- [ ] Error handling
- [ ] Rate limiting
- [ ] Monitoring (Sentry)
- [ ] Production deployment

## Future Enhancements

### Phase 2 (Months 2-3)
- Statement upload (PDF/CSV parsing)
- Merchant database
- Budget alerts
- Savings goals tracking

### Phase 3 (Months 4-6)
- ML-based categorization
- Spending predictions
- Bill reminders
- Export reports (PDF/CSV)

### Phase 4 (Months 7-12)
- API integrations (Mono, Okra)
- Collaborative merchant data
- Financial health scoring
- Personalized insights

## Performance Targets

- API response time: < 200ms (p95)
- SMS processing: < 2s
- Database queries: < 50ms
- Uptime: 99.5%

## Monitoring

**Key metrics:**
- Transactions ingested/day
- Categorization accuracy
- API error rate
- Database query performance

**Tools:**
- Prometheus + Grafana (metrics)
- Sentry (error tracking)
- PostgreSQL slow query log

## Cost Breakdown (10k users)

| Item | Cost/month |
|------|------------|
| Server (2 vCPU, 4GB) | $12 |
| Database (managed) | $15 |
| Redis (optional) | $5 |
| Monitoring | $0 (free tier) |
| **Total** | **$32** |

## Contributing

1. Fork the repo
2. Create feature branch
3. Write tests
4. Submit PR

## License

MIT

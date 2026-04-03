# SPENDALT SYSTEM ARCHITECTURE

## EXECUTIVE SUMMARY

Spendalt is a cost-efficient, scalable personal finance analytics platform for Nigeria. This document outlines the complete technical architecture designed for a small team (1-3 developers) to scale from 0 to 1M users.

---

## 1. SYSTEM ARCHITECTURE

### High-Level Design

```
┌──────────────────────────────────────────────────────────┐
│                    MOBILE LAYER                          │
│  ┌────────────────┐              ┌────────────────┐     │
│  │   Android      │              │      iOS       │     │
│  │ (SMS Reader)   │              │ (Manual Only)  │     │
│  └────────┬───────┘              └────────┬───────┘     │
└───────────┼──────────────────────────────┼──────────────┘
            │                              │
            └──────────────┬───────────────┘
                           │ HTTPS/REST
                           ▼
┌──────────────────────────────────────────────────────────┐
│                   API GATEWAY (Go Fiber)                 │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Rate Limiter │ Auth │ CORS │ Logging │ Metrics │    │
│  └─────────────────────────────────────────────────┘    │
└───────────┬──────────────────────────────────────────────┘
            │
    ┌───────┴────────┬──────────┬──────────┬──────────┐
    ▼                ▼          ▼          ▼          ▼
┌────────┐      ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│  Auth  │      │Ingest  │ │Process │ │Analytics│ │ User  │
│Service │      │Service │ │Service │ │Service  │ │Service│
└────────┘      └───┬────┘ └───┬────┘ └────────┘ └────────┘
                    │          │
                    └────┬─────┘
                         ▼
                  ┌──────────────┐
                  │ Message Queue│
                  │  (Optional)  │
                  └──────┬───────┘
                         │
            ┌────────────┴────────────┐
            ▼                         ▼
    ┌───────────────┐         ┌──────────────┐
    │  PostgreSQL   │         │    Redis     │
    │  (Primary DB) │         │   (Cache)    │
    └───────────────┘         └──────────────┘
```

### Service Breakdown

**1. API Gateway (Go Fiber)**
- Single entry point for all requests
- Handles authentication, rate limiting, CORS
- Routes to appropriate service handlers
- Lightweight, fast (handles 100k+ req/s)

**2. Ingestion Service**
- Receives raw data (SMS, manual, statements)
- Minimal validation
- Saves to `raw_transactions` table
- Triggers async processing

**3. Processing Service**
- Categorizes transactions
- Normalizes merchant names
- Generates fingerprints for deduplication
- Saves to `transactions` table

**4. Analytics Service**
- Aggregates transaction data
- Calculates insights (spending, trends, budgets)
- Caches results in Redis

**5. User Service**
- Profile management
- Settings, preferences
- Budget and goal tracking

### Event-Driven vs Synchronous

**MVP: Synchronous with async processing**
```go
// Synchronous ingestion
POST /ingest/sms → Save to raw_transactions → Return 201

// Async processing (goroutine)
go processTransaction(rawID)
```

**Scale (10k+ users): Event-driven**
```
Ingestion → Publish to Queue → Worker consumes → Process → Save
```

**Why start synchronous?**
- Simpler to debug
- No queue infrastructure needed
- Good enough for 10k users
- Easy to migrate later

### Queue Strategy (Future)

**When to add queue:**
- Processing takes > 2 seconds
- Need retry logic
- Multiple workers needed
- 10k+ transactions/day

**Options:**
1. **RabbitMQ** (simple, reliable)
2. **Redis Streams** (already have Redis)
3. **Kafka** (overkill for MVP)

**Recommendation:** Start with Redis Streams when needed.

---

## 2. DATA INGESTION ENGINE

### Unified Ingestion Pipeline

```
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│   SMS   │  │ Manual  │  │Statement│  │   API   │
└────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
     │            │            │            │
     └────────────┴────────────┴────────────┘
                  │
                  ▼
          ┌───────────────┐
          │ Normalize to  │
          │ RawTransaction│
          └───────┬───────┘
                  │
                  ▼
          ┌───────────────┐
          │ Save to DB    │
          │ (raw_txns)    │
          └───────┬───────┘
                  │
                  ▼
          ┌───────────────┐
          │ Async Process │
          │ (categorize)  │
          └───────┬───────┘
                  │
                  ▼
          ┌───────────────┐
          │ Deduplicate   │
          │ (fingerprint) │
          └───────┬───────┘
                  │
                  ▼
          ┌───────────────┐
          │ Save to DB    │
          │ (transactions)│
          └───────────────┘
```

### Idempotency Strategy

**Fingerprint generation:**
```go
fingerprint = SHA256(
    user_id + 
    amount + 
    date (YYYY-MM-DD) + 
    merchant_normalized
)
```

**Database constraint:**
```sql
CREATE UNIQUE INDEX idx_fingerprint ON transactions(fingerprint);
```

**Result:** Duplicate inserts fail silently (ON CONFLICT DO NOTHING)

### Handling Missing Data

**SMS parsing failures:**
```go
if amount == 0 {
    // Save as raw, flag for manual review
    metadata["needs_review"] = true
    category = "Unprocessed"
}
```

**Missing merchant:**
```go
if merchant == "" {
    merchant = "Unknown"
    category = "Other"
}
```

**Strategy:** Save everything, improve later

### Retry & Failure Handling

**Transient failures (network, DB timeout):**
- Retry 3 times with exponential backoff
- Log to error tracking (Sentry)

**Permanent failures (invalid data):**
- Save to `failed_transactions` table
- Alert admin
- User can retry manually

---

## 3. DATABASE DESIGN

### Schema Philosophy

**Two-tier storage:**
1. **Raw layer**: Immutable, original data
2. **Processed layer**: Clean, queryable data

**Why?**
- Can reprocess if logic improves
- Audit trail for compliance
- ML training data

### Core Tables

```sql
-- Raw ingestion
raw_transactions (
    id, user_id, source, raw_text, 
    amount, transaction_type, metadata, 
    processed, created_at
)

-- Processed data
transactions (
    id, user_id, amount, transaction_type,
    category, merchant_id, description,
    transaction_date, fingerprint, created_at
)

-- Merchant intelligence
merchants (
    id, name, normalized_name, 
    category, aliases[], logo_url
)

-- Categorization rules
categories (
    id, name, icon, color, keywords[]
)
```

### Indexing Strategy

**Critical indexes:**
```sql
-- User's recent transactions (most common query)
CREATE INDEX idx_tx_user_date ON transactions(user_id, transaction_date DESC);

-- Category aggregation
CREATE INDEX idx_tx_category ON transactions(user_id, category);

-- Deduplication
CREATE UNIQUE INDEX idx_fingerprint ON transactions(fingerprint);

-- Merchant lookup
CREATE INDEX idx_merchant_normalized ON merchants(normalized_name);
```

**Query optimization:**
- Limit results (LIMIT 50)
- Use covering indexes
- Partition by date (future)

### Performance Targets

| Query | Target | Strategy |
|-------|--------|----------|
| Get recent transactions | < 50ms | Index on (user_id, date) |
| Category breakdown | < 100ms | Materialized view (future) |
| Monthly summary | < 200ms | Redis cache |

---

## 4. MERCHANT & CATEGORY INTELLIGENCE

### Rule-Based System (MVP)

**Keyword matching:**
```go
keywords := map[string][]string{
    "Food": {"restaurant", "kfc", "dominos", "cafe"},
    "Transport": {"uber", "bolt", "fuel"},
    "Bills": {"nepa", "dstv", "mtn", "airtel"},
}

func categorize(description string) string {
    desc := strings.ToLower(description)
    for category, words := range keywords {
        for _, word := range words {
            if strings.Contains(desc, word) {
                return category
            }
        }
    }
    return "Other"
}
```

### Merchant Normalization

**Problem:** "KFC Lagos", "KFC - VI", "K.F.C" → Same merchant

**Solution:**
```go
func normalize(name string) string {
    name = strings.ToLower(name)
    name = regexp.ReplaceAll(`[^a-z0-9\s]`, "", name)
    name = strings.TrimSpace(name)
    return name
}
// "kfc lagos" → "kfc"
```

**Merchant database:**
```sql
INSERT INTO merchants (name, normalized_name, aliases)
VALUES ('KFC', 'kfc', ARRAY['kfc lagos', 'kentucky fried chicken']);
```

### ML-Based Categorization (Future)

**Phase 1: Collect training data**
- User corrections (when they change category)
- Store in `category_corrections` table

**Phase 2: Train model**
- Simple logistic regression
- Features: merchant, amount, time, day of week
- Train monthly with new data

**Phase 3: Hybrid approach**
- Rule-based for known patterns
- ML for ambiguous cases
- Confidence score > 0.8 → auto-categorize
- Confidence < 0.8 → ask user

### Improvement Over Time

**Feedback loop:**
```
User corrects category → Save correction → 
Retrain model → Better predictions
```

**Collaborative learning:**
- Aggregate merchant data across users
- "100 users categorized 'Shoprite' as 'Shopping'"
- Apply to new users automatically

---

## 5. COST OPTIMIZATION STRATEGY

### API Usage Minimization

**Problem:** Mono charges per API call
- Account balance: ₦5/call
- Transaction sync: ₦10/call
- 10k users × 30 calls/month = ₦3M/month ($4k)

**Solution: Avoid APIs entirely (MVP)**
- SMS parsing (free)
- Manual entry (free)
- Statement upload (free)

**When to use APIs:**
- User explicitly requests sync
- Charge user premium fee
- Or limit to 1 sync/week

### Caching Strategy

**Redis caching:**
```go
// Cache monthly summary (expensive query)
key := fmt.Sprintf("insights:%s:%s", userID, month)
cached, err := redis.Get(key)
if err == nil {
    return cached
}

// Calculate and cache for 1 hour
result := calculateInsights(userID, month)
redis.Set(key, result, 1*time.Hour)
```

**What to cache:**
- Monthly insights (1 hour TTL)
- Category breakdown (1 hour TTL)
- Merchant logos (24 hour TTL)

**What NOT to cache:**
- Recent transactions (changes frequently)
- User profile (small, fast query)

### Database Optimization

**Connection pooling:**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Query optimization:**
- Use prepared statements
- Avoid N+1 queries
- Batch inserts

---

## 6. SCALABILITY PLAN

### 0 → 10k Users

**Infrastructure:**
- 1 server (2 vCPU, 4GB RAM)
- PostgreSQL on same server
- No Redis needed yet

**Cost:** $20/month (DigitalOcean, Hetzner)

**Bottlenecks:**
- None (can handle 100k req/day)

### 10k → 100k Users

**Infrastructure:**
- 2-3 API servers (load balanced)
- Separate DB server (4 vCPU, 8GB RAM)
- Redis for caching
- CDN for static assets

**Cost:** $100-150/month

**Bottlenecks:**
- Database writes (add read replica)
- Processing queue (add workers)

### 100k → 1M Users

**Infrastructure:**
- 5-10 API servers (auto-scaling)
- Database cluster (primary + 2 replicas)
- Message queue (RabbitMQ/Kafka)
- Microservices split
- Monitoring (Prometheus, Grafana)

**Cost:** $500-1000/month

**Bottlenecks:**
- Database sharding (by user_id)
- Distributed caching
- CDN for API responses

### Horizontal Scaling Strategy

**Stateless API servers:**
- No session storage (use JWT)
- Any server can handle any request
- Easy to add/remove servers

**Database scaling:**
```
Primary (writes) → Replicas (reads)
```

**Load balancing:**
- Nginx or cloud load balancer
- Round-robin or least connections

---

## 7. SECURITY & COMPLIANCE

### Data Encryption

**At rest:**
- PostgreSQL: Encrypted volumes
- Backups: Encrypted with AES-256

**In transit:**
- HTTPS only (TLS 1.3)
- Certificate from Let's Encrypt

### Token Storage

**JWT tokens:**
- Stored in mobile app secure storage
- 30-day expiry
- Refresh token for renewal

**Never store:**
- Passwords in plain text
- API keys in code
- Secrets in version control

### Access Control

**Role-based:**
```go
type Role string
const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
)
```

**Principle of least privilege:**
- Users can only access their own data
- Admins can view (not modify) user data

### Nigeria Compliance (NDPR)

**Requirements:**
1. User consent for data collection
2. Right to data deletion
3. Data breach notification (72 hours)
4. Data localization (store in Nigeria)

**Implementation:**
```sql
-- Soft delete
UPDATE users SET deleted_at = NOW() WHERE id = $1;

-- Hard delete (after 30 days)
DELETE FROM users WHERE deleted_at < NOW() - INTERVAL '30 days';
```

---

## 8. MOBILE APP ARCHITECTURE

### React Native Structure

```
app/
├── screens/          # UI screens
├── components/       # Reusable components
├── services/         # API calls
├── store/            # State management
├── utils/            # Helpers
└── constants/        # Config
```

### State Management

**Zustand (recommended):**
```javascript
const useStore = create((set) => ({
  transactions: [],
  addTransaction: (tx) => set((state) => ({
    transactions: [tx, ...state.transactions]
  })),
}));
```

**Why Zustand?**
- Simpler than Redux
- No boilerplate
- Good performance

### Offline-First Strategy

**Local storage:**
```javascript
// Save transaction locally
await AsyncStorage.setItem('pending_tx', JSON.stringify(tx));

// Sync when online
if (isOnline) {
  await api.syncTransactions();
}
```

**Conflict resolution:**
- Server wins (simpler)
- Or last-write-wins with timestamp

### Sync Strategy

**Pull-based (recommended):**
```javascript
// User opens app
onAppOpen(() => {
  fetchRecentTransactions();
});

// Background sync (every 6 hours)
BackgroundFetch.configure({
  minimumFetchInterval: 360,
  callback: syncTransactions,
});
```

**Push-based (future):**
- WebSocket connection
- Real-time updates
- More complex, higher cost

---

## 9. MVP BUILD PLAN

### Phase 1: Foundation (Week 1)

**Backend:**
- [x] Database schema
- [x] Auth endpoints (signup/login)
- [x] Basic API structure
- [ ] Deploy to staging (Railway, Render)

**Mobile:**
- [ ] Auth screens (login, signup)
- [ ] API service layer
- [ ] Token storage

**Goal:** User can create account and login

### Phase 2: Ingestion (Week 2)

**Backend:**
- [x] SMS parsing logic
- [x] Manual entry endpoint
- [ ] Deduplication testing

**Mobile:**
- [ ] SMS permission request (Android)
- [ ] SMS reader service
- [ ] Manual entry form
- [ ] Transaction list screen

**Goal:** User can add transactions

### Phase 3: Analytics (Week 3)

**Backend:**
- [x] Category breakdown endpoint
- [x] Weekly trend endpoint
- [ ] Budget tracking

**Mobile:**
- [ ] Dashboard screen
- [ ] Insights screen
- [ ] Charts (react-native-chart-kit)

**Goal:** User sees spending insights

### Phase 4: Polish (Week 4)

**Backend:**
- [ ] Error handling
- [ ] Rate limiting
- [ ] Logging (structured)
- [ ] Monitoring (Sentry)

**Mobile:**
- [ ] Error boundaries
- [ ] Loading states
- [ ] Empty states
- [ ] Onboarding flow

**Goal:** Production-ready MVP

### What to Delay

**Not in MVP:**
- Statement upload (complex parsing)
- API integrations (expensive)
- ML categorization (need data first)
- Bill reminders (nice-to-have)
- Export reports (low priority)

**Add in Phase 2 (Months 2-3)**

### Tradeoffs

| Decision | Tradeoff |
|----------|----------|
| SMS-only (no API) | Limited to Android, but saves $4k/month |
| Rule-based categorization | 80% accuracy, but simple and fast |
| Synchronous processing | Slower at scale, but easier to debug |
| Single server | No redundancy, but costs $20/month |

---

## CONCLUSION

This architecture is designed for:
- **Cost efficiency**: $20/month for 10k users
- **Simplicity**: 1-3 developers can build and maintain
- **Scalability**: Clear path to 1M users
- **Pragmatism**: MVP in 4 weeks, iterate based on data

**Next steps:**
1. Set up development environment
2. Deploy staging server
3. Build mobile app auth flow
4. Test SMS ingestion
5. Launch beta with 100 users
6. Iterate based on feedback

**Success metrics:**
- 1000 users in Month 1
- 80% categorization accuracy
- < 200ms API response time
- $0 API costs (SMS-based)

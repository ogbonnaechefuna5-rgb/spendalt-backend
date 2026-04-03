# Spendalt API Documentation

Base URL: `http://localhost:8080/api/v1`

## Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

---

## Auth Endpoints

### POST /auth/signup
Create a new user account.

**Request:**
```json
{
  "phone": "+2348012345678",
  "password": "securepassword123",
  "full_name": "John Doe"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "uuid-here"
}
```

### POST /auth/login
Login to existing account.

**Request:**
```json
{
  "phone": "+2348012345678",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "uuid-here"
}
```

---

## User Profile Endpoints

### GET /user/profile
Get current user profile.

**Response:**
```json
{
  "id": "uuid",
  "phone": "+2348012345678",
  "email": "john@example.com",
  "full_name": "John Doe",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### PUT /user/profile
Update user profile.

**Request:**
```json
{
  "full_name": "John Updated",
  "email": "newemail@example.com"
}
```

**Response:**
```json
{
  "success": true
}
```

### POST /user/change-password
Change user password.

**Request:**
```json
{
  "old_password": "oldpass123",
  "new_password": "newpass456"
}
```

**Response:**
```json
{
  "success": true
}
```

### DELETE /user/account
Delete user account (soft delete).

**Response:**
```json
{
  "success": true,
  "message": "account deleted"
}
```

---

## Transaction Endpoints

### POST /transactions/ingest/sms
Ingest transaction from SMS.

**Request:**
```json
{
  "raw_text": "GTBank: Your account has been debited with NGN5,000.00 at KFC Lagos. Balance: NGN45,000.00"
}
```

**Response:**
```json
{
  "id": "uuid",
  "status": "processing"
}
```

### POST /transactions/ingest/manual
Manually add a transaction.

**Request:**
```json
{
  "amount": 5000.00,
  "transaction_type": "debit",
  "description": "Lunch at KFC",
  "category": "Food",
  "transaction_date": "2024-01-15T12:30:00Z"
}
```

**Response:**
```json
{
  "id": "uuid"
}
```

### GET /transactions
Get user's transactions (last 50).

**Response:**
```json
[
  {
    "id": "uuid",
    "amount": 5000.00,
    "transaction_type": "debit",
    "category": "Food",
    "description": "Lunch at KFC",
    "transaction_date": "2024-01-15T12:30:00Z",
    "created_at": "2024-01-15T12:31:00Z"
  }
]
```

---

## Category Endpoints

### GET /categories
Get all available categories.

**Response:**
```json
[
  {
    "id": "uuid",
    "name": "Food",
    "icon": "🍔",
    "color": "#F59E0B"
  }
]
```

### GET /categories/breakdown?period=month
Get spending breakdown by category.

**Query Parameters:**
- `period`: `week`, `month` (default), or `year`

**Response:**
```json
[
  {
    "category": "Food",
    "count": 15,
    "total": 45000.00
  },
  {
    "category": "Transport",
    "count": 8,
    "total": 12000.00
  }
]
```

---

## Budget Endpoints

### POST /budgets
Create a new budget.

**Request:**
```json
{
  "category": "Food",
  "amount": 50000.00,
  "period": "monthly",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

**Response:**
```json
{
  "id": "uuid"
}
```

### GET /budgets
Get all user budgets with spending.

**Response:**
```json
[
  {
    "id": "uuid",
    "category": "Food",
    "amount": 50000.00,
    "spent": 32000.00,
    "period": "monthly",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### PUT /budgets/:id
Update a budget.

**Request:**
```json
{
  "category": "Food",
  "amount": 60000.00,
  "period": "monthly",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

**Response:**
```json
{
  "success": true
}
```

### DELETE /budgets/:id
Delete a budget.

**Response:**
```json
{
  "success": true
}
```

---

## Savings Goals Endpoints

### POST /savings
Create a savings goal.

**Request:**
```json
{
  "name": "Laptop Fund",
  "target_amount": 500000.00,
  "deadline": "2024-12-31"
}
```

**Response:**
```json
{
  "id": "uuid"
}
```

### GET /savings
Get all savings goals.

**Response:**
```json
[
  {
    "id": "uuid",
    "name": "Laptop Fund",
    "target_amount": 500000.00,
    "current_amount": 120000.00,
    "deadline": "2024-12-31T00:00:00Z",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### PUT /savings/:id/progress
Update savings progress (add money).

**Request:**
```json
{
  "amount": 10000.00
}
```

**Response:**
```json
{
  "success": true
}
```

### DELETE /savings/:id
Delete a savings goal.

**Response:**
```json
{
  "success": true
}
```

---

## Analytics Endpoints

### GET /analytics/insights
Get monthly financial insights.

**Response:**
```json
{
  "total_spending": 145000.00,
  "total_income": 250000.00,
  "categories": [
    {
      "category": "Food",
      "amount": 58000.00
    },
    {
      "category": "Transport",
      "amount": 29000.00
    }
  ]
}
```

### GET /analytics/weekly-trend
Get 7-day spending trend.

**Response:**
```json
[
  {
    "date": "2024-01-15",
    "amount": 12000.00
  },
  {
    "date": "2024-01-16",
    "amount": 8500.00
  }
]
```

---

## Financial Health Endpoints

### GET /health/score
Get financial health score and insights.

**Response:**
```json
{
  "score": 82,
  "grade": "Excellent",
  "percentile": 85,
  "insights": [
    {
      "category": "Savings Ratio",
      "status": "Excellent",
      "score": 85,
      "description": "Percentage of income saved each month"
    },
    {
      "category": "Budget Adherence",
      "status": "Good",
      "score": 78,
      "description": "How well you stick to your budgets"
    },
    {
      "category": "Spending Consistency",
      "status": "Excellent",
      "score": 83,
      "description": "Stability of your spending patterns"
    }
  ],
  "recommendations": [
    "Great job! Keep maintaining your financial discipline"
  ]
}
```

---

## Error Responses

All endpoints may return these error responses:

**400 Bad Request:**
```json
{
  "error": "invalid request"
}
```

**401 Unauthorized:**
```json
{
  "error": "unauthorized"
}
```

**404 Not Found:**
```json
{
  "error": "resource not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "server error"
}
```

---

## Rate Limiting

- 100 requests per minute per IP
- 1000 requests per hour per user

---

## Mobile App Integration Examples

### React Native - SMS Ingestion
```typescript
// After receiving SMS on Android
const ingestSMS = async (smsText: string) => {
  const response = await fetch('http://api.spendalt.com/api/v1/transactions/ingest/sms', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ raw_text: smsText }),
  });
  return response.json();
};
```

### React Native - Get Dashboard Data
```typescript
const getDashboardData = async () => {
  const [transactions, insights, budgets] = await Promise.all([
    fetch('/api/v1/transactions', { headers: { Authorization: `Bearer ${token}` } }),
    fetch('/api/v1/analytics/insights', { headers: { Authorization: `Bearer ${token}` } }),
    fetch('/api/v1/budgets', { headers: { Authorization: `Bearer ${token}` } }),
  ]);
  
  return {
    transactions: await transactions.json(),
    insights: await insights.json(),
    budgets: await budgets.json(),
  };
};
```

---

## Testing with cURL

### Signup
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"phone":"+2348012345678","password":"test123","full_name":"Test User"}'
```

### Get Transactions
```bash
curl http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Create Budget
```bash
curl -X POST http://localhost:8080/api/v1/budgets \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"category":"Food","amount":50000,"period":"monthly","start_date":"2024-01-01","end_date":"2024-01-31"}'
```

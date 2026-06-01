# Notification Module

Complete in-app notification system with local push notification support.

## Backend (`internal/notification`)

### Database
- **Table**: `notifications`
- **Columns**: `id`, `user_id`, `type`, `title`, `body`, `read`, `read_at`, `ref_id`, `created_at`
- **Indexes**: `user_id + created_at DESC`, `user_id + read` (for unread queries)

### API Endpoints
All routes require authentication (`/api/v1/notifications`):

- `GET /notifications?page=1&limit=30` — List notifications with pagination + unread count
- `POST /notifications/:id/read` — Mark one notification as read
- `POST /notifications/read-all` — Mark all notifications as read
- `DELETE /notifications/:id` — Delete a notification

### Types
```go
type NotifType string

const (
    TypeTransaction  NotifType = "transaction"
    TypeBudgetAlert  NotifType = "budget_alert"
    TypeAIInsight    NotifType = "ai_insight"
    TypeSavings      NotifType = "savings"
    TypeSystem       NotifType = "system"
)
```

### Creating Notifications (Server-Side)
```go
notifService.Create(
    userID,
    notification.TypeBudgetAlert,
    "Budget Alert",
    "You've used 95% of your Bills budget",
    &budgetID, // optional ref_id
)
```

## Flutter (`lib/`)

### Models
- `NotificationItem` (`models/notification_item.dart`) — matches backend `Notification` struct
- `NotifType` enum — `transaction`, `budgetAlert`, `aiInsight`, `savings`, `system`

### Services
- `NotificationApi` (`services/notification_api.dart`) — HTTP client for notification endpoints
- `NotificationService` (`services/notification_service.dart`) — Local push notification helpers

### Provider
- `NotificationProvider` (`providers/notification_provider.dart`) — manages in-app notification list
  - `load()` — initial fetch
  - `loadMore()` — pagination
  - `markRead(id)` — mark one read
  - `markAllRead()` — mark all read
  - `delete(id)` — delete one

### UI
- `NotificationPane` (`widgets/notification_pane.dart`) — bottom sheet with real data from `NotificationProvider`
  - Swipe-to-delete
  - Infinite scroll
  - Unread badge
  - "Mark all read" button

### Local Push Notifications
`NotificationService` provides typed helpers for each category:

```dart
// Transaction alert
await NotificationService.showTransactionAlert(
  title: 'New Transaction',
  body: 'GTBank — ₦12,500 debited for Shoprite',
);

// Budget warning
await NotificationService.showBudgetWarning(
  title: 'Budget Alert',
  body: "You've used 95% of your Bills budget",
);

// AI insight
await NotificationService.showAIInsight(
  title: 'Spending Pattern',
  body: 'You spend 40% more on weekends',
);

// General
await NotificationService.showGeneral(
  title: 'Sync Complete',
  body: 'All 3 accounts synced successfully',
);
```

### Permission Handling
```dart
// Request permission (Android 13+ / iOS)
final granted = await NotificationService.requestPermission();

// Check current permission status
final hasPermission = await NotificationService.hasPermission();
```

### Channels (Android)
- `moninte_transactions` — High priority, sound
- `moninte_budget` — High priority, sound
- `moninte_insights` — Default priority
- `moninte_general` — Default priority

## Integration Points

### Onboarding
- Step 5 now saves notification preferences to the backend via `savePreferences()`

### Permissions Screen
- Notification toggle now calls `NotificationService.requestPermission()` and updates state based on actual OS permission

### Dashboard
- Bell icon opens `NotificationPane` (already wired)
- Debug FAB (debug mode only) fires a test notification

### Preferences Screen
- Notification toggles sync to backend and control which types of notifications the user receives
- Backend preferences are stored in `user_preferences` table

## Future Enhancements

### Push Notifications (FCM/APNs)
To add remote push:
1. Add `firebase_messaging` package to `pubspec.yaml`
2. Create `POST /user/device-token` endpoint to register FCM/APNs tokens
3. Store tokens in `device_tokens` table
4. Add Firebase Admin SDK (Go) or APNs HTTP/2 client to backend
5. Call push service when creating notifications server-side

### Budget Threshold Alerts
1. Add `spent` field to `budgets` table
2. Update `spent` after each transaction ingestion
3. Check `spent / amount` ratio and create notification if > 0.8 or > 1.0

### Transaction Worker Hook
In `transaction/worker.go`, after `w.repo.Create(tx)`:
```go
// Create in-app notification
notifService.Create(
    tx.UserID,
    notification.TypeTransaction,
    "New Transaction",
    fmt.Sprintf("%s — %s %s", tx.Merchant, formatCurrency(tx.Amount), tx.Type),
    &tx.ID,
)

// Fire local push (requires FCM/APNs token lookup + send)
```

### AI Insights
Weekly cron job that:
1. Analyzes spending patterns per user
2. Creates `TypeAIInsight` notifications with personalized tips
3. Sends push notification if user has `ai_insights` preference enabled

## Testing

### Backend
```bash
# Start server
cd /Users/Nightwisper/Documents/backend
go run cmd/api/main.go

# Create a test notification (requires auth token)
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "transaction",
    "title": "Test Transaction",
    "body": "This is a test notification"
  }'

# List notifications
curl http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN"
```

### Flutter
```bash
cd /Users/Nightwisper/moninte

# Run on emulator
flutter run

# Test local push (tap debug FAB on dashboard in debug mode)
# Test in-app list (tap bell icon on dashboard)
```

## Files Changed/Created

### Backend
- `internal/notification/models.go` ✨
- `internal/notification/migrations/0001_init.sql` ✨
- `internal/notification/migrations.go` ✨
- `internal/notification/repository.go` ✨
- `internal/notification/service.go` ✨
- `internal/notification/response.go` ✨
- `internal/notification/handler.go` ✨
- `cmd/api/main.go` (wired notification package)

### Flutter
- `lib/models/notification_item.dart` ✨
- `lib/services/notification_api.dart` ✨
- `lib/services/notification_service.dart` ✨
- `lib/providers/notification_provider.dart` ✨
- `lib/services/api_service.dart` (added notification methods)
- `lib/services/api_client.dart` (fixed base URL for emulator)
- `lib/widgets/notification_pane.dart` (replaced mock data with real provider)
- `lib/screens/dashboard_screen.dart` (updated debug FAB to use NotificationService)
- `lib/screens/permissions_screen.dart` (added real notification permission request)
- `lib/screens/onboarding_screen.dart` (save notification preferences)
- `lib/main.dart` (wired NotificationProvider, init NotificationService)

✨ = new file

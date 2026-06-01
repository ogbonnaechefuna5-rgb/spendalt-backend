package notification

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMSender sends push notifications via Firebase Cloud Messaging.
// It is nil-safe: if Firebase was not configured (no credentials) all
// Send calls are silently skipped so the rest of the app keeps working.
type FCMSender struct {
	client *messaging.Client
}

// NewFCMSender initialises a Firebase app from a service-account JSON file.
// credentialsFile is the path to the downloaded service-account JSON.
// Returns a no-op sender (not nil) when credentialsFile is empty so callers
// never need to nil-check.
func NewFCMSender(ctx context.Context, credentialsFile string) *FCMSender {
	if credentialsFile == "" {
		log.Println("[fcm] FIREBASE_CREDENTIALS not set — push notifications disabled")
		return &FCMSender{}
	}
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Printf("[fcm] failed to init Firebase app: %v — push notifications disabled", err)
		return &FCMSender{}
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("[fcm] failed to get Messaging client: %v — push notifications disabled", err)
		return &FCMSender{}
	}
	log.Println("[fcm] Firebase Messaging initialised")
	return &FCMSender{client: client}
}

// Enabled reports whether FCM is configured and ready.
func (f *FCMSender) Enabled() bool { return f.client != nil }

// SendToTokens sends a notification to a list of FCM registration tokens.
// Silently skips if FCM is not configured or the token list is empty.
func (f *FCMSender) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) {
	if !f.Enabled() || len(tokens) == 0 {
		return
	}
	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: fcmChannelID(data["type"]),
				Icon:      "ic_notification",
				Sound:     "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: intPtr(1),
				},
			},
		},
		Data: data,
	}
	resp, err := f.client.SendEachForMulticast(ctx, msg)
	if err != nil {
		log.Printf("[fcm] SendEachForMulticast error: %v", err)
		return
	}
	if resp.FailureCount > 0 {
		log.Printf("[fcm] %d/%d messages failed", resp.FailureCount, len(tokens))
	}
}

// fcmChannelID maps a notification type string to an Android channel ID.
func fcmChannelID(notifType string) string {
	switch notifType {
	case string(TypeTransaction):
		return "moninte_transactions"
	case string(TypeBudgetAlert):
		return "moninte_budget"
	case string(TypeAIInsight):
		return "moninte_insights"
	default:
		return "moninte_general"
	}
}

func intPtr(i int) *int { return &i }

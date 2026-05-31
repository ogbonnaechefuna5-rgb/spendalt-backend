package payment

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type PaystackProvider struct {
	secretKey string
}

func NewPaystackProvider(secretKey string) *PaystackProvider {
	return &PaystackProvider{secretKey: secretKey}
}

func (p *PaystackProvider) Name() string { return "paystack" }

func (p *PaystackProvider) HandleWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	if !p.verifySignature(payload, signature) {
		return nil, errors.New("invalid paystack signature")
	}

	var body struct {
		Event string `json:"event"`
		Data  struct {
			Reference string  `json:"reference"`
			Amount    float64 `json:"amount"` // in kobo
			Status    string  `json:"status"`
			Metadata  struct {
				UserID string `json:"user_id"`
				PlanID string `json:"plan_id"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, err
	}
	if body.Event != "charge.success" || body.Data.Status != "success" {
		return nil, fmt.Errorf("unhandled event: %s", body.Event)
	}

	periodEnd := time.Now().AddDate(0, 1, 0) // default: 1 month
	return &WebhookEvent{
		Type:      EventChargeSuccess,
		UserID:    body.Data.Metadata.UserID,
		PlanID:    body.Data.Metadata.PlanID,
		Reference: body.Data.Reference,
		Amount:    body.Data.Amount / 100, // kobo → naira
		PeriodEnd: &periodEnd,
	}, nil
}

func (p *PaystackProvider) VerifyTransaction(reference string) (*TransactionResult, error) {
	req, _ := http.NewRequest("GET", "https://api.paystack.co/transaction/verify/"+reference, nil)
	req.Header.Set("Authorization", "Bearer "+p.secretKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body struct {
		Status bool `json:"status"`
		Data   struct {
			Reference string  `json:"reference"`
			Amount    float64 `json:"amount"`
			Status    string  `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &TransactionResult{
		Reference: body.Data.Reference,
		Amount:    body.Data.Amount / 100,
		Status:    body.Data.Status,
	}, nil
}

func (p *PaystackProvider) verifySignature(payload []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(p.secretKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

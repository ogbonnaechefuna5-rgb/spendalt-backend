package transaction

// IngestSMSRequest is the body for POST /transactions/ingest/sms.
type IngestSMSRequest struct {
	SMSText string `json:"sms_text"`
}

// IngestManualRequest is the body for POST /transactions/ingest/manual.
type IngestManualRequest struct {
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Merchant    string  `json:"merchant"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

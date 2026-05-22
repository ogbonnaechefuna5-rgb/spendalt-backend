package transaction

import "strings"

type Categorizer interface {
	Categorize(merchant string) string
}

type ruleEngine struct{}

func NewRuleEngine() Categorizer { return &ruleEngine{} }

// categoryRule pairs a category name with its keywords.
// Order matters: first match wins, so more specific categories come first.
type categoryRule struct {
	category string
	keywords []string
}

// categoryRules is an ordered slice so iteration is deterministic.
// A merchant matching keywords in two categories always resolves to the
// first matching rule — no random map iteration.
var categoryRules = []categoryRule{
	{"Income", []string{"salary", "credit alert", "received from", "inflow", "deposit", "refund", "cashback", "dividend", "interest"}},
	{"Airtime & Data", []string{"airtime", "data", "recharge", "mtn", "glo", "airtel", "9mobile", "etisalat", "topup", "bundle"}},
	{"Food & Dining", []string{"restaurant", "food", "cafe", "pizza", "burger", "kfc", "dominos", "eatery", "canteen", "kitchen", "suya", "shawarma", "chicken", "rice", "lunch", "dinner", "breakfast"}},
	{"Transportation", []string{"uber", "bolt", "fuel", "petrol", "transport", "bus", "taxi", "ride", "okada", "tricycle", "keke", "logistics", "dispatch", "lasgidi", "move"}},
	{"Shopping", []string{"mall", "store", "shop", "market", "supermarket", "shoprite", "jumia", "konga", "spar", "ebeano", "grocery", "fashion", "clothing", "shoes", "bag", "electronics"}},
	{"Entertainment", []string{"cinema", "movie", "game", "club", "bar", "netflix", "spotify", "showmax", "dstv", "gotv", "startimes", "concert", "event", "ticket", "streaming"}},
	{"Utilities", []string{"electric", "nepa", "phcn", "water", "internet", "ikedc", "ekedc", "aedc", "phedc", "ibedc", "cable", "wifi", "broadband", "bill", "subscription"}},
	{"Transfers", []string{"transfer", "sent to", "payment to", "remittance", "send money", "wire"}},
	{"Health", []string{"hospital", "pharmacy", "clinic", "doctor", "medical", "health", "drug", "chemist", "lab", "surgery"}},
	{"Education", []string{"school", "tuition", "fees", "university", "college", "course", "training", "exam", "waec", "jamb", "neco"}},
}

func (r *ruleEngine) Categorize(merchant string) string {
	m := strings.ToLower(strings.TrimSpace(merchant))
	for _, rule := range categoryRules {
		for _, kw := range rule.keywords {
			if strings.Contains(m, kw) {
				return rule.category
			}
		}
	}
	return "Other"
}

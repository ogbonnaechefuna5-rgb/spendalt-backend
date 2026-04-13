package transaction

import "strings"

type Categorizer interface {
	Categorize(merchant string) string
}

type ruleEngine struct{}

func NewRuleEngine() Categorizer { return &ruleEngine{} }

var categoryRules = map[string][]string{
	"Food & Dining":  {"restaurant", "food", "cafe", "pizza", "burger", "kfc", "dominos"},
	"Transportation": {"uber", "bolt", "fuel", "petrol", "transport", "bus", "taxi"},
	"Shopping":       {"mall", "store", "shop", "market", "supermarket", "shoprite", "jumia"},
	"Entertainment":  {"cinema", "movie", "game", "club", "bar", "netflix", "spotify"},
	"Utilities":      {"electric", "water", "internet", "phone", "bill", "mtn", "airtel"},
}

func (r *ruleEngine) Categorize(merchant string) string {
	m := strings.ToLower(strings.TrimSpace(merchant))
	for category, keywords := range categoryRules {
		for _, kw := range keywords {
			if strings.Contains(m, kw) {
				return category
			}
		}
	}
	return "Other"
}

package category

import (
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
)

type Repository interface {
	Seed() error
	GetAll(limit, offset int) ([]*Category, error)
	GetBreakdownByUserID(userID string, limit, offset int) ([]*CategoryBreakdown, error)
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(limit, offset int) ([]*Category, error) {
	rows, err := r.db.Query(
		`SELECT id, name, COALESCE(icon,''), COALESCE(color,''), COALESCE(keywords, ARRAY[]::text[])
		 FROM categories ORDER BY name LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(c *Category) []interface{} {
		return []interface{}{&c.ID, &c.Name, &c.Icon, &c.Color, &c.Keywords}
	})
}

func (r *repository) Seed() error {
	_, err := r.db.Exec(`
		INSERT INTO categories (name, icon, color, keywords) VALUES
		('Food & Dining',  '🍔', '#FF8C42', ARRAY['restaurant','food','cafe','pizza','burger','kfc','dominos','eatery','canteen','suya','shawarma','rice','lunch','dinner','breakfast']),
		('Transportation', '🚗', '#4D9FFF', ARRAY['uber','bolt','fuel','petrol','transport','bus','taxi','ride','okada','keke','logistics','dispatch']),
		('Shopping',       '🛒', '#A855F7', ARRAY['mall','store','shop','market','supermarket','shoprite','jumia','konga','spar','grocery','fashion','clothing','shoes','electronics']),
		('Entertainment',  '🎬', '#FF69B4', ARRAY['cinema','movie','game','club','bar','netflix','spotify','showmax','dstv','gotv','startimes','concert','ticket','streaming']),
		('Utilities',      '⚡', '#FFB830', ARRAY['electric','nepa','phcn','water','internet','ikedc','ekedc','aedc','cable','wifi','broadband','bill','subscription']),
		('Airtime & Data', '📱', '#4DFF91', ARRAY['airtime','data','recharge','mtn','glo','airtel','9mobile','etisalat','topup','bundle']),
		('Transfers',      '💸', '#8A9E90', ARRAY['transfer','sent to','payment to','remittance','send money','wire']),
		('Income',         '💰', '#A8FF3E', ARRAY['salary','credit alert','received from','inflow','deposit','refund','cashback','dividend','interest']),
		('Health',         '🏥', '#FF6B6B', ARRAY['hospital','pharmacy','clinic','doctor','medical','health','drug','chemist','lab','surgery']),
		('Education',      '📚', '#6BC5FF', ARRAY['school','tuition','fees','university','college','course','training','exam','waec','jamb','neco']),
		('Other',          '📦', '#8A9E90', ARRAY[]::text[])
		ON CONFLICT (name) DO UPDATE
		  SET icon = EXCLUDED.icon, color = EXCLUDED.color, keywords = EXCLUDED.keywords
	`)
	return err
}

func (r *repository) GetBreakdownByUserID(userID string, limit, offset int) ([]*CategoryBreakdown, error) {
	rows, err := r.db.Query(
		`SELECT COALESCE(category, 'Other'), COUNT(*), SUM(amount) as total
		 FROM transactions WHERE user_id = $1 AND transaction_type = 'debit'
		 GROUP BY category ORDER BY total DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(cb *CategoryBreakdown) []interface{} {
		return []interface{}{&cb.Category, &cb.Count, &cb.Total}
	})
}

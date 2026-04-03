CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    icon VARCHAR(50),
    color VARCHAR(7),
    keywords TEXT[]
);

INSERT INTO categories (name, icon, keywords) VALUES
('Food', '🍔', ARRAY['restaurant', 'food', 'lunch', 'dinner', 'cafe', 'pizza', 'kfc']),
('Transport', '🚗', ARRAY['uber', 'bolt', 'fuel', 'petrol', 'taxi']),
('Shopping', '🛍️', ARRAY['jumia', 'konga', 'market', 'shopping', 'store']),
('Bills', '💡', ARRAY['nepa', 'phcn', 'dstv', 'gotv', 'electricity', 'water']),
('Airtime', '📱', ARRAY['airtime', 'data', 'recharge', 'mtn', 'glo', 'airtel', '9mobile']),
('Transfer', '💸', ARRAY['transfer', 'sent to', 'payment to']),
('Income', '💰', ARRAY['salary', 'credit alert', 'received from']),
('Other', '📦', ARRAY[])
ON CONFLICT (name) DO NOTHING;

-- +goose Up
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    icon VARCHAR(50),
    color VARCHAR(7),
    keywords TEXT[]
);

INSERT INTO categories (name, icon, color, keywords) VALUES
('Food & Dining',  '🍔', '#FF8C42', ARRAY['restaurant','food','cafe','pizza','burger','kfc','dominos','eatery','canteen','kitchen','suya','shawarma','chicken','rice','lunch','dinner','breakfast']),
('Transportation', '🚗', '#4D9FFF', ARRAY['uber','bolt','fuel','petrol','transport','bus','taxi','ride','okada','tricycle','keke','logistics','dispatch','lasgidi','move']),
('Shopping',       '🛒', '#A855F7', ARRAY['mall','store','shop','market','supermarket','shoprite','jumia','konga','spar','ebeano','grocery','fashion','clothing','shoes','bag','electronics']),
('Entertainment',  '🎬', '#FF69B4', ARRAY['cinema','movie','game','club','bar','netflix','spotify','showmax','dstv','gotv','startimes','concert','event','ticket','streaming']),
('Utilities',      '⚡', '#FFB830', ARRAY['electric','nepa','phcn','water','internet','ikedc','ekedc','aedc','phedc','ibedc','cable','wifi','broadband','bill','subscription']),
('Airtime & Data', '📱', '#4DFF91', ARRAY['airtime','data','recharge','mtn','glo','airtel','9mobile','etisalat','topup','bundle']),
('Transfers',      '💸', '#8A9E90', ARRAY['transfer','sent to','payment to','remittance','send money','wire']),
('Income',         '💰', '#A8FF3E', ARRAY['salary','credit alert','received from','inflow','deposit','refund','cashback','dividend','interest']),
('Health',         '🏥', '#FF6B6B', ARRAY['hospital','pharmacy','clinic','doctor','medical','health','drug','chemist','lab','surgery']),
('Education',      '📚', '#6BC5FF', ARRAY['school','tuition','fees','university','college','course','training','exam','waec','jamb','neco']),
('Other',          '📦', '#8A9E90', ARRAY[]::text[])
ON CONFLICT (name) DO UPDATE
  SET icon = EXCLUDED.icon, color = EXCLUDED.color, keywords = EXCLUDED.keywords;

-- +goose Down
DROP TABLE IF EXISTS categories;

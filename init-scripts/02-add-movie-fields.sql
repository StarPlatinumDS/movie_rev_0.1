-- Добавляем новые поля в таблицу movies
ALTER TABLE movies 
ADD COLUMN IF NOT EXISTS genres TEXT[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS year INTEGER,
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS rating INTEGER CHECK (rating >= 1 AND rating <= 5),
ADD COLUMN IF NOT EXISTS world_rating NUMERIC(3,1) CHECK (world_rating >= 0.0 AND world_rating <= 10.0),
ADD COLUMN IF NOT EXISTS comment VARCHAR(250);
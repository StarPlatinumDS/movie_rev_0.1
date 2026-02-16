CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    picture_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    genres TEXT[] DEFAULT '{}',
    year INTEGER,
    description TEXT,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    world_rating NUMERIC(3,1) CHECK (world_rating >= 0.0 AND world_rating <= 10.0),
    comment VARCHAR(250)
);
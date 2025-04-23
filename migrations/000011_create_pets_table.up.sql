CREATE TABLE IF NOT EXISTS users (
       id SERIAL PRIMARY KEY,
       username TEXT,
       firstName TEXT,
       lastName TEXT,
       email TEXT,
       phone TEXT,
       password TEXT,
       userStatus INTEGER
);
CREATE TABLE IF NOT EXISTS token_blacklist (
    token_id TEXT PRIMARY KEY,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires ON token_blacklist (expires_at);
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
CREATE TABLE IF NOT EXISTS pets (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        photoUrls TEXT[],
        category_id INTEGER REFERENCES categories(id),
        status TEXT  
);
CREATE TABLE IF NOT EXISTS pet_tags (
    pet_id INTEGER REFERENCES pets(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (pet_id, tag_id)
);
 CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        complete BOOLEAN,
        petId INTEGER NOT NULL,
        quantity INTEGER NOT NULL DEFAULT 1,
        shipDate TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        status TEXT CHECK (status IN ('placed', 'approved', 'delivered')), 
        FOREIGN KEY (petId) REFERENCES pets(id)
);


CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'User',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    event_name VARCHAR(200),
    customer_name VARCHAR(200),
    phone VARCHAR(50),
    address TEXT,
    venue TEXT,
    status VARCHAR(20),
    date_time TIMESTAMP,
    total_cost NUMERIC(10,2),
    paid NUMERIC(10,2),
    balance NUMERIC(10,2),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    token TEXT NOT NULL,
    expires_at TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);

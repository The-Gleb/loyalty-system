CREATE TABLE IF NOT EXISTS users (
    login varchar(255) UNIQUE,
    password TEXT,
    current NUMERIC,
    withrawn NUMERIC
);
CREATE TABLE IF NOT EXISTS orders (
    order_user varchar(255),
    order_number TEXT UNIQUE,
    order_status varchar(255),
    order_accrual NUMERIC,
    uploaded_at timestamp

);
CREATE TABLE IF NOT EXISTS sessions (
    user varchar(255),
    session_token TEXT UNIQUE,
    expiry timestamp
);
CREATE TABLE IF NOT EXISTS withdrawals (
    user varchar(255),
    order TEXT UNIQUE,
    sum NUMERIC,
    processed_at timestamp
);
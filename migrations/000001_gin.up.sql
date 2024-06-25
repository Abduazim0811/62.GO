CREATE TABLE IF NOT EXISTS messages(
    username TEXT,    
    message TEXT,
    timestamp TIMESTAMPTZ,
    ip_address VARCHAR(64)
);
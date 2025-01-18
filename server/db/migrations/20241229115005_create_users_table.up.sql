
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    username VARCHAR(255) NOT NULL UNIQUE,         
    email VARCHAR(255) NOT NULL UNIQUE,            
    password TEXT NOT NULL,                        
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
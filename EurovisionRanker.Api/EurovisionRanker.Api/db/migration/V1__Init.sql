CREATE SCHEMA IF NOT EXISTS ranker;

-- 1. USERS
CREATE TABLE ranker.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE, 
    email TEXT NOT NULL,           
    password_hash TEXT NOT NULL,
    profile_pic_url TEXT,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add explicit unique constraint for email (essential for login lookups)
CREATE UNIQUE INDEX IF NOT EXISTS users_email ON ranker.users(email);


-- 2. RANKINGS
CREATE TABLE ranker.ranking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES ranker.users(id) ON DELETE CASCADE, -- Added Cascade for cleanup
    
    name TEXT NOT NULL,
    description TEXT,
    year INT NOT NULL,
    ranking_string TEXT NOT NULL,
    
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- FK Index: Essential for "Get all rankings for this user"
CREATE INDEX IF NOT EXISTS ranking_user_id ON ranker.ranking(user_id);

-- Composite Index: Highly efficient for "Get my rankings for 2025"
CREATE INDEX IF NOT EXISTS ranking_user_year ON ranker.ranking(user_id, year);


-- 3. GROUPS
CREATE TABLE ranker.groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    public_id TEXT NOT NULL UNIQUE 
);


-- 4. GROUP MEMBERS
CREATE TABLE ranker.group_member (
    group_id UUID REFERENCES ranker.groups(id) ON DELETE CASCADE,
    user_id UUID REFERENCES ranker.users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id) 
);

CREATE INDEX IF NOT EXISTS group_member_user_id ON ranker.group_member(user_id);
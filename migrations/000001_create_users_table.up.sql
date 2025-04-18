CREATE TABLE users (
  user_id UUID DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  email TEXT NOT NULL UNIQUE,
  password BYTEA NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  PRIMARY KEY(user_id)
);

CREATE TABLE sessions (
  session_id INT GENERATED ALWAYS AS IDENTITY,
  hash BYTEA NOT NULL,
  user_id UUID NOT NULL,
  valid_till TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  PRIMARY KEY(session_id),
  CONSTRAINT sessions_users_fk FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

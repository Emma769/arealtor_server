CREATE TABLE IF NOT EXISTS landlords(
  landlord_id UUID DEFAULT gen_random_uuid(),
  first_name VARCHAR(60) NOT NULL,
  last_name VARCHAR(60),
  email VARCHAR(60),
  phone TEXT NOT NULL UNIQUE,
  registered_by UUID NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  updated_at TIMESTAMP WITH TIME ZONE,
  PRIMARY KEY(landlord_id),
  CONSTRAINT landlords_users_fk FOREIGN KEY(registered_by) REFERENCES users(user_id) ON DELETE SET NULL
);

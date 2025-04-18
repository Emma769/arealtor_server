CREATE TABLE IF NOT EXISTS tenants (
  tenant_id UUID DEFAULT gen_random_uuid(),
  first_name VARCHAR(60) NOT NULL,
  last_name VARCHAR(60),
  gender VARCHAR(6) NOT NULL,
  dob DATE,
  image TEXT,
  email VARCHAR(60),
  phone VARCHAR(11) UNIQUE NOT NULL,
  state_of_origin VARCHAR(60),
  nationality VARCHAR(60),
  occupation VARCHAR(60),
  additional_info JSON DEFAULT '{}',
  registered_by UUID,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
  updated_at TIMESTAMP WITH TIME ZONE,
  PRIMARY KEY(tenant_id),
  CONSTRAINT tenants_users_fk FOREIGN KEY(registered_by) REFERENCES users(user_id) ON DELETE SET NULL
);

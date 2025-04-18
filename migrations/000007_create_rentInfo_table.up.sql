CREATE TABLE IF NOT EXISTS rent_info (
  rent_info_id INT GENERATED ALWAYS AS IDENTITY,
  start_date TIMESTAMP WITH TIME ZONE,
  maturity_date TIMESTAMP WITH TIME ZONE,
  renewal_date TIMESTAMP WITH TIME ZONE, 
  landlord_id UUID,
  tenant_id UUID NOT NULL,
  address TEXT NOT NULL,
  rent_fee NUMERIC NOT NULL,
  PRIMARY KEY(rent_info_id),
  CONSTRAINT rent_info_landlords_fk FOREIGN KEY(landlord_id) REFERENCES landlords(landlord_id) ON DELETE SET NULL,
  CONSTRAINT rent_info_tenants_fk FOREIGN KEY(tenant_id) REFERENCES tenants(tenant_id) ON DELETE CASCADE
);

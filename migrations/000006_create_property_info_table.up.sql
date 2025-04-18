CREATE TABLE IF NOT EXISTS property_info ( 
  property_info_id INT GENERATED ALWAYS AS IDENTITY,
  address TEXT NOT NULL,
  property_type INT DEFAULT 1,
  additional_info JSON DEFAULT '{}'::JSON,
  lease_price NUMERIC NOT NULL,
  lease_period INT NOT NULL,
  start_date TIMESTAMP WITH TIME ZONE NOT NULL,
  end_date TIMESTAMP WITH TIME ZONE NOT NULL,
  landlord_id UUID,
  PRIMARY KEY(property_info_id),
  CONSTRAINT property_info_landlords_fk FOREIGN KEY(landlord_id) REFERENCES landlords(landlord_id) ON DELETE SET NULL
);

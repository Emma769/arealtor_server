CREATE OR REPLACE function remove_old_session() RETURNS TRIGGER AS $$
BEGIN
  DELETE FROM sessions WHERE valid_till < CURRENT_TIMESTAMP - INTERVAL '1 minute';
  return NEW;
END;
$$ LANGUAGE PLPGSQL;

CREATE TRIGGER remove_old_session_trigger AFTER INSERT ON sessions EXECUTE PROCEDURE remove_old_session();

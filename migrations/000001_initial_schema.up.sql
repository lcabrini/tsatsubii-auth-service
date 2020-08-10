BEGIN;

CREATE TABLE IF NOT EXISTS users(
  id uuid,
  username VARCHAR(50) NOT NULL,
  password VARCHAR(100) NOT NULL,
  email VARCHAR(50) NOT NULL,
  phone VARCHAR(50) NOT NULL DEFAULT '',
  token VARCHAR(50) NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(id),
  UNIQUE(username),
  UNIQUE(email),
  UNIQUE(token)
);

DROP VIEW IF EXISTS userlist;
CREATE VIEW userlist AS
  SELECT id, username FROM USERS
  WHERE active = 't'
  ORDER BY username;

DROP FUNCTION IF EXISTS on_delete_user();
CREATE FUNCTION on_delete_user() RETURNS TRIGGER AS $$
BEGIN
  IF old.username = 'sa' THEN
    RAISE EXCEPTION 'cannot delete the system administrator';
  ELSE
    RETURN old;
  END IF;
END;
$$ LANGUAGE plpgsql;

COMMIT;

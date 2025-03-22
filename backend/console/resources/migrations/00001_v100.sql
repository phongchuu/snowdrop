-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_timestamp_trigger()
RETURNS TRIGGER AS $$
DECLARE
  current_username TEXT := current_setting('session.requester', true);
BEGIN
  current_username := COALESCE(NULLIF(TRIM(current_username), ''), 'system');

  IF TG_OP = 'INSERT' THEN
    NEW.created_at = CURRENT_TIMESTAMP;
    NEW.created_by = current_username;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.updated_by = current_username;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS public.users(
  id UUID PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password VARCHAR(60) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  created_by VARCHAR(255) NOT NULL,
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR(255)
);

CREATE TRIGGER trg_users_update_audit_timestamp
BEFORE INSERT OR UPDATE ON public.users
FOR EACH ROW
EXECUTE FUNCTION audit_timestamp_trigger();

INSERT INTO public.users (id, username, email, password)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin', 'admin@internal.com', crypt('Keep!t5ecret', gen_salt('bf', 12)));

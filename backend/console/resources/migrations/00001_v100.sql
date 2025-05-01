-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_timestamp_trigger()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    NEW.created_at = CURRENT_TIMESTAMP;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    NEW.updated_at = CURRENT_TIMESTAMP;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION audit_user_trigger()
RETURNS TRIGGER AS $$
DECLARE
  current_username TEXT := current_setting('session.requester', true);
BEGIN
  current_username := COALESCE(NULLIF(TRIM(current_username), ''), 'system');

  IF TG_OP = 'INSERT' THEN
    NEW.created_by = current_username;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    NEW.updated_by = current_username;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- [START] Table: public.users
CREATE TABLE IF NOT EXISTS public.users (
  id UUID PRIMARY KEY,
  username VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password VARCHAR(60) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  created_by VARCHAR(255) NOT NULL,
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR(255)
);
--
CREATE UNIQUE INDEX IF NOT EXISTS uniq_username_email
ON public.users
USING btree (username, email);
--
CREATE TRIGGER trg_users_set_timestamp
BEFORE INSERT OR UPDATE ON public.users
FOR EACH ROW
EXECUTE FUNCTION audit_timestamp_trigger();
--
CREATE TRIGGER trg_users_set_user_audit
BEFORE INSERT OR UPDATE ON public.users
FOR EACH ROW
EXECUTE FUNCTION audit_user_trigger();
--
INSERT INTO public.users (id, username, email, password)
VALUES
(
  '00000000-03e8-7000-8000-29d1c630b42e',
  'system',
  'system@internal.com',
-- +goose ENVSUB ON
  crypt('${APP_DEFAULT_SYSTEM_PASSWORD?Missing env: APP_DEFAULT_SYSTEM_PASSWORD}', gen_salt('bf', 12))
-- +goose ENVSUB OFF
),
(
  '00000000-07d0-7000-8000-8fbd4b99a4c9',
  'admin',
  'admin@internal.com',
  crypt(encode(gen_random_bytes(72), 'base64'), gen_salt('bf', 12))
),
(
  '00000000-0bb8-7000-8000-3e1b6bb7b5e4',
  'anonymous',
  'anonymous@internal.com',
  crypt(encode(gen_random_bytes(72), 'base64'), gen_salt('bf', 12))
);
-- [END] Table: public.users


-- [START] Table: public.sessions
CREATE TABLE public.sessions (
  id VARCHAR(255) PRIMARY KEY,
  user_id UUID NULL,
  data JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  last_accessed_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT chk_expires_after_created CHECK (expires_at > created_at),
  CONSTRAINT chk_last_accessed_after_created CHECK (last_accessed_at >= created_at)
);
--
COMMENT ON COLUMN public.sessions.user_id IS 'NULL for annonymous session';
--
CREATE INDEX idx_sessions_user_id ON sessions (user_id) WHERE user_id IS NOT NULL;
--
CREATE INDEX idx_sessions_expires ON sessions (expires_at);
-- [END] Table: public.sessions

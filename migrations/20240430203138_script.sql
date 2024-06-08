-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS public.script (
  id public.xid NOT NULL DEFAULT xid(),
  name string NOT NULL,
  workflow JSON NOT NULL,
  body_presets JSON NOT NULL,
  header_presets JSON NOT NULL,
  author string NOT NULL,
  warehouse_api_key VARCHAR(255) NOT NULL,
)

ALTER TABLE public.script
ADD CONSTRAINT script_pkey PRIMARY KEY (id, name);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

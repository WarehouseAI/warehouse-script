-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS public.nodes (
  id public.xid NOT NULL,
  name VARCHAR(120) NOT NULL,
  url VARCHAR(255) NOT NULL,
  method VARCHAR(10) NOT NULL,
  headers JSON NOT NULL,
  body JSON,
  request_mime VARCHAR(50) NOT NULL,
  response_mime VARCHAR(50) NOT NULL,
  response_direction TEXT NOT NULL,
  api_key VARCHAR(255) NOT NULL,
)
ALTER TABLE public.nodes
ADD CONSTRAINT node_pkey PRIMARY KEY (id, name);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE public.nodes;

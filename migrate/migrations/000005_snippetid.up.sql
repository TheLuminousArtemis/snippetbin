CREATE EXTENSION IF NOT EXISTS pgcrypto;
ALTER TABLE snippets ALTER COLUMN id DROP DEFAULT;

ALTER TABLE snippets
ALTER COLUMN id SET DEFAULT (
  (
    ('x' || substr(encode(gen_random_bytes(8), 'hex'), 1, 16))::bit(64)::bigint
  ) & 9223372036854775807
);
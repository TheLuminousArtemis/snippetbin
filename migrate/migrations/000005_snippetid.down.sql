ALTER TABLE snippets ALTER COLUMN id DROP DEFAULT;

-- Create a new sequence
CREATE SEQUENCE IF NOT EXISTS snippets_id_seq;

-- Attach it back to the id column
ALTER TABLE snippets ALTER COLUMN id SET DEFAULT nextval('snippets_id_seq');

-- Sync sequence to max(id) so it doesn't generate duplicates
SELECT setval('snippets_id_seq', (SELECT COALESCE(MAX(id), 1) FROM snippets));
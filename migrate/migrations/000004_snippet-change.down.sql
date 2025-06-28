ALTER TABLE snippets DROP COLUMN iv;
ALTER TABLE snippets ALTER COLUMN content TYPE TEXT USING content::text;
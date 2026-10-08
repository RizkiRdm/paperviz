-- 006_document_title.sql — Paper history titles (Chunk 2.1)

ALTER TABLE documents ADD COLUMN title TEXT NOT NULL DEFAULT '';
UPDATE documents SET title = substr(original_text, 1, 200) WHERE title = '';

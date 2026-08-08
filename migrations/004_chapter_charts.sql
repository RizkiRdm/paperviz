-- 004_chapter_charts.sql — link charts to chapters for tabbed view

ALTER TABLE charts ADD COLUMN chapter_id TEXT REFERENCES chapters(id) ON DELETE SET NULL;

CREATE INDEX idx_charts_chapter_id ON charts(chapter_id);

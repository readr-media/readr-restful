ALTER TABLE posts MODIFY COLUMN project_id bigint(20) unsigned DEFAULT 0;
UPDATE posts SET project_id = 0 WHERE project_id IS NULL;
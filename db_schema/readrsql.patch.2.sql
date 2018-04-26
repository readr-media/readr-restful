ALTER TABLE memos ADD COLUMN memo_order int DEFAULT 0;
UPDATE projects SET project_order=0 WHERE project_order IS NULL;
ALTER TABLE projects MODIFY COLUMN project_order int DEFAULT 0;
ALTER TABLE memos ADD COLUMN publish_status tinyint DEFAULT 0;
DROP INDEX author ON memos;
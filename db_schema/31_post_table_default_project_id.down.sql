ALTER TABLE posts MODIFY COLUMN project_id bigint(20) unsigned DEFAULT NULL;
UPDATE posts SET project_id = NULL WHERE project_id =0;
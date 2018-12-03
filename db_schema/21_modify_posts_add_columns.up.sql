ALTER TABLE posts ADD COLUMN project_id bigint(20) unsigned DEFAULT NULL;
ALTER TABLE posts ADD COLUMN post_order int DEFAULT NULL;
ALTER TABLE posts ADD COLUMN hero_image varchar(256) DEFAULT NULL AFTER content;
ALTER TABLE posts ADD COLUMN slug varchar(64) DEFAULT NULL AFTER content;
ALTER TABLE posts MODIFY COLUMN created_at datetime DEFAULT CURRENT_TIMESTAMP AFTER active;
ALTER TABLE posts MODIFY COLUMN link varchar(512) DEFAULT NULL AFTER published_at;
ALTER TABLE posts MODIFY COLUMN active int(11) DEFAULT 1 AFTER video_views;
UPDATE posts SET active=1 WHERE publish_status=2;
UPDATE posts SET active=2 WHERE publish_status=4, active=1;
UPDATE posts SET active=3 WHERE publish_status=1, active=1;
UPDATE posts SET active=4 WHERE publish_status=0, active=1;
ALTER TABLE posts ADD COLUMN publish_status tinyint DEFAULT 0;

# Create Author table for recording authors of posts
DROP TABLE IF EXISTS authors;
CREATE TABLE authors LIKE report_authors;
INSERT authors SELECT * FROM report_authors;

# Alter Author table columns
ALTER TABLE authors CHANGE `report_id` `resource_id` bigint(20) unsigned NOT NULL;
ALTER TABLE authors ADD COLUMN resource_type tinyint unsigned DEFAULT NULL;
ALTER TABLE authors ADD COLUMN author_type tinyint DEFAULT 0;
UPDATE authors SET resource_type = 4, author_type = 0;
DROP INDEX report_authors_reverse ON authors;
DROP INDEX report_authors ON authors;
CREATE INDEX resources ON authors(resource_type, resource_id);
CREATE INDEX resource_authors ON authors(author_id);
CREATE UNIQUE INDEX unique_author ON authors(resource_id, resource_type, author_id, author_type);


# Migrate post author and memo author to Author table
INSERT INTO authors 
	(resource_id, author_id, resource_type, author_type) SELECT 
	post_id, author, type, 0 FROM posts WHERE author IS NOT NULL;
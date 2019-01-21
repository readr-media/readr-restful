# Move the comments from post to their corresponsing memo and reports
ALTER TABLE comments MODIFY COLUMN resource varchar(512) DEFAULT NULL;
UPDATE comments INNER JOIN (
	SELECT comments.id, ref.link FROM (
		SELECT post_id, type, link, CONCAT('https://readr.tw/post/', post_id) AS post_link FROM posts WHERE type IN (4,5)
	) AS ref LEFT JOIN comments ON comments.resource = ref.post_link
) AS t ON comments.id = t.id
SET resource = link;


# Delete the old memo and report posts
DELETE FROM posts WHERE type in (4,5);


# select from memos and insert to posts
INSERT INTO posts 
	(video_views, comment_amount, title, content, slug, author, created_at, updated_at, updated_by, published_at, type, active, publish_status, project_id, post_order) 
	SELECT memos.memo_id, memos.comment_amount, memos.title, memos.content, projects.slug, memos.author, memos.created_at, memos.updated_at, memos.updated_by, memos.published_at, 5, memos.active, memos.publish_status, memos.project_id, memos.memo_order 
	FROM memos LEFT JOIN projects ON memos.project_id = projects.project_id;

# update comment resource for memos since resource url changed after migration
UPDATE comments INNER JOIN (
	SELECT CONCAT('https://readr.tw/series/', slug, '/', post_id) AS new_res, CONCAT('https://readr.tw/series/', slug, '/', video_views) AS old_res
	FROM posts WHERE type = 5) AS ref ON ref.old_res = comments.resource 
SET comments.resource = ref.new_res;

# update following table for memos, since we change the reference of target_id from memos table to posts table id
UPDATE following INNER JOIN (
	SELECT post_id AS new_res, video_views AS old_res
	FROM posts WHERE type = 5) AS ref ON ref.old_res = following.target_id AND following.type = 4 
SET following.target_id = ref.new_res;

# update post link for memos
UPDATE posts SET link = CONCAT('https://readr.tw/series/', slug, '/', post_id), video_views = NULL WHERE type = 5;



# select from reports and insert to posts
INSERT INTO posts 
	(video_views, like_amount, comment_amount, title, content, slug, hero_image, og_title, og_description, og_image, created_at, updated_at, updated_by, published_at, type, active, publish_status, project_id) SELECT 
	id, like_amount, comment_amount, title, description, slug, hero_image, og_title, og_description, og_image, created_at, updated_at, updated_by, published_at, 4, active, publish_status, project_id FROM reports;

# update report_authors for migrated report posts
UPDATE report_authors INNER JOIN (SELECT video_views AS report_id, post_id FROM posts WHERE type = 4) AS posts ON report_authors.report_id = posts.report_id SET report_authors.report_id = posts.post_id;

# update following table for reports, since we change the reference of target_id from reports table to posts table id
UPDATE following INNER JOIN (
	SELECT post_id AS new_res, video_views AS old_res
	FROM posts WHERE type = 4) AS ref ON ref.old_res = following.target_id AND following.type = 5 
SET following.target_id = ref.new_res;

# update post link for reports
UPDATE posts SET link = CONCAT('https://readr.tw/project/', slug), video_views = NULL WHERE type = 4;


SELECT reports.id, posts.post_id FROM reports LEFT JOIN posts ON reports.slug = posts.slug;

# Add column "subtitle" to posts table
ALTER TABLE posts ADD COLUMN `subtitle` varchar(256) DEFAULT NULL AFTER `title`;
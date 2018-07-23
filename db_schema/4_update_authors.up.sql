-- Change PRIMARY KEY of members to auto incremental new id. --
ALTER TABLE members DROP PRIMARY KEY;
ALTER TABLE members ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Upgrade author and updated_by in posts --
UPDATE posts p LEFT JOIN (SELECT id, member_id FROM members) m ON p.author = m.member_id LEFT JOIN (SELECT id, member_id FROM members) u ON p.updated_by = u.member_id SET p.author = m.id, p.updated_by = u.id;
ALTER TABLE posts MODIFY author BIGINT unsigned, MODIFY updated_by BIGINT unsigned;

-- Upgrade member_id to BIGINT unsigned in points --
UPDATE points p LEFT JOIN (SELECT id, member_id FROM members) m ON p.member_id = m.member_id LEFT JOIN (SELECT id, member_id FROM members) u ON p.updated_by = u.member_id SET p.member_id = m.id, p.updated_by = u.id;
ALTER TABLE points MODIFY member_id BIGINT unsigned NOT NULL, MODIFY updated_by BIGINT unsigned;

-- Upgrade author and updated_by in memos --
UPDATE memos m LEFT JOIN (SELECT id, member_id FROM members) a ON m.author = a.member_id LEFT JOIN (SELECT id, member_id FROM members) u ON m.updated_by = u.member_id SET m.author = a.id, m.updated_by = u.id;
ALTER TABLE memos MODIFY author BIGINT unsigned, MODIFY updated_by BIGINT unsigned;

-- Upgrade updated_by in tags --
UPDATE tags t LEFT JOIN (SELECT id, member_id FROM members) u ON t.updated_by = u.member_id SET t.updated_by = u.id;
ALTER TABLE tags MODIFY updated_by BIGINT unsigned;

-- Upgrade following_projects member_id to BIGINT unsigned --
UPDATE following_projects fp LEFT JOIN (SELECT id, member_id FROM members) m ON fp.member_id = m.member_id SET fp.member_id = m.id;
ALTER TABLE following_projects MODIFY member_id BIGINT unsigned NOT NULL;

-- Upgrade following_posts member_id to BIGINT unsigned --
UPDATE following_posts fp LEFT JOIN (SELECT id, member_id FROM members) m ON fp.member_id = m.member_id SET fp.member_id = m.id;
ALTER TABLE following_posts MODIFY member_id BIGINT unsigned NOT NULL;

-- Downgrade following_members member_id back to VARCHAR(48) --
UPDATE following_members fm LEFT JOIN (SELECT id, member_id FROM members) m ON fm.member_id = m.member_id LEFT JOIN (SELECT id, member_id FROM members) c ON fm.custom_editor = c.member_id SET fm.member_id = m.id, fm.custom_editor = c.id;
ALTER TABLE following_members MODIFY member_id BIGINT unsigned NOT NULL, MODIFY custom_editor BIGINT unsigned NOT NULL;

-- Changed points_id to id --
ALTER TABLE points CHANGE points_id id BIGINT UNSIGNED AUTO_INCREMENT;

-- Add id to table permissions --
ALTER TABLE permissions DROP PRIMARY KEY;
ALTER TABLE permissions ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Add auto incremental PRIMARY KEY id to post_tags --
ALTER TABLE post_tags DROP PRIMARY KEY;
ALTER TABLE post_tags ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;
CREATE INDEX post_tag ON post_tags(post_id, tag_id);
CREATE INDEX tag_post ON post_tags(tag_id, post_id);

-- Add reverse index to following_members -- 
ALTER TABLE following_members DROP PRIMARY KEY;
ALTER TABLE following_members ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;
CREATE INDEX member_follow ON following_members(member_id, custom_editor);
CREATE INDEX follow_member ON following_members(custom_editor, member_id);

-- Add reverse index to following_projects -- 
ALTER TABLE following_projects DROP PRIMARY KEY;
ALTER TABLE following_projects ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;
ALTER TABLE following_projects ADD COLUMN `join` INT DEFAULT 0;
CREATE INDEX member_follow ON following_projects(member_id, project_id);
CREATE INDEX follow_member ON following_projects(project_id, member_id);

-- Add reverse index to following_posts --
ALTER TABLE following_posts DROP PRIMARY KEY;
ALTER TABLE following_posts ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST; 
CREATE INDEX member_follow ON following_posts(member_id, post_id);
CREATE INDEX follow_member ON following_posts(post_id, member_id);

-- Add id to article_comments --
ALTER TABLE article_comments MODIFY post_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE article_comments DROP PRIMARY KEY;
ALTER TABLE article_comments ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Add id to project_comments --
ALTER TABLE project_comments MODIFY project_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE project_comments DROP PRIMARY KEY;
ALTER TABLE project_comments ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;
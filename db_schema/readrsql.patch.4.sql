-- Change PRIMARY KEY of members to auto incremental new id. --
ALTER TABLE members DROP PRIMARY KEY;
ALTER TABLE members ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Changed points_id to id --
ALTER TABLE points CHANGE points_id id BIGINT UNSIGNED AUTO_INCREMENT;

-- Add auto incremental PRIMARY KEY id to post_tags --
ALTER TABLE post_tags DROP PRIMARY KEY;
ALTER TABLE post_tags ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Add id to table permissions --
ALTER TABLE permissions DROP PRIMARY KEY;
ALTER TABLE permissions ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Add id to following_members -- 
ALTER TABLE following_members DROP PRIMARY KEY;
ALTER TABLE following_members MODIFY COLUMN member_id BIGINT UNSIGNED;
ALTER TABLE following_members MODIFY COLUMN custom_editor BIGINT UNSIGNED;
ALTER TABLE following_members ADD PRIMARY KEY(member_id,custom_editor);

-- Add id to following_projects -- 
ALTER TABLE following_projects DROP PRIMARY KEY;
ALTER TABLE following_projects MODIFY COLUMN member_id BIGINT UNSIGNED;
ALTER TABLE following_projects MODIFY COLUMN project_id BIGINT UNSIGNED;
ALTER TABLE following_projects ADD PRIMARY KEY(member_id,project_id);
ALTER TABLE following_projects ADD COLUMN `join` INT DEFAULT 0;

-- Add id to following_posts -- 
ALTER TABLE following_posts DROP PRIMARY KEY;
ALTER TABLE following_posts MODIFY COLUMN member_id BIGINT UNSIGNED;
ALTER TABLE following_posts MODIFY COLUMN post_id BIGINT UNSIGNED;
ALTER TABLE following_posts ADD PRIMARY KEY(member_id,post_id);

-- Add id to article_comments --
ALTER TABLE article_comments MODIFY post_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE article_comments DROP PRIMARY KEY;
ALTER TABLE article_comments ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;

-- Add id to project_comments --
ALTER TABLE project_comments MODIFY project_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE project_comments DROP PRIMARY KEY;
ALTER TABLE project_comments ADD COLUMN id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT FIRST;
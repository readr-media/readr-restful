-- Change utf8 --
SET NAMES utf8 COLLATE utf8_general_ci;

ALTER DATABASE `memberdb` CHARACTER SET = utf8 COLLATE = utf8_general_ci;

ALTER TABLE `article_comments` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;  
ALTER TABLE `comment_infos` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 
ALTER TABLE `following_members` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 
ALTER TABLE `following_posts` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 
ALTER TABLE `following_projects` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 
ALTER TABLE `members` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `memos` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `permissions` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `points` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `post_tags` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `posts` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 
ALTER TABLE `project_comments` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `projects` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `roles` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `tags` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 

-- Drop table project_authors --
DROP TABLE project_authors;

DROP TABLE reports;


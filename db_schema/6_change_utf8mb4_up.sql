-- Change utf8mb4 -- 
SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

ALTER DATABASE `memberdb` CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;

ALTER TABLE `article_comments` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;  
ALTER TABLE `comment_infos` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 
ALTER TABLE `following_members` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 
ALTER TABLE `following_posts` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 
ALTER TABLE `following_projects` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 
ALTER TABLE `members` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `memos` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `permissions` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `points` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `post_tags` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `posts` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 
ALTER TABLE `project_comments` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `projects` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `roles` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `tags` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 

-- Create table project_author --
-- await INDEX --
CREATE TABLE `project_authors` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `project_id` INT UNSIGNED NOT NULL,
    `author_id` INT UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `reports` (
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `slug` VARCHAR(64),
    `title` VARCHAR(256),
    `description` TEXT,
    `author` TEXT,
    `active` INT DEFAULT 1,
    `project_id` INT UNSIGNED,
    `created_at` DATETIME DEFAULT NOW(),
    `views` INT,
    `like_amount` INT,
    `comment_amount` INT,
    `og_title` VARCHAR(256),
    `og_description` VARCHAR(256),
    `og_image` VARCHAR(128),
    `updated_at` DATETIME DEFAULT NOW(),
    `updated_by` VARCHAR(48),
    `published_at` DATETIME,
    `publish_status` TINYINT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


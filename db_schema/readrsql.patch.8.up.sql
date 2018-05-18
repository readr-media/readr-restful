ALTER TABLE `comments` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE `comments_reported` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci; 

CREATE TABLE reports ( 
	`id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, 
	`created_at` DATETIME DEFAULT CURRENT_TIMESTAMP, 
	`like_amount` INT, 
	`comment_amount` INT, 
	`title` VARCHAR(256), 
	`description` TEXT, 
	`hero_image` VARCHAR(128), 
	`og_title` VARCHAR(256), 
	`og_description` VARCHAR(256), 
	`og_image` VARCHAR(128), 
	`active` TINYINT DEFAULT 1 , 
	`project_id` BIGINT UNSIGNED, 
	`updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP, 
	`updated_by` BIGINT UNSIGNED, 
	`published_at` DATETIME, 
	`slug` VARCHAR(64),
	`views` INT, 
	`publish_status` TINYINT, 
	INDEX (title), 
	INDEX (project_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE report_authors (
	`id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, 
	`report_id` BIGINT UNSIGNED NOT NULL, 
	`author_id` BIGINT UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE INDEX report_authors ON report_authors(report_id, author_id);
CREATE INDEX report_authors_reverse ON report_authors(author_id, report_id);
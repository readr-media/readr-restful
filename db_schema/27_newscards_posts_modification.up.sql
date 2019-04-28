# Modify posts table, add js and css columns
ALTER TABLE posts ADD COLUMN css text DEFAULT NULL;
ALTER TABLE posts ADD COLUMN javascript text DEFAULT NULL;

# Create card table
CREATE TABLE IF NOT EXISTS `newscards` (
	`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
	`post_id` bigint(20) unsigned NOT NULL,
	`title` varchar(256) NOT NULL,
	`description` text DEFAULT NULL,
	`background_image` varchar(256) DEFAULT NULL,
	`background_color` varchar(32) DEFAULT NULL,
	`image`	varchar(256) DEFAULT NULL,
	`video`	varchar(256) DEFAULT NULL,
	`created_at` datetime DEFAULT CURRENT_TIMESTAMP,
	`updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
	`order` int unsigned DEFAULT 0,
	`active` tinyint unsigned NOT NULL DEFAULT 0,
	`status` tinyint unsigned NOT NULL DEFAULT 0,
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
# Create promtion table
CREATE TABLE IF NOT EXISTS `promotions` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `status` tinyint unsigned NOT NULL DEFAULT 0,
  `active` tinyint unsigned NOT NULL DEFAULT 0,
  `title` varchar(256) NOT NULL DEFAULT '',
  `description` text DEFAULT NULL,
  `image` varchar(256) DEFAULT NULL,
  `link` varchar(256) DEFAULT NULL,
  `order` int unsigned DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `published_at` datetime DEFAULT NULL,
  `updated_by` bigint(20) unsigned DEFAULT NULL,
  PRIMARY KEY(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

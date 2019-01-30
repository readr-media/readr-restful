CREATE TABLE IF NOT EXISTS `assets` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `active` tinyint(4) DEFAULT 1,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP ,
  `created_by` bigint(20) unsigned DEFAULT NULL ,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_by` bigint(20) unsigned DEFAULT NULL ,
  `url` varchar(256)  DEFAULT NULL,
  `content_type` varchar(32) DEFAULT NULL,

  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
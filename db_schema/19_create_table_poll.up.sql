CREATE TABLE IF NOT EXISTS `polls` (
  `id`  bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `status` tinyint(3) unsigned NOT NULL DEFAULT 0,
  `active` tinyint(3) unsigned NOT NULL DEFAULT 0,
  `title` varchar(256) DEFAULT NULL,
  `description` text,
  `total_vote` int(11) DEFAULT 0,
  `frequency` varchar(16) DEFAULT NULL,
  `start_at` datetime DEFAULT NULL,
  `end_at`   datetime DEFAULT NULL,
  `max_choice` tinyint(3) unsigned NOT NULL DEFAULT 1,
  `changeable` tinyint(3) unsigned NOT NULL DEFAULT 0,
  `published_at` datetime,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `created_by` bigint(20) unsigned DEFAULT NULL,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_by` bigint(20) unsigned DEFAULT NULL,

  PRIMARY KEY (`id`),
  INDEX (created_by),
  INDEX (updated_by),

  FOREIGN KEY (created_by)
    REFERENCES members (id) ON UPDATE CASCADE ON DELETE SET NULL,
  FOREIGN KEY (updated_by)
    REFERENCES members (id) ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `polls_choices` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `choice` varchar(256) DEFAULT NULL,
  `total_vote` int(11) DEFAULT 0,
  `poll_id` bigint(20) unsigned NOT NULL,
  `active` tinyint(3) unsigned NOT NULL DEFAULT 0,
  `group_order` tinyint(3) unsigned NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `created_by` bigint(20) unsigned DEFAULT NULL,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_by` bigint(20) unsigned DEFAULT NULL,

  PRIMARY KEY (`id`),
  INDEX (poll_id),
  INDEX (created_by),
  INDEX (updated_by),

  FOREIGN KEY (poll_id)
    REFERENCES polls (id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (created_by)
    REFERENCES members (id) ON UPDATE CASCADE ON DELETE SET NULL,
  FOREIGN KEY (updated_by)
    REFERENCES members (id) ON UPDATE CASCADE ON DELETE SET NULL,
  UNIQUE KEY unique_order (poll_id, group_order)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `polls_chosen_choice` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `member_id` bigint(20) unsigned NOT NULL,
  `poll_id` bigint(20) unsigned NOT NULL,
  `choice_id` bigint(20) unsigned NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  INDEX (member_id),
  INDEX (poll_id),
  INDEX (choice_id),

  FOREIGN KEY (member_id)
    REFERENCES members(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (poll_id)
    REFERENCES polls(id) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (choice_id)
    REFERENCES polls_choices(id) ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


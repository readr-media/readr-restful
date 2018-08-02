CREATE TABLE `tagging` (
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `type` TINYINT(1) NOT NULL DEFAULT 0,
    `tag_id` BIGINT UNSIGNED NOT NULL,
    `target_id` BIGINT UNSIGNED NOT NULL,
    `created_at` DATETIME DEFAULT NOW(),
    CONSTRAINT type_member_target UNIQUE (type, target_id, tag_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO `tagging` (type, tag_id, target_id, created_at) SELECT 1, post_tags.tag_id, post_tags.post_id, now() FROM post_tags;
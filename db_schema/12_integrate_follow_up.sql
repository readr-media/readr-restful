ALTER TABLE report_authors ADD UNIQUE unique_author(author_id, report_id);

CREATE TABLE `following` (
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `type` TINYINT(1) NOT NULL DEFAULT 0,
    `member_id` BIGINT UNSIGNED NOT NULL,
    `target_id` BIGINT UNSIGNED NOT NULL,
    `emotion` TINYINT(1) NOT NULL DEFAULT 0,
    `created_at` DATETIME DEFAULT NOW(),
    CONSTRAINT type_member_target UNIQUE (type, member_id, target_id, emotion)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO `following` (type, member_id, target_id, created_at) SELECT 1, following_members.member_id, following_members.custom_editor, following_members.created_at FROM following_members;
INSERT IGNORE INTO `following` (type, member_id, target_id, created_at) SELECT 2, following_posts.member_id, following_posts.post_id, following_posts.created_at FROM following_posts;
INSERT IGNORE INTO `following` (type, member_id, target_id, created_at) SELECT 3, following_projects.member_id, following_projects.project_id, following_projects.created_at FROM following_projects;

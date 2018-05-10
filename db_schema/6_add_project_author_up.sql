-- Create table project_author --
CREATE TABLE `project_authors` (
    `id` INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `project_id` INT UNSIGNED NOT NULL,
    `author_id` INT UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- await INDEX --

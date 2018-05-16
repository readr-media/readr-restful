CREATE TABLE comments ( id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, body varchar(500), og_title varchar(256), og_description varchar(256), og_image varchar(128), like_amount int, parent_id bigint UNSIGNED, resource varchar(128), created_at datetime, updated_at datetime, author BIGINT UNSIGNED, active tinyint, status tinyint, ip int unsigned);

CREATE TABLE comments_reported ( id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, comment_id BIGINT UNSIGNED, reporter BIGINT UNSIGNED, reason varchar(500), solved tinyint, created_at datetime, updated_at datetime);

UPDATE members SET mail='' WHERE mail IS NULL;
ALTER TABLE members MODIFY COLUMN mail varchar(48) NOT NULL DEFAULT '';
ALTER TABLE points ADD COLUMN currency int(11) NOT NULL DEFAULT 0 AFTER points;
ALTER TABLE points ADD COLUMN status tinyint unsigned DEFAULT 0;
ALTER TABLE points ADD COLUMN member_name varchar(128) DEFAULT NULL AFTER member_id;
ALTER TABLE points ADD COLUMN member_mail varchar(128) DEFAULT NULL AFTER member_name;
ALTER TABLE points MODIFY COLUMN member_id varchar(48) DEFAULT NULL;
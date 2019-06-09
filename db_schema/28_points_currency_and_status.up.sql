ALTER TABLE points ADD COLUMN currency int(11) NOT NULL DEFAULT 0 AFTER points;
ALTER TABLE points ADD COLUMN status tinyint unsigned DEFAULT 0;
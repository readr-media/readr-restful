ALTER TABLE assets CHANGE `url` `destination` varchar(256) NOT NULL DEFAULT '';
ALTER TABLE assets CHANGE `content_type` `file_type` varchar(32) DEFAULT NULL;
ALTER TABLE assets ADD COLUMN file_name varchar(256) DEFAULT NULL;
ALTER TABLE assets ADD COLUMN file_extension varchar(32) DEFAULT NULL;
ALTER TABLE assets ADD COLUMN asset_type tinyint(1) DEFAULT NULL;
ALTER TABLE assets ADD COLUMN title varchar(256) DEFAULT NULL;
ALTER TABLE assets ADD COLUMN copyright tinyint(1) NOT NULL DEFAULT 0;

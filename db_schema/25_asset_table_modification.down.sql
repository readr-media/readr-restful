ALTER TABLE assets CHANGE `destination` `url` varchar(256) DEFAULT NULL;
ALTER TABLE assets CHANGE `file_type` `content_type` varchar(32) DEFAULT NULL;
ALTER TABLE assets DROP COLUMN file_name;
ALTER TABLE assets DROP COLUMN file_extension;
ALTER TABLE assets DROP COLUMN asset_type;
ALTER TABLE assets DROP COLUMN title;
ALTER TABLE assets DROP COLUMN copyright;
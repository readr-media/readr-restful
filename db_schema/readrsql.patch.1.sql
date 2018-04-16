-- Add Colimn uuid to Members table --
ALTER TABLE `members` ADD COLUMN uuid VARCHAR(36) UNIQUE;
UPDATE `members` set `uuid` = uuid();
ALTER TABLE `members` MODIFY `uuid` VARCHAR(36) NOT NULL UNIQUE;

-- Add Column publish_status to Posts table -- 
ALTER TABLE projects ADD COLUMN publish_status tinyint NOT NULL DEFAULT 0;

-- Extends content length for member.profile_image, posts.link_image and posts.link --
Alter table members modify column profile_image text;
Alter table posts modify column link_image text;
Alter table posts modify column link varchar(512);

-- Add Column progress and memo_points to Projects table --
ALTER TABLE projects ADD COLUMN progress FLOAT DEFAULT 0;
ALTER TABLE projects ADD COLUMN memo_points int(8);
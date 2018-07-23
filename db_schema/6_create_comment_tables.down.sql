DROP TABLE comments;
DROP TABLE comments_reported;

ALTER TABLE members MODIFY COLUMN mail varchar(48) DEFAULT '';
UPDATE members SET mail = NULL WHERE mail='';
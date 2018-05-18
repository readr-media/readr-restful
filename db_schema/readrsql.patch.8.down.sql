ALTER TABLE `comments` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE `comments_reported` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci; 

DROP TABLE reports;
DROP TABLE report_authors;
-- Delete entry with repeat report_id and author_id
DELETE r1 FROM report_authors r1, report_authors r2 WHERE r1.id > r2.id AND r1.report_id <=> r2.report_id AND r1.author_id <=> r2.author_id;

ALTER TABLE report_authors DROP INDEX report_authors;
ALTER TABLE report_authors ADD UNIQUE report_authors(author_id, report_id);

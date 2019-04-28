# Modify posts table, remove js and css columns
ALTER TABLE posts DROP COLUMN css;
ALTER TABLE posts DROP COLUMN javascript;

# Drop card table
DROP TABLE newscards;
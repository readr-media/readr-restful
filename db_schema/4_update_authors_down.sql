-- Downgrade author and updated_by in posts --
ALTER TABLE posts MODIFY author VARCHAR(48), MODIFY updated_by VARCHAR(48);
UPDATE posts p LEFT JOIN (SELECT id, member_id FROM members) m ON p.author = m.id LEFT JOIN (SELECT id, member_id FROM members) u ON p.updated_by = u.id SET p.author = m.member_id, p.updated_by = u.member_id;

-- Downgrade member_id back to VARCHAR(48) --
ALTER TABLE points MODIFY member_id VARCHAR(48) NOT NULL, MODIFY updated_by VARCHAR(48);
UPDATE points p LEFT JOIN (SELECT id, member_id FROM members) m ON p.member_id = m.id LEFT JOIN (SELECT id, member_id FROM members) u ON p.updated_by = u.id SET p.member_id = m.member_id, p.updated_by = u.member_id;

-- Downgrade memos member_id back to VARCHAR(48) --
ALTER TABLE memos MODIFY author VARCHAR(48), MODIFY updated_by VARCHAR(48);
UPDATE memos m LEFT JOIN (SELECT id, member_id FROM members) a ON m.author = a.id LEFT JOIN (SELECT id, member_id FROM members) u ON m.updated_by = u.id SET m.author = a.member_id, m.updated_by = u.member_id;

-- Downgrade following_projects member_id back to VARCHAR(48) --
ALTER TABLE following_projects MODIFY member_id VARCHAR(48) NOT NULL;
UPDATE following_projects fp LEFT JOIN (SELECT id, member_id FROM members) m ON fp.member_id = m.id SET fp.member_id = m.member_id;

-- Downgrade following_posts member_id back to VARCHAR(48) --
ALTER TABLE following_posts MODIFY member_id VARCHAR(48) NOT NULL;
UPDATE following_posts fp LEFT JOIN (SELECT id, member_id FROM members) m ON fp.member_id = m.id SET fp.member_id = m.member_id;

-- Downgrade following_members member_id back to VARCHAR(48) --
ALTER TABLE following_members MODIFY member_id VARCHAR(48) NOT NULL, MODIFY custom_editor VARCHAR(48) NOT NULL;
UPDATE following_members fm LEFT JOIN (SELECT id, member_id FROM members) m ON fm.member_id = m.id LEFT JOIN (SELECT id, member_id FROM members) c ON fm.custom_editor = c.id SET fm.member_id = m.member_id, fm.custom_editor = c.member_id;

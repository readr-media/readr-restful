-- Upgrade author and updated_by in posts --
UPDATE posts p LEFT JOIN (SELECT id, member_id FROM members) m ON p.author = m.member_id LEFT JOIN (SELECT id, member_id FROM members) u ON p.updated_by = u.member_id SET p.author = m.id, p.updated_by = u.id;
ALTER TABLE posts MODIFY author BIGINT unsigned, MODIFY updated_by BIGINT unsigned;

-- Upgrade member_id to BIGINT unsigned in points --
UPDATE points p LEFT JOIN (SELECT id, member_id FROM members) m ON p.member_id = m.member_id LEFT JOIN (SELECT id, member_id FROM members) u ON p.updated_by = u.member_id SET p.member_id = m.id, p.updated_by = u.id;
ALTER TABLE points MODIFY member_id BIGINT unsigned NOT NULL, MODIFY updated_by BIGINT unsigned;

-- Upgrade author and updated_by in memos --
UPDATE memos m LEFT JOIN (SELECT id, member_id FROM members) a ON m.author = a.member_id LEFT JOIN (SELECT id, member_id FROM members) u ON m.updated_by = u.member_id SET m.author = a.id, m.updated_by = u.id;
ALTER TABLE memos MODIFY author BIGINT unsigned, MODIFY updated_by BIGINT unsigned;

-- Upgrade following_projects member_id to BIGINT unsigned --
UPDATE following_projects fp LEFT JOIN (SELECT id, member_id FROM members) m ON fp.member_id = m.member_id SET fp.member_id = m.id;
ALTER TABLE following_projects MODIFY member_id BIGINT unsigned NOT NULL;

-- Upgrade following_posts member_id to BIGINT unsigned --
UPDATE following_posts fp LEFT JOIN (SELECT id, member_id FROM members) m ON fp.member_id = m.member_id SET fp.member_id = m.id;
ALTER TABLE following_posts MODIFY member_id BIGINT unsigned NOT NULL;

-- Downgrade following_members member_id back to VARCHAR(48) --
UPDATE following_members fm LEFT JOIN (SELECT id, member_id FROM members) m ON fm.member_id = m.member_id LEFT JOIN (SELECT id, member_id FROM members) c ON fm.custom_editor = c.member_id SET fm.member_id = m.id, fm.custom_editor = c.id;
ALTER TABLE following_members MODIFY member_id BIGINT unsigned NOT NULL, MODIFY custom_editor BIGINT unsigned NOT NULL;
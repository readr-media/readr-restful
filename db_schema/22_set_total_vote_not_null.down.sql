ALTER TABLE polls
MODIFY total_vote int(11) DEFAULT 0,
MODIFY updated_at datetime DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE polls_choices
MODIFY total_vote int(11) DEFAULT 0,
MODIFY updated_at datetime DEFAULT CURRENT_TIMESTAMP;


UPDATE polls SET frequency = '0';
ALTER TABLE polls MODIFY `frequency` tinyint(3) NOT NULL DEFAULT 0;
ALTER TABLE polls_chosen_choice ADD COLUMN `active` tinyint(3) NOT NULL DEFAULT 0 AFTER `choice_id`;
ALTER TABLE polls_chosen_choice MODIFY `created_at` datetime NOT NULL DEFAULT NOW();

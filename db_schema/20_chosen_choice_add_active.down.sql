ALTER TABLE polls MODIFY `frequency` varchar(16) DEFAULT NULL;
ALTER TABLE polls_chosen_choice DROP COLUMN `active`;
ALTER TABLE polls_chosen_choice MODIFY `created_at` datetime DEFAULT NOW();

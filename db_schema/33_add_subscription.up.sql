-- Create new table subscriptions
CREATE TABLE IF NOT EXISTS `subscriptions` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `member_id` bigint(20) unsigned DEFAULT NULL,
    `email` varchar(32) NOT NULL DEFAULT '',
    `amount` int(11) NOT NULL,
    `payment_service` varchar(32) NOT NULL DEFAULT '',
    `invoice_service` varchar(32) NOT NULL DEFAULT '',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
    `last_paid_at` datetime,
    `status` int(11) NOT NULL DEFAULT 0,
    `payment_infos` JSON,
    `invoice_infos` JSON,
    
    PRIMARY KEY(`id`),
    INDEX (member_id),

    FOREIGN KEY (member_id)
        REFERENCES members (id) ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

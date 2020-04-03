-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `billing_transactions` ADD `checked_at` TIMESTAMP DEFAULT NOW();
ALTER TABLE `billing_transactions` ADD `is_locked` TINYINT(1) DEFAULT 0;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `billing_transactions` DROP COLUMN `checked_at`;
ALTER TABLE `billing_transactions` DROP COLUMN `is_locked`;
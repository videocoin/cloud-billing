-- +goose Up
-- SQL in this section is executed when the migration is applied.
INSERT INTO `billing_accounts` (`id`, `user_id`, `email`) VALUES ('bank', 0, 'bank@videocoin.net');

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DELETE FROM `billing_accounts` WHERE id = 'bank';
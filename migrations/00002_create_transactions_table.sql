-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE billing_transactions (
    `id`                  VARCHAR(255) PRIMARY KEY NOT NULL,
    `account_id`          VARCHAR(255) NOT NULL,
    `created_at`          TIMESTAMP DEFAULT NOW(),
    `type`                VARCHAR(50) NOT NULL,
    `checkout_session_id` VARCHAR(255),
    `payment_intent_id`   VARCHAR(255),
    `payment_status`      VARCHAR(255),
    `amount`              INT(20) NOT NULL,
    INDEX `account_id_idx` (`account_id`),
    FOREIGN KEY (`account_id`) REFERENCES `billing_accounts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE billing_transactions;
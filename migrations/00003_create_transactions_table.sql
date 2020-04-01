-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE billing_transactions (
    `id`         VARCHAR(255) PRIMARY KEY NOT NULL,
    `from`       VARCHAR(255) NOT NULL,
    `to`         VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP DEFAULT NOW(),
    `amount`     INT(20) NOT NULL,
    `status`     VARCHAR(255) NOT NULL,

    `payment_intent_secret` VARCHAR(255) DEFAULT NULL,
    `payment_intent_id`     VARCHAR(255) DEFAULT NULL,
    `payment_status`        VARCHAR(255) DEFAULT NULL,

    `stream_id` VARCHAR(255) DEFAULT NULL,
    `profile_id` VARCHAR(255) DEFAULT NULL,

    INDEX `from_idx` (`from`),
    INDEX `to_idx` (`to`),
    FOREIGN KEY (`from`) REFERENCES `billing_accounts` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`to`) REFERENCES `billing_accounts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE billing_transactions;
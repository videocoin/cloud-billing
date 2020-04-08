-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE billing_transactions (
    `id`         VARCHAR(255) PRIMARY KEY NOT NULL,
    `from`       VARCHAR(255) NOT NULL,
    `to`         VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP DEFAULT NOW(),
    `amount`     DECIMAL(10,4) NOT NULL,
    `status`     VARCHAR(255) NOT NULL,

    `checked_at` TIMESTAMP DEFAULT NOW(),
    `is_locked`  TINYINT(1) DEFAULT 0,

    `payment_intent_secret` VARCHAR(255) DEFAULT NULL,
    `payment_intent_id`     VARCHAR(255) DEFAULT NULL,
    `payment_status`        VARCHAR(255) DEFAULT NULL,

    `stream_id`               VARCHAR(255) DEFAULT NULL,
    `stream_name`             VARCHAR(255) DEFAULT NULL,
    `stream_contract_address` VARCHAR(255) DEFAULT NULL,
    `stream_is_live`          TINYINT(1) DEFAULT 0,

    `profile_id`              VARCHAR(255) DEFAULT NULL,
    `profile_name`            VARCHAR(255) DEFAULT NULL,
    `profile_cost`            DECIMAL(10,4) DEFAULT NULL,

    `task_id`                 VARCHAR(255) DEFAULT NULL,
    `chunk_num`               INT DEFAULT NULL,
    `duration`                INT DEFAULT NULL,
    `price`                   DECIMAL(10,4) DEFAULT NULL,

    INDEX `from_idx` (`from`),
    INDEX `to_idx` (`to`),
    INDEX `sca_cn_idx` (`stream_contract_address`, `chunk_num`),

    FOREIGN KEY (`from`) REFERENCES `billing_accounts` (`id`) ON DELETE CASCADE,
    FOREIGN KEY (`to`) REFERENCES `billing_accounts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE billing_transactions;
-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS `billing_accounts` (
  `id`          VARCHAR(255) NOT NULL,
  `user_id`     VARCHAR(255) NOT NULL,
  `email`       VARCHAR(255) NOT NULL,
  `created_at`  TIMESTAMP DEFAULT NOW(),
  `updated_at`  TIMESTAMP NULL DEFAULT NULL,
  `balance`     INT(20) DEFAULT 0,
  `customer_id` VARCHAR(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE billing_accounts;
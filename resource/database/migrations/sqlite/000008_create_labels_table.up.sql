CREATE TABLE IF NOT EXISTS `labels`
(
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `namespace`  VARCHAR(255) NOT NULL,
    `name`       VARCHAR(255) NOT NULL,
    `kind`       VARCHAR(255) NOT NULL,
    `key`        VARCHAR(255) NOT NULL,
    `value`      VARCHAR(255),

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);

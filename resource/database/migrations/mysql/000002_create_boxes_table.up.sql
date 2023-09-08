CREATE TABLE IF NOT EXISTS `boxes`
(
    `id`         INTEGER AUTO_INCREMENT,
    `namespace`  VARCHAR(255) NOT NULL,
    `name`       VARCHAR(255) NOT NULL,
    `data`       TEXT,
    `enabled`    TINYINT,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
CREATE TABLE IF NOT EXISTS `builds`
(
    `id`         INTEGER AUTO_INCREMENT,
    `box_id`     INTEGER      NOT NULL,
    `number`     INTEGER      NOT NULL,
    `phase`      VARCHAR(50)  NOT NULL,
    `title`      VARCHAR(255) NOT NULL,
    `settings`   TEXT,
    `started`    INTEGER,
    `stopped`    INTEGER,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
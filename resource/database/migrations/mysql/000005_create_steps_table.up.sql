CREATE TABLE IF NOT EXISTS `steps`
(
    `id`         INTEGER AUTO_INCREMENT,
    `stage_id`   INTEGER      NOT NULL,
    `number`     INTEGER      NOT NULL,
    `phase`      VARCHAR(50)  NOT NULL,
    `name`       VARCHAR(255) NOT NULL,
    `started`    INTEGER,
    `stopped`    INTEGER,
    `exit_code`  INTEGER,
    `error`      VARCHAR(1000),

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;
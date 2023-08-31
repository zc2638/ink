CREATE TABLE IF NOT EXISTS `step_statuses`
(
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `stage_id`   INTEGER      NOT NULL,
    `number`     INTEGER      NOT NULL,
    `phase`      VARCHAR(50)  NOT NULL,
    `name`       VARCHAR(255) NOT NULL,
    `started`    INTEGER,
    `stopped`    INTEGER,
    `exit_code`  INTEGER,
    `error`      VARCHAR(1000),

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);
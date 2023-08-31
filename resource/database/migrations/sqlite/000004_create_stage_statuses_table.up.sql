CREATE TABLE IF NOT EXISTS `stage_statuses`
(
    `id`          INTEGER PRIMARY KEY AUTOINCREMENT,
    `box_id`      INTEGER      NOT NULL,
    `build_id`    INTEGER      NOT NULL,
    `number`      INTEGER      NOT NULL,
    `phase`       VARCHAR(50)  NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `worker_name` VARCHAR(255),
    `worker`      TEXT,
    `started`     INTEGER,
    `stopped`     INTEGER,
    `error`       VARCHAR(1000),

    `created_at`  DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME DEFAULT CURRENT_TIMESTAMP
);
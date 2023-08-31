CREATE TABLE IF NOT EXISTS `builds`
(
    `id`         INTEGER PRIMARY KEY AUTOINCREMENT,
    `box_id`     INTEGER      NOT NULL,
    `number`     INTEGER      NOT NULL,
    `phase`      VARCHAR(50)  NOT NULL,
    `title`      VARCHAR(255) NOT NULL,
    `started`    INTEGER,
    `stopped`    INTEGER,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);
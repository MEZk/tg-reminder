-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id          INTEGER PRIMARY KEY,
    name        TEXT      NOT NULL,
    status      TEXT      NOT NULL,
    created_at  TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reminders
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER   NOT NULL,
    chat_id       INTEGER   NOT NULL,
    text          TEXT      NOT NULL,
    remind_at     TIMESTAMP NOT NULL,
    status        TEXT      NOT NULL,
    attempts_left INTEGER   NOT NULL,
    created_at    TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at   TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bot_states
(
    user_id     INTEGER PRIMARY KEY,
    name        TEXT      NOT NULL,
    context     BLOB,
    modified_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
DROP TABLE reminders;

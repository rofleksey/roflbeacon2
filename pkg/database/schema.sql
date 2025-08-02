CREATE TABLE IF NOT EXISTS account
(
    id                 BIGSERIAL PRIMARY KEY,
    token              VARCHAR(255) NOT NULL,
    name               VARCHAR(255) NOT NULL UNIQUE,
    chat_id            BIGINT,
    status             JSONB        NOT NULL
);

CREATE TABLE IF NOT EXISTS updates
(
    id         BIGSERIAL PRIMARY KEY,
    account_id BIGINT    NOT NULL,
    created    TIMESTAMP NOT NULL,
    data       JSONB     NOT NULL,
    CONSTRAINT fk_updates_account FOREIGN KEY (account_id) REFERENCES account (id)
);

CREATE TABLE IF NOT EXISTS fence
(
    id        BIGSERIAL PRIMARY KEY,
    name      VARCHAR(255)     NOT NULL UNIQUE,
    longitude DOUBLE PRECISION NOT NULL,
    latitude  DOUBLE PRECISION NOT NULL,
    radius    DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS migration
(
    id      VARCHAR(255) PRIMARY KEY,
    applied TIMESTAMP NOT NULL
);

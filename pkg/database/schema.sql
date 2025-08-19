CREATE TABLE IF NOT EXISTS account
(
    id                 BIGSERIAL PRIMARY KEY,
    token              VARCHAR(255) NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL UNIQUE,
    chat_id            BIGINT,
    status             JSONB        NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_account_chat_id ON account (chat_id);

CREATE TABLE IF NOT EXISTS updates
(
    id         BIGSERIAL PRIMARY KEY,
    account_id BIGINT    NOT NULL,
    created    TIMESTAMP NOT NULL,
    data       JSONB     NOT NULL,
    CONSTRAINT fk_updates_account FOREIGN KEY (account_id) REFERENCES account (id)
);
CREATE INDEX IF NOT EXISTS idx_updates_account_id ON updates (account_id);
CREATE INDEX IF NOT EXISTS idx_updates_account_id_created_desc ON updates (account_id, created DESC);
CREATE INDEX IF NOT EXISTS idx_updates_created ON updates (created);

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

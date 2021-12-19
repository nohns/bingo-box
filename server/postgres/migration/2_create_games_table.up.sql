CREATE TABLE IF NOT EXISTS games (
    id uuid DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    host_id uuid NOT NULL,
    next_card_number INT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (host_id) REFERENCES users (id)
);

CREATE INDEX games_host_idx ON games (host_id);
CREATE TABLE IF NOT EXISTS game_cards (
    id uuid DEFAULT gen_random_uuid(),
    game_id uuid NOT NULL,
    number INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (game_id) REFERENCES games (id)
);

CREATE INDEX game_cards_game_idx ON game_cards (game_id);
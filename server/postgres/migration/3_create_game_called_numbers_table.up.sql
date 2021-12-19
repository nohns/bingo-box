CREATE TABLE IF NOT EXISTS game_called_numbers (
    id uuid DEFAULT gen_random_uuid(),
    game_id uuid NOT NULL,
    number INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (game_id) REFERENCES games (id)
);

CREATE INDEX game_called_numbers_game_idx ON game_called_numbers (game_id);
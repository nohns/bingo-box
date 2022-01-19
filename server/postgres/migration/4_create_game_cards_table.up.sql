CREATE TABLE IF NOT EXISTS game_cards (
    id uuid DEFAULT gen_random_uuid(),
    game_id uuid NOT NULL,
    number INT NOT NULL,
    player_id INT,
    PRIMARY KEY (id),
    FOREIGN KEY (game_id) REFERENCES games (id),
    FOREIGN KEY (player_id) REFERENCES players (id)
);

CREATE UNIQUE INDEX game_cards_game_id_numberx ON game_cards (game_id,number);
CREATE INDEX game_cards_game_idx ON game_cards (game_id);
CREATE TABLE IF NOT EXISTS game_card_spots (
    id uuid DEFAULT gen_random_uuid(),
    game_card_id uuid NOT NULL,
    row INT NOT NULL,
    col INT NOT NULL,
    number INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (game_card_id) REFERENCES game_cards (id)
);

CREATE INDEX game_card_spots_card_idx ON game_card_spots (card_id);
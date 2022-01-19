CREATE TABLE IF NOT EXISTS players (
    id uuid DEFAULT gen_random_uuid(),
    game_invitation_id uuid NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    card_amount INT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (game_invitation_id) REFERENCES game_invitations (id)
);

CREATE UNIQUE INDEX players_game_invitation_id_players_emailx ON game_invitations(game_invitation_id,email)

CREATE INDEX players_game_invitation_idx ON game_invitations(game_invitation_id);
CREATE TYPE INVITATION_DELIVERY_METHOD AS ENUM ('EMAIL', 'DOWNLOAD');

CREATE TABLE IF NOT EXISTS game_invitations (
    id uuid DEFAULT gen_random_uuid(),
    game_id uuid NOT NULL,
    delivery_method INVITATION_DELIVERY_METHOD NOT NULL,
    max_card_amount INT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (game_id) REFERENCES games (id)
);

CREATE UNIQUE INDEX __idx ON game_invitations(a,c,d)

CREATE INDEX game_invitations_game_idx ON game_invitations(game_id);
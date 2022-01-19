CREATE TYPE INVITATION_CRITERIUM_KIND AS ENUM ('REGEX');

CREATE TABLE IF NOT EXISTS game_invitation_criteria (
    id uuid DEFAULT gen_random_uuid(),
    game_invitation_id uuid NOT NULL,
    kind INVITATION_CRITERIUM_KIND NOT NULL,
    field VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (game_invitation_id) REFERENCES game_invitations (id)
);

CREATE INDEX game_invitation_criteria_game_invitation_idx ON game_invitation_criteria (game_invitation_id);
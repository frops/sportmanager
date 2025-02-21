-- +migrate Up
CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    location VARCHAR(255),
    venue_name VARCHAR(255) DEFAULT 'Nova Sports Soccer Field',
    map_link VARCHAR(255),
    min_players INTEGER DEFAULT 10,
    max_players INTEGER DEFAULT 12,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS match_players (
    match_id INTEGER REFERENCES matches(id),
    player_id INTEGER REFERENCES players(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (match_id, player_id)
);

-- +migrate Down
DROP TABLE IF EXISTS match_players;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS matches; 
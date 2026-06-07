#!/bin/bash
set -e

    export PGPASSWORD="$POSTGRES_PASSWORD"

    echo "!!! Creating tables in  ${DB_NAME}..."

    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL

CREATE TABLE IF NOT EXISTS users (
    uuid uuid PRIMARY KEY DEFAULT uuidv7(),
    username text NOT NULL UNIQUE,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rooms (
    uuid uuid PRIMARY KEY DEFAULT uuidv7(),
    name text NOT NULL,
    room_description text,
    owner_uuid uuid NOT NULL,
    FOREIGN KEY(owner_uuid) REFERENCES users(uuid) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rooms_users_relations (
    room_uuid uuid NOT NULL,
    FOREIGN KEY(room_uuid) REFERENCES rooms(uuid) ON DELETE CASCADE,
    member_uuid uuid NOT NULL,
    FOREIGN KEY(member_uuid) REFERENCES users(uuid) ON DELETE CASCADE,
    PRIMARY KEY (member_uuid, room_uuid),
    joined_at timestamptz NOT NULL DEFAULT now()
);

CREATE DOMAIN special_stat AS SMALLINT CHECK (VALUE BETWEEN 0 AND 10);

CREATE TABLE IF NOT EXISTS characters (
    char_uuid uuid PRIMARY KEY DEFAULT uuidv7(),
    name text NOT NULL,
    character_description text,
    owner_uuid uuid NOT NULL,
    FOREIGN KEY(owner_uuid) REFERENCES users(uuid) ON DELETE CASCADE,
    strength     special_stat NOT NULL DEFAULT 0,
    perception   special_stat NOT NULL DEFAULT 0,
    endurance    special_stat NOT NULL DEFAULT 0,
    charisma     special_stat NOT NULL DEFAULT 0,
    intelligence special_stat NOT NULL DEFAULT 0,
    agility      special_stat NOT NULL DEFAULT 0,
    luck         special_stat NOT NULL DEFAULT 0,
    level SMALLINT NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now()
);

EOSQL

echo "!!! All tables successfully created"

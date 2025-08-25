-- 000001_create_rooms_and_members.up.sql

CREATE TABLE rooms (
                       id BIGSERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       deleted_at TIMESTAMPTZ
);

CREATE TABLE room_members (
                              id BIGSERIAL PRIMARY KEY,
                              room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
                              user_id BIGINT NOT NULL,
                              joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

-- Индексы
CREATE INDEX idx_room_members_room_id ON room_members(room_id);
CREATE INDEX idx_room_members_user_id ON room_members(user_id);

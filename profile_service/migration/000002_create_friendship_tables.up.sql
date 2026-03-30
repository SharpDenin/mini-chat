-- +migrate Up

-- Таблица запросов в друзья
CREATE TABLE IF NOT EXISTS friend_requests (
                                               id BIGSERIAL PRIMARY KEY,
                                               sender_id BIGINT NOT NULL,
                                               receiver_id BIGINT NOT NULL,
                                               status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
                             );

-- Индексы для friend_requests
CREATE INDEX IF NOT EXISTS idx_friend_requests_sender ON friend_requests(sender_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver ON friend_requests(receiver_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_status ON friend_requests(status);
CREATE INDEX IF NOT EXISTS idx_friend_requests_deleted_at ON friend_requests(deleted_at);

-- Уникальный индекс для активных запросов
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_pending_request
    ON friend_requests(sender_id, receiver_id)
    WHERE status = 'pending' AND deleted_at IS NULL;

-- Таблица друзей
CREATE TABLE IF NOT EXISTS friends (
                                       id BIGSERIAL PRIMARY KEY,
                                       user_id BIGINT NOT NULL,
                                       friend_id BIGINT NOT NULL,
                                       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
                             );

-- Индексы для friends
CREATE INDEX IF NOT EXISTS idx_friends_user ON friends(user_id);
CREATE INDEX IF NOT EXISTS idx_friends_friend ON friends(friend_id);
CREATE INDEX IF NOT EXISTS idx_friends_deleted_at ON friends(deleted_at);

-- Уникальный индекс для пары друзей
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_friends_pair
    ON friends(user_id, friend_id)
    WHERE deleted_at IS NULL;

-- Таблица заблокированных пользователей
CREATE TABLE IF NOT EXISTS blocked_users (
                                             id BIGSERIAL PRIMARY KEY,
                                             blocker_id BIGINT NOT NULL,
                                             blocked_id BIGINT NOT NULL,
                                             reason TEXT,
                                             created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
                             );

-- Индексы для blocked_users
CREATE INDEX IF NOT EXISTS idx_blocked_users_blocker ON blocked_users(blocker_id);
CREATE INDEX IF NOT EXISTS idx_blocked_users_blocked ON blocked_users(blocked_id);
CREATE INDEX IF NOT EXISTS idx_blocked_users_deleted_at ON blocked_users(deleted_at);

-- Уникальный индекс для блокировок
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_block_pair
    ON blocked_users(blocker_id, blocked_id)
    WHERE deleted_at IS NULL;

-- Таблица истории дружбы (для Kafka и аналитики)
CREATE TABLE IF NOT EXISTS friendship_history (
                                                  id BIGSERIAL PRIMARY KEY,
                                                  event_type VARCHAR(50) NOT NULL,
    user_id BIGINT NOT NULL,
    target_id BIGINT NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    metadata JSONB,
    request_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Индексы для friendship_history
CREATE INDEX IF NOT EXISTS idx_friendship_history_user ON friendship_history(user_id);
CREATE INDEX IF NOT EXISTS idx_friendship_history_target ON friendship_history(target_id);
CREATE INDEX IF NOT EXISTS idx_friendship_history_created ON friendship_history(created_at);
CREATE INDEX IF NOT EXISTS idx_friendship_history_event ON friendship_history(event_type);
CREATE INDEX IF NOT EXISTS idx_friendship_history_request ON friendship_history(request_id);

-- Комментарии к таблицам
COMMENT ON TABLE friend_requests IS 'Запросы в друзья';
COMMENT ON TABLE friends IS 'Дружеские связи между пользователями';
COMMENT ON TABLE blocked_users IS 'Заблокированные пользователи';
COMMENT ON TABLE friendship_history IS 'История всех событий дружбы для аналитики';

-- Комментарии к полям
COMMENT ON COLUMN friend_requests.status IS 'pending, accepted, rejected, cancelled';
COMMENT ON COLUMN friendship_history.event_type IS 'request_sent, request_accepted, request_rejected, request_cancelled, unfriended, blocked, unblocked';
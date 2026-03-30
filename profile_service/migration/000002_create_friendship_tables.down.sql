-- +migrate Down

-- Удаляем индексы
DROP INDEX IF EXISTS idx_unique_pending_request;
DROP INDEX IF EXISTS idx_friend_requests_sender;
DROP INDEX IF EXISTS idx_friend_requests_receiver;
DROP INDEX IF EXISTS idx_friend_requests_status;
DROP INDEX IF EXISTS idx_friend_requests_deleted_at;

DROP INDEX IF EXISTS idx_unique_friends_pair;
DROP INDEX IF EXISTS idx_friends_user;
DROP INDEX IF EXISTS idx_friends_friend;
DROP INDEX IF EXISTS idx_friends_deleted_at;

DROP INDEX IF EXISTS idx_unique_block_pair;
DROP INDEX IF EXISTS idx_blocked_users_blocker;
DROP INDEX IF EXISTS idx_blocked_users_blocked;
DROP INDEX IF EXISTS idx_blocked_users_deleted_at;

DROP INDEX IF EXISTS idx_friendship_history_user;
DROP INDEX IF EXISTS idx_friendship_history_target;
DROP INDEX IF EXISTS idx_friendship_history_created;
DROP INDEX IF EXISTS idx_friendship_history_event;
DROP INDEX IF EXISTS idx_friendship_history_request;

-- Удаляем таблицы
DROP TABLE IF EXISTS friendship_history;
DROP TABLE IF EXISTS blocked_users;
DROP TABLE IF EXISTS friends;
DROP TABLE IF EXISTS friend_requests;
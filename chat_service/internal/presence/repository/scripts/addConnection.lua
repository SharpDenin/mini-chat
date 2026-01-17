-- KEYS
-- 1 = connKey
-- 2 = userConnSet

-- ARGV
-- 1 = userId
-- 2 = connId
-- 3 = device
-- 4 = nowMs
-- 5 = ttlSec

redis.call('HSET', KEYS[1],
  'user_id', ARGV[1],
  'device', ARGV[3],
  'connected_at', ARGV[4],
  'last_activity', ARGV[4]
)

redis.call('EXPIRE', KEYS[1], ARGV[5])
redis.call('SADD', KEYS[2], ARGV[2])

return 1

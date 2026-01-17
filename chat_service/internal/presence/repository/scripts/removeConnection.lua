-- KEYS
-- 1 = connKey
-- 2 = userConnSet

-- ARGV
-- 1 = connId

redis.call('DEL', KEYS[1])
redis.call('SREM', KEYS[2], ARGV[1])

return 1

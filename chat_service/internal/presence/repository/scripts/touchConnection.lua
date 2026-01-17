-- KEYS
-- 1 = connKey

-- ARGV
-- 1 = nowMs
-- 2 = ttlSec

if redis.call('EXISTS', KEYS[1]) == 1 then
    redis.call('HSET', KEYS[1], 'last_activity', ARGV[1])
    redis.call('EXPIRE', KEYS[1], ARGV[2])
    return 1
end

return 0

-- KEYS
-- 1 = userConnSet

local conns = redis.call('SMEMBERS', KEYS[1])
local removed = 0

for _, connId in ipairs(conns) do
    local connKey = 'conn:' .. connId
    if redis.call('EXISTS', connKey) == 0 then
        redis.call('SREM', KEYS[1], connId)
        removed = removed + 1
    end
end

return removed

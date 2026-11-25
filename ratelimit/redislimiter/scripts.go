package redislimiter

// All keys are passed explicitly and share one Redis Cluster hash slot.
// Redis TIME is used so workers do not need synchronized clocks.
const cleanupScript = `
local tm = redis.call('TIME')
local now = tm[1] * 1000 + math.floor(tm[2] / 1000)
local expired = redis.call('ZRANGEBYSCORE', KEYS[5], '-inf', now, 'LIMIT', 0, 100)
for _, entry in ipairs(expired) do
  local kind = string.sub(entry, 1, 2)
  local id = string.sub(entry, 3)
  if kind == 'r|' then redis.call('HDEL', KEYS[2], id)
  elseif kind == 'b|' then redis.call('HDEL', KEYS[3], id .. ':n', id .. ':t')
  elseif kind == 'p|' then redis.call('HDEL', KEYS[4], id .. ':o', id .. ':t') end
  redis.call('ZREM', KEYS[5], entry)
end
`

const acquireScript = cleanupScript + `
local route, owner = ARGV[1], ARGV[2]
local lease, ttl, globalLimit = tonumber(ARGV[3]), tonumber(ARGV[4]), tonumber(ARGV[5])
local configured = redis.call('HGET', KEYS[1], 'limit')
if configured and tonumber(configured) ~= globalLimit then return redis.error_reply('redislimiter global limit mismatch') end
if not configured then redis.call('HSET', KEYS[1], 'limit', globalLimit) end
local reset = tonumber(redis.call('HGET', KEYS[1], 'reset') or '0')
if reset > now then return {reset - now, ''} end
redis.call('ZREMRANGEBYSCORE', KEYS[6], '-inf', now - 1000)
if redis.call('ZCARD', KEYS[6]) >= globalLimit then
  local oldest = redis.call('ZRANGE', KEYS[6], 0, 0, 'WITHSCORES')
  return {math.max(1, tonumber(oldest[2]) + 1000 - now), ''}
end

local id = redis.call('HGET', KEYS[2], route) or ('r:' .. route)
local remaining = redis.call('HGET', KEYS[3], id .. ':n')
local bucketReset = tonumber(redis.call('HGET', KEYS[3], id .. ':t') or '0')
local probe = ''
if not remaining or bucketReset <= now then
  local untilTime = tonumber(redis.call('HGET', KEYS[4], id .. ':t') or '0')
  if untilTime > now then return {untilTime - now, ''} end
  redis.call('HSET', KEYS[4], id .. ':o', owner, id .. ':t', now + lease)
  redis.call('ZADD', KEYS[5], now + lease, 'p|' .. id)
  probe = id
else
  if tonumber(remaining) <= 0 then return {bucketReset - now, ''} end
  redis.call('HINCRBY', KEYS[3], id .. ':n', -1)
end
redis.call('ZADD', KEYS[6], now, owner)
redis.call('PEXPIRE', KEYS[6], 2000)
if redis.call('HEXISTS', KEYS[2], route) == 1 then
  local routeUntil = math.max(now, bucketReset)
  if probe ~= '' then routeUntil = math.max(routeUntil, now + lease) end
  redis.call('ZADD', KEYS[5], routeUntil + ttl, 'r|' .. route)
end
return {0, probe}
`

const updateScript = cleanupScript + `
local route, bucket = ARGV[1], ARGV[2]
local remaining, delay, global, ttl, globalLimit = tonumber(ARGV[3]), tonumber(ARGV[4]), tonumber(ARGV[5]), tonumber(ARGV[6]), tonumber(ARGV[7])
local configured = redis.call('HGET', KEYS[1], 'limit')
if configured and tonumber(configured) ~= globalLimit then return redis.error_reply('redislimiter global limit mismatch') end
if not configured then redis.call('HSET', KEYS[1], 'limit', globalLimit) end
local reset = now + delay
local oldID = redis.call('HGET', KEYS[2], route) or ('r:' .. route)
local id = oldID
if bucket ~= '' then id = 'b:' .. bucket end
local oldRemaining = redis.call('HGET', KEYS[3], id .. ':n')
local oldReset = tonumber(redis.call('HGET', KEYS[3], id .. ':t') or '0')
if not oldRemaining and oldID ~= id then
  oldRemaining = redis.call('HGET', KEYS[3], oldID .. ':n')
  oldReset = tonumber(redis.call('HGET', KEYS[3], oldID .. ':t') or '0')
end
if oldRemaining and oldReset > now then
  remaining = math.min(remaining, tonumber(oldRemaining))
  reset = math.max(reset, oldReset)
end
redis.call('HSET', KEYS[3], id .. ':n', remaining, id .. ':t', reset)
redis.call('ZADD', KEYS[5], math.max(now, reset) + ttl, 'b|' .. id)
if bucket ~= '' then
  redis.call('HSET', KEYS[2], route, id)
  redis.call('ZADD', KEYS[5], math.max(now, reset) + ttl, 'r|' .. route)
end
if global == 1 then
  local oldGlobal = tonumber(redis.call('HGET', KEYS[1], 'reset') or '0')
  local globalReset = math.max(oldGlobal, now + delay)
  redis.call('HSET', KEYS[1], 'reset', globalReset)
end
return 1
`

const releaseScript = `
local id, owner = ARGV[1], ARGV[2]
if redis.call('HGET', KEYS[4], id .. ':o') == owner then
  redis.call('HDEL', KEYS[4], id .. ':o', id .. ':t')
  redis.call('ZREM', KEYS[5], 'p|' .. id)
end
return 1
`

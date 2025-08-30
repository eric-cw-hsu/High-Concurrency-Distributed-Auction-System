-- KEYS[1]: stock key
-- ARGV[1]: quantity to decrease
-- Returns:
-- {status, timestamp}
-- status codes:
-- 0 if successful
-- -1 if stock key does not exist
-- -2 if insufficient stock

local stock = tonumber(redis.call("GET", KEYS[1]) or "-1")
local qty = tonumber(ARGV[1])

if stock == -1 then
  return {-1, 0} -- item not found
end

if stock < qty then
  return {-2, 0} -- out of stock
end

redis.call("SET", KEYS[1], stock - qty)

local redis_time = redis.call("TIME")
local timestamp = redis_time[1] * 1000 + math.floor(redis_time[2] / 1000)

return {0, timestamp} -- success, new stock level, timestamp
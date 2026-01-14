package redis

const (
	// ReserveStockScript is the Lua script for atomic stock reservation
	ReserveStockScript = `
		local current = tonumber(redis.call('GET', KEYS[1]) or '0')
		local quantity = tonumber(ARGV[1])
		
		if current < quantity then
			return {0, current}
		end
		
		local new_stock = redis.call('DECRBY', KEYS[1], quantity)
		redis.call('SETEX', KEYS[2], tonumber(ARGV[3]), ARGV[2])
		
		return {1, new_stock}
	`

	// ReleaseStockScript is the Lua script for releasing reserved stock
	ReleaseStockScript = `
		local reservation_exists = redis.call('EXISTS', KEYS[2])
		if reservation_exists == 0 then
			return {0, "reservation not found"}
		end
		
		redis.call('DEL', KEYS[2])
		local new_stock = redis.call('INCRBY', KEYS[1], tonumber(ARGV[1]))
		
		return {1, new_stock}
	`
)

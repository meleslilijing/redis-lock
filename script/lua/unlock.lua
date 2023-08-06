if redis.call("get", KEYS[1]) == ARGV[1] then 
  return redis.call("del", KEYS[1])
else 
  -- key不存在
  -- 或者key[1]对应的value不等于ARGV[1]，和要解锁的key不一致
  return 0
end
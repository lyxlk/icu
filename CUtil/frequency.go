package CUtil

import (
	"app/icu/RedisKey"
	"app/icu/db"
	"github.com/garyburd/redigo/redis"
)


/**
 * 并发检测 成功设置返回 ok,nil ; 并发返回 "",error
 *
 * param: uint64 UserId
 * param: string UniqueKey
 * param: int64  TimeOut
 * param: string command
 */
func ReqAnti(UserId uint64,UniqueKey string,TimeOut int64,command string) (string, error)  {
	redisKey 	:= RedisKey.ReqAntiConcurrency(UserId,UniqueKey)
	CRedis 		:= db.CRedis.Get()
	defer CRedis.Close()

	var err error
	var ret string

	switch command {

	case "EX"	:
		ret, err = redis.String(CRedis.Do("SET", redisKey, 1, "EX", TimeOut, "NX"))

	case "PX"	:
		ret, err = redis.String(CRedis.Do("SET", redisKey, 1, "PX", TimeOut, "NX"))

	case "DEL"	:
		_, err   = redis.Int64(CRedis.Do("DEL",redisKey)) //主动解锁
		if err != nil {
			ret = ""
		}

	}

	return ret,err
}


/**
 * 记录一段时间内执行次数
 */
func ReqAntiTimes(UserId uint64,UniqueKey string,IncrBy int,TimeOut int64,command string) (int,error)  {
	redisKey 	:= RedisKey.ReqAntiConcurrencyByTime(UserId,UniqueKey)
	CRedis 		:= db.CRedis.Get()
	defer CRedis.Close()

	switch command {
	case "GET" :
		return redis.Int(CRedis.Do("GET",redisKey))
	case "TTL" :
		return redis.Int(CRedis.Do("TTL",redisKey))
	case "SET" :
		ret,err := redis.Int(CRedis.Do("INCRBY",redisKey,IncrBy))
		ttl,_   := redis.Int64(CRedis.Do("TTL",redisKey))
		if ttl < 0 {
			CRedis.Do("EXPIRE",redisKey,TimeOut)
		}

		return ret,err
	}

	return 0,nil
}

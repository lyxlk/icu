/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/27
 * Time: 17:11
 */
package models

import (
	"app/icu/CUtil"
	"app/icu/RedisKey"
	"app/icu/config"
	"app/icu/db"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"strconv"
)

const Expire = 86400 * 7

const LoginTimesByHour = 60 //每小时允许登录次数

func SessionKey(UserId uint64) string {
	rKey := RedisKey.SessionKey(UserId)
	return rKey
}

//验证是否登录
func CheckLogin(UserId uint64,SessId,srcIp string) (bool,int64) {
	rKey    := SessionKey(UserId)

	CRedis  := db.CRedis.Get()
	defer CRedis.Close()

	ttl,err := redis.Int64(CRedis.Do("TTL",rKey))

	if (err != nil) || (ttl <=0) {
		return false,0
	}

	oSessId,err := redis.String(CRedis.Do("HGET",rKey,"sessid"))

	if err != nil {
		return false,0
	}

	if oSessId != SessId {
		return false,0
	}

	//全局保持本次客户端IP
	if srcIp != "" {
		_,err = CRedis.Do("HSET",rKey,"srcIp",srcIp)
		if err != nil {
			logs.Error("session保存用户IP失败",err,UserId)
		}
	}

	//session 续期
	RedisKeyRenewal(rKey,config.SessionExpire,"EXPIRE")

	return true,ttl
}

/**
 * 获取客户端用户IP
 *
 * param: uint64 UserId
 */
func GetUserIP(UserId uint64) (string,error) {
	rKey      := SessionKey(UserId)
	CRedis    := db.CRedis.Get()
	defer CRedis.Close()

	return redis.String(CRedis.Do("HGET",rKey,"srcIp"))

}

/**
 * 初始化session
 *
 * param: uint64            UserId
 * param: map[string]string aMap
 */
func InitUserLoginTarget(UserId uint64,aMap map[string]string) (string,error) {
	rKey      := SessionKey(UserId)
	CRedis    := db.CRedis.Get()
	defer CRedis.Close()

	oId       := CUtil.CreateOrderId(UserId,99)
	sessid    := CUtil.MD5(oId)

	if aMap == nil {
		aMap = make(map[string]string)
	}

	aMap["sessid"] = sessid

	//校验session有效期
	SessionExpire := config.SessionExpire
	if expire, ok := aMap["ExpiresIn"]; ok{
		expiresIn,_ := strconv.Atoi(expire)
		if expiresIn <= SessionExpire {
			SessionExpire = expiresIn
		}
	}

	//golang 三个点省略号的作用总结 https://studygolang.com/articles/27130?fr=sidebar
	//此处标识变长的函数参数
	if _,err := CRedis.Do("HMSET",redis.Args{}.Add(rKey).AddFlat(aMap)...); err != nil {
		logs.Error("Session初始化失败：",err,UserId)
		return "",errors.New("session初始化失败")
	}

	//设置过期时间
	RedisKeyRenewal(rKey,int64(SessionExpire),"EXPIRE")

	return sessid,nil
}

/**
 * session 续期
 */
func RedisKeyRenewal(rKey string, timeOut int64,opt string) (error,int) {
	CRedis    := db.CRedis.Get()
	defer CRedis.Close()

	switch opt {
	case "EXPIRE"	: fallthrough
	case "EXPIREAT"	:
		_,err := CRedis.Do(opt,rKey,timeOut)
		if err != nil {
			logs.Error("RedisKeyRenewal(1):",opt,err)
			return errors.New("设置过期时间失败"),0
		}
	default:
		return errors.New("未知操作"),0
	}

	ttl,err := redis.Int(CRedis.Do("TTL",rKey))

	if err != nil {
		logs.Error("RedisKeyRenewal(2):",err)
		return errors.New("获取过期时间失败"),0
	}

	return nil,ttl
}

/**
 * 退出系统
 *
 * param: uint64 UserId
 * return: error
 */
func LogOut(UserId uint64) error {
	rKey      := SessionKey(UserId)
	CRedis    := db.CRedis.Get()
	defer CRedis.Close()

	_,err 	  := redis.Int(CRedis.Do("DEL",rKey))

	if err != nil {

		if err == redis.ErrNil {
			return nil
		}

		return err
	}

	return nil
}
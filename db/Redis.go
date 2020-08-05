package db

//文档 https://godoc.org/github.com/garyburd/redigo/redis
import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"time"
)

var CRedis *redis.Pool

func InitRedis() {
	redisHost   := beego.AppConfig.String("redisHost")
	redisPort   := beego.AppConfig.String("redisPort")
	redisAuth   := beego.AppConfig.String("redisAuth")
	redisDb,err := beego.AppConfig.Int("redisDb")

	if err != nil {
		logs.Error("redis db err:",err,redisHost,redisPort,redisAuth,redisDb)
		panic("Redis 数据库读取失败")
	}

	redisMaxIdle,err      := beego.AppConfig.Int("redisMaxIdle")
	if err != nil {
		logs.Error("redisMaxIdle err:",err)
		panic("Redis 数据库读取失败")
	}

	redisIdleTimeout,err  := beego.AppConfig.Int("redisIdleTimeout")
	if err != nil {
		logs.Error("redisIdleTimeout err:",err)
		panic("Redis 最大空闲时间设置失败")
	}
	redisMaxActive,err    := beego.AppConfig.Int("redisMaxActive")
	if err != nil {
		logs.Error("redisMaxActive err:",err)
		panic("Redis 最大链接数设置失败")
	}

	CRedis = &redis.Pool{
		MaxIdle:     redisMaxIdle,// 池子里的最大空闲连接
		MaxActive:   redisMaxActive, // 最大链接数
		IdleTimeout: time.Duration(redisIdleTimeout) * time.Second, //空闲时间
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(
				"tcp",
				redisHost + ":" + redisPort,
				redis.DialConnectTimeout(time.Second * 2),
				redis.DialReadTimeout(time.Second * 2),
				redis.DialWriteTimeout(time.Second * 2),
				redis.DialPassword(redisAuth),
				redis.DialDatabase(redisDb),
				)
			if err != nil {
				panic("Redis 初始化失败" + err.Error())
			}

			return conn, nil
		},
	}

	logs.Info("===Redis连接池 初始化完成===")
}
/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/28
 * Time: 18:16
 */
package models

import (
	"app/icu/CUtil"
	"app/icu/RedisKey"
	"app/icu/config"
	"app/icu/db"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"time"
)

type GoldFlowCountStruct struct {
	Count 	int 	`json:"count"    orm:"column(num)"`
}

/**
 * 金币流水日志
 * 由计划任务每日创建表
 *
 * param: string date
 * return: string
 */
func getTbTGoldFlow(date string) string {

	if date == "" {
		date = CUtil.GetTheDate(time.Now().Unix(),"")
	}

	return fmt.Sprintf("`db_fight_log`.`t_flow_gold_%v`",date)
}

/**
 * 每天凌晨零点执行建表sql 提前建好今天和明天数据表
 *
 * return: error
 */
func CreateTbTGoldFlow() error {
	o := orm.NewOrm()
	for index:=0; index<=1 ;index++  {
		date := CUtil.GetTodayBeforeOrAfterDayDateTime(int64(index),0,3)
		tb   := getTbTGoldFlow(date)
		sql  := "CREATE TABLE IF NOT EXISTS %s LIKE `db_fight_log`.`t_flow_gold`"
		sql   = fmt.Sprintf(sql,tb)

		_,err := o.Raw(sql).Exec()

		if err != nil {
			logs.Error("建表失败",err,sql)
		}

		logs.Info("定时建表",sql)
	}

	return  nil
}

/**
 * 每天执行一次 清理一下一年前今天的旧表
 *
 * return: error
 */
func CleanTbTGoldFlow() error {
	location,_     	 := time.LoadLocation(config.TimeZone)
	year, month, day := time.Now().Date()
	date 			 := time.Date(year, month, day, 0, 0, 0, 0, location).AddDate(-1,0,0).Format(config.BaseYmdNoFix)
	o       		 := orm.NewOrm()
	tb   			 := getTbTGoldFlow(date)
	sql  			 := "DROP TABLE IF EXISTS %s"
	sql  	 		  = fmt.Sprintf(sql,tb)

	_,err 			 := o.Raw(sql).Exec()
	if err != nil {
		logs.Error("清理旧表失败",err,sql)
	}

	logs.Info("清理旧表",sql)

	return  nil
}

/**
 * 获取资产流水记录总数key
 *
 * param: uint64 UserId
 * param: int    aType
 * return: string
 */
func getAGoldFlowRecordCountKey(UserId uint64,aType int) string {

	return RedisKey.GetAGoldFlowRecordCount(UserId,aType)
}

/**
 * 删除资产流水记录总数缓存
 *
 * param: uint64 UserId
 * param: int    aType
 * return: bool
 */
func delAGoldFlowRecordCountKeyCache(UserId uint64,aType int) bool {
	rKey   := getAGoldFlowRecordCountKey(UserId,aType)
	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	ret,err := redis.Bool(CRedis.Do("DEL",rKey))

	if err != nil {
		if err == redis.ErrNil {
			return true
		}

		return false
	}

	return ret
}

/**
 * 流水记录
 *
 * param: string    orderId
 * param: uint64    UserId
 * param: int       aType
 * param: int       gold
 * param: string    srcIP
 * param: orm.Ormer oldOrmer 使用调用方创建的orm 便于事物控制
 * return: error
 */
func RecordTbGoldFlow(orderId string,UserId uint64,aType int,gold int,srcIP string,oldOrmer orm.Ormer ) error {

	now     := time.Now().Unix()
	date  	:= CUtil.GetTheDate(now,"-")
	clock  	:= CUtil.GetTheClock(now,":")
	ctime   := fmt.Sprintf("%s %s",date,clock)


	table := getTbTGoldFlow("")

	sql   := "INSERT IGNORE INTO %s (`order_id`,`user_id`,`type`,`gold`,`src_ip`,`ctime`) VALUE (?,?,?,?,?,?)"
	sql    = fmt.Sprintf(sql,table)

	if oldOrmer == nil {
		oldOrmer = orm.NewOrm()
	}
	_,err := oldOrmer.Raw(sql,orderId,UserId,aType,gold,srcIP,ctime).Exec()

	delAGoldFlowRecordCountKeyCache(UserId,aType)

	return err
}

/**
 * 获取资产流水记录总数
 *
 * param: uint64 UserId
 * param: int    aType
 */
func GetAGoldFlowRecordCount(UserId uint64, aType int) (int,error) {

	rKey  := getAGoldFlowRecordCountKey(UserId,aType)

	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	count,err := redis.Int(CRedis.Do("GET",rKey))

	if err != nil && err != redis.ErrNil {

		var aData GoldFlowCountStruct

		table := getTbTGoldFlow("")
		sql   := "SELECT COUNT(*) AS `num` FROM %s WHERE `user_id`=? AND `type`=? LIMIT 1"
		sql    = fmt.Sprintf(sql,table)

		o     := orm.NewOrm()
		err   := o.Raw(sql,UserId,aType).QueryRow(&aData)
		if err != nil && err != orm.ErrNoRows {
			logs.Error("获取资产流水记录失败",err,sql)
			return 0,errors.New("获取资产流水记录失败")
		}

		count = aData.Count

		CRedis.Do("SET",rKey,aData.Count,"EX",600)
	}

	return count,nil
}
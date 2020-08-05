/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/28
 * Time: 18:16
 */
package models

import (
	"app/icu/CUtil"
	"app/icu/config"
	"app/icu/db"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

/**
 * 每月用户押注流水日志
 * 由计划任务按月创建表
 *
 * param: string date
 * return: string
 */
func getTbBetLog(date string) string {

	if date == "" {

		YmdStr := CUtil.GetTheDate(time.Now().Unix(),"-")

		strArrayNew:= strings.Split(YmdStr,"-")

		date = fmt.Sprintf("%s%s",strArrayNew[0],strArrayNew[1])
	}

	return fmt.Sprintf("`db_fight_log`.`t_lhj_bet_log_%v`",date)
}


/**
 * 每天凌晨零点执行建表sql 提前建好本月和下月数据表
 *
 * return: error
 */
func CreateTbBetLog() error {

	o  			   := orm.NewOrm()
	location,_     := time.LoadLocation(config.TimeZone)
	year, month, _ := time.Now().Date()

	//从本月1号开始,否则 AddDate 在日期有31天的月份会出问题
	thisMonth 	   := time.Date(year, month, 1, 0, 0, 0, 0, location)

	for index:=0; index<=1 ;index++  {
		date := thisMonth.AddDate(0,index,0).Format(config.BaseYmNoFix)
		tb   := getTbBetLog(date)

		sql  := "CREATE TABLE IF NOT EXISTS %s LIKE `db_fight_log`.`t_lhj_bet_log`"
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
 * 每天执行一次 清理一下一年前本月的旧表
 *
 * return: error
 */
func CleanTbBetLog() error {
	location,_     	 := time.LoadLocation(config.TimeZone)
	year, month, day := time.Now().Date()
	date 			 := time.Date(year, month, day, 0, 0, 0, 0, location).AddDate(-1,0,0).Format(config.BaseYmNoFix)

	o       		 := orm.NewOrm()
	tb   			 := getTbBetLog(date)
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
 * 每月用户押注流水日志
 *
 * param: uint64 UserId
 * param: int64  roundId
 * param: string betTxt
 * param: uint   add
 * return: error
 */
func RecordTbBetLog(UserId uint64,roundId int64,betTxt string,add uint) error {

	now     := time.Now().Unix()
	table 	:= getTbBetLog("")

	sql   	:= "INSERT INTO %s (`round_id`,`user_id`,`%s`,`bet_time`) VALUE (?,?,?,?) ON DUPLICATE KEY UPDATE `%s`=`%s`+?"
	sql   	 = fmt.Sprintf(sql,table,betTxt,betTxt,betTxt)
	o 		:= orm.NewOrm()

	_,err   := o.Raw(sql,roundId,UserId,add,now,add).Exec()

	return err
}

func GetMyBet(UserId uint64,roundId int64) (myBet map[string]int64) {

	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	myBet    = make(map[string]int64)
	for betTxt  := range config.LhjBetTxt {
		rKey    := GetRedisLhjGameAddBetKey(roundId,betTxt)
		aNum,_  := redis.Int64(CRedis.Do("ZSCORE",rKey,UserId))
		myBet[betTxt] = aNum
	}

	return
}

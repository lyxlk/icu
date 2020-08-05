/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/5/29
 * Time: 17:56
 */
package models

import (
	"app/icu/CUtil"
	"app/icu/RedisKey"
	"app/icu/config"
	"app/icu/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

const LogPages = 20 //最多可以看多少页数据

type ChatLogStruct struct {
	RoomId		string		`json:"name"           orm:"column(room_id)"`
	UserId		uint64		`json:"user_id"        orm:"column(user_id)"`
	Avatar		int			`json:"avatar"         orm:"-"`
	Nick		string		`json:"nick"           orm:"-"`
	Content		string		`json:"content"        orm:"column(content)"`
	Time		int64		`json:"time"           orm:"column(time)"`
}

//数据库记录总数查询结构体
type ChatLogCountStruct struct {
	Count 	uint 	`json:"count"    orm:"column(num)"`
}


func getTbChatBaseTb() string {
	return "db_fight_log.t_chat_content"
}

/**
 * 聊天记录列表
 *
 * param: string roomId
 * return: string
 */
func getChatListKey(roomId config.RoomIdType) string {
	return RedisKey.GetChatListKey(roomId)
}

/**
 * 每日计划任务建表
 */
func CreateChatLogTb() {

	dbTb    := getTbChatBaseTb()

	dbTbArr := strings.Split(dbTb,".")

	o       := orm.NewOrm()

	for i:=0;i<=1 ;i++  {
		date := CUtil.GetTodayBeforeOrAfterMonthDateTime(i,1,0)

		sql   := "CREATE TABLE IF NOT EXISTS `%s`.`%s_%s` LIKE `%s`.`%s`"

		sql    = fmt.Sprintf(sql,dbTbArr[0],dbTbArr[1],date,dbTbArr[0],dbTbArr[1])

		_,err := o.Raw(sql).Exec()

		if err != nil {
			logs.Error("聊天记录表创建失败",sql,err)
		}
	}
}


func getTbChatLog(unixTime int64) (table string) {

	dbTb    := getTbChatBaseTb()

	dateStr := CUtil.GetTheDate(unixTime,"-")

	dateArr := strings.Split(dateStr,"-")

	dbTbArr := strings.Split(dbTb,".")

	table    = fmt.Sprintf("`%s`.`%s_%s%s`",dbTbArr[0],dbTbArr[1],dateArr[0],dateArr[1])

	return table
}


func SaveChatLog (roomId config.RoomIdType,UserId uint64,content string) (err error) {

	nowTime := time.Now().Unix()
	table 	:= getTbChatLog(nowTime)

	o 		:= orm.NewOrm()
	sql 	:= "INSERT INTO %s(`room_id`,`user_id`,`content`,`time`) VALUE( ?, ?, ?, ?)"
	sql 	 = fmt.Sprintf(sql,table)

	_,err    = o.Raw(sql,roomId,UserId,content,nowTime).Exec()

	if err != nil {
		logs.Error("聊天记录失败",err)

		return errors.New("聊天记录失败")
	}

	return nil
}

/**
 * 获取聊天记录列表
 *
 * param: string roomId
 * param: uint   page
 * param: uint   pageSize
 * return: error
 * return: []ChatLogStruct
 */
func GetChatLogList(roomId config.RoomIdType,page uint,pageSize uint) (err error,aData []ChatLogStruct) {

	if page > LogPages {
		return errors.New("默认获取最近200条记录"),aData
	}

	funcName := "GetChatLogList"

	if roomId == "" {
		return nil, aData
	}

	allParams := []interface{}{funcName,page,pageSize}
	subKey    := CUtil.FuncGetArgs(allParams...)

	rKey 	  := getChatListKey(roomId)

	CRedis 	:= db.CRedis.Get()
	defer CRedis.Close()

	aJson,err := redis.String(CRedis.Do("HGET",rKey,subKey))

	if ( err != nil ) || ( aJson == "") {

		logs.Info("==GetChatLogList==",err)

		_,err := CUtil.ReqAnti(0,"GetChatLogList",10,"EX")
		if err != nil {
			return errors.New("网络繁忙，请稍后"),aData
		}

		offset,pageSize := CUtil.Pagination(page,pageSize)

		table := getTbChatLog(time.Now().Unix())

		sql := "SELECT `room_id`,`user_id`,`content`,`time` FROM %s WHERE `room_id` = ? ORDER BY `id` DESC LIMIT ?,?"
		sql  = fmt.Sprintf(sql,table)
		o   := orm.NewOrm()
		num,err := o.Raw(sql,roomId,offset,pageSize).QueryRows(&aData)
		if err != nil {

			if err == orm.ErrNoRows {
				return nil,aData
			}

			logs.Error("获取聊天记录失败",err,sql)
			return errors.New("获取聊天记录失败"),aData
		}


		if num <= 0 {
			return nil,aData
		}

		aJson,err := json.Marshal(aData)
		if err != nil {
			return err,aData
		}

		_,err = redis.Int64(CRedis.Do("HSET",rKey,subKey,string(aJson)))
		if err != nil {
			logs.Error(funcName,err,rKey)
		}

		ttl,_ := redis.Int(CRedis.Do("TTL",rKey))

		if ttl < 0 {
			CRedis.Do("EXPIRE",rKey,300)
		}


		//首次初始化全部记录到缓存
		if page == 1 {
			go func() {
				_,count := GetChatLogListCount(roomId)
				pages 	:= CUtil.GetPaginationPages(uint(count),pageSize)

				if pages > LogPages {
					pages = LogPages
				}

				for index:= page+1; index <= pages; index++  {
					err,_ := GetChatLogList(roomId,index,pageSize)
					if err != nil {
						logs.Info("===========",err,index)
					}
				}
			}()
		}

		//并发查询解锁
		CUtil.ReqAnti(0,"GetChatLogList",0,"DEL")

	} else {
		err := json.Unmarshal([]byte(aJson),&aData)
		if err != nil {
			return errors.New("获取数据失败"),aData
		}
	}

	if len(aData) > 0 {
		for key := range aData {
			aUser := GetOneByUid(aData[key].UserId)
			aData[key].Avatar = aUser.Avatar
			aData[key].Nick = aUser.Nick
		}
	}

	return nil,aData
}


/**
 * 获取聊天总记录数
 *
 * param: string roomId
 * return: error
 * return: int
 */
func GetChatLogListCount(roomId config.RoomIdType) (err error, count int) {

	funcName := "GetChatLogListCount"

	var aData ChatLogCountStruct
	if roomId == "" {
		return nil, 0
	}

	allParams := []interface{}{funcName}
	subKey    := CUtil.FuncGetArgs(allParams...)
	rKey 	  := getChatListKey(roomId)

	CRedis 	:= db.CRedis.Get()
	defer CRedis.Close()

	count,err = redis.Int(CRedis.Do("HGET",rKey,subKey))

	if err != nil {

		_,err := CUtil.ReqAnti(0,"GetChatLogListCount",5,"EX")
		if err != nil {
			return errors.New("网络繁忙，请稍后"),0
		}

		logs.Info("==GetChatLogListCount==",err)

		table := getTbChatLog(time.Now().Unix())

		sql := "SELECT COUNT(*) AS `num` FROM %s WHERE `room_id` = ?"
		sql  = fmt.Sprintf(sql,table)

		o   := orm.NewOrm()

		err = o.Raw(sql,roomId).QueryRow(&aData)
		if err != nil {

			if err == orm.ErrNoRows {
				return nil,0
			}

			logs.Error("获取聊天记录失败总数失败",err)
			return errors.New("获取聊天记录失败总数失败"),0
		}

		count = int(aData.Count)

		_,err = redis.Int64(CRedis.Do("HSET",rKey,subKey,aData.Count))
		if err != nil {
			logs.Error(funcName,err,rKey)
		}

		CRedis.Do("EXPIRE",rKey,300)

		//并发查询解锁
		CUtil.ReqAnti(0,"GetChatLogListCount",0,"DEL")
	}

	if count < 0 {
		return nil,0
	}

	return nil,int(count)
}
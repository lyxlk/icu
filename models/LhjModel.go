/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/6/23
 * Time: 17:13
 */
package models

import (
	"app/icu/CUtil"
	"app/icu/RedisKey"
	"app/icu/config"
	"app/icu/db"
	sql2 "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"time"
)

type LhjLuckLogsStruct struct {
	LuckId    int    `json:"luck_id"`
	LuckName  string `json:"luck_name"`
	LuckTime  int64  `json:"luck_time"`
}


//客户端可展示的数据
type LhjBetInfoShow struct {
	RoundId				int64				`json:"round_id"`
	EndBox				int					`json:"end_box"`
	Status				uint				`json:"status"`
	TopList				uint				`json:"toplist"`
	TotalBet            map[string]int64 	`json:"total_bet"`
	MyBet               map[string]int64 	`json:"my_bet"`
	Gold				uint				`json:"gold"`
	LuckLogs			[]LhjLuckLogsStruct	`json:"luck_logs"`
}

//可查询的数据
type LhjBetInfo struct {
	RoundId				int64	`json:"round_id"                             orm:"column(round_id)"`
	BarTotal			uint	`json:"lhj_bet_txt_bar_total"                orm:"column(lhj_bet_txt_bar_total)"`
	SevenTotal			uint	`json:"lhj_bet_txt_seven_total"              orm:"column(lhj_bet_txt_seven_total)"`
	StartTotal			uint	`json:"lhj_bet_txt_star_total"               orm:"column(lhj_bet_txt_star_total)"`
	WatermelonsTotal	uint	`json:"lhj_bet_txt_watermelons_total"        orm:"column(lhj_bet_txt_watermelons_total)"`
	AlarmTotal			uint	`json:"lhj_bet_txt_alarm_total"              orm:"column(lhj_bet_txt_alarm_total)"`
	CoconutTotal		uint	`json:"lhj_bet_txt_coconut_total"            orm:"column(lhj_bet_txt_coconut_total)"`
	OrangeTotal			uint	`json:"lhj_bet_txt_orange_total"             orm:"column(lhj_bet_txt_orange_total)"`
	AppleTotal			uint	`json:"lhj_bet_txt_apple_total"              orm:"column(lhj_bet_txt_apple_total)"`
	Status				uint	`json:"status"                               orm:"column(status)"`
	Ctime				int64	`json:"ctime"                                orm:"column(ctime)"`
	EndTime				int64	`json:"endtime"                              orm:"column(endtime)"`
	TopList				uint	`json:"toplist"                              orm:"column(toplist)"`
	ResultId			uint	`json:"resultid"                             orm:"column(resultid)"`
	UserCount			uint	`json:"usercount"                            orm:"column(usercount)"`
	Income				uint	`json:"income"                               orm:"column(income)"`
	Cost				uint	`json:"cost"                                 orm:"column(cost)"`
}


//水果机中奖历史数
const LhjLuckIdLogsCount = 24

func getTbLhjBetInfo() string {
	return "`db_fight_log`.`t_lhj_bet_info`"
}


func redisLhjGetLastRoundInfoKey() string {
	return RedisKey.LhjGetLastRoundInfoKey()
}

/**
 * 水果机押注集合
 *
 * param: int64  roundId
 * param: string betTxt
 * return: string
 */
func GetRedisLhjGameAddBetKey(roundId int64,betTxt string) string {
	return RedisKey.LhjGameAddBetKey(roundId,betTxt)
}

/**
 每局游戏参与用户
 */
func GetLhjGameTakePartInKey(roundId int64) string {
	return RedisKey.LhjGameTakePartInKey(roundId)
}

/**
 * 老虎机最近中奖信息
 */
func GetLhjGameLuckIdLogs() string {
	return RedisKey.LhjGameLuckIdLogs()
}


/**
 * 缓存清理
 *
 * param: uint64 UserId
 * return: bool
 */
func delLhjGetLastRoundInfoCache() bool {
	rKey   := redisLhjGetLastRoundInfoKey()
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
 * 获取最后一局信息
 *
 * return: LhjBetInfo
 */
func LhjGetLastRoundInfo() (aData LhjBetInfo) {

	rKey    := redisLhjGetLastRoundInfoKey()
	CRedis  := db.CRedis.Get()
	defer CRedis.Close()

	aJson,err := redis.String(CRedis.Do("GET",rKey))

	if ( err != nil ) || ( aJson == "") {
		o       := orm.NewOrm()
		table   := getTbLhjBetInfo()
		sql     := "SELECT * FROM %s ORDER BY `round_id` DESC LIMIT 1"
		sql      = fmt.Sprintf(sql,table)

		err     := o.Raw(sql).QueryRow(&aData)

		if err != nil {

			return LhjBetInfo{}
		}

		aJson,err := json.Marshal(aData)

		if err != nil {
			return LhjBetInfo{}
		}

		_,err = redis.String(CRedis.Do("SET",rKey,string(aJson),"EX",60))
		if err != nil {
			logs.Error("GetOneIncomeLog",err,rKey)
		}

	} else {
		err := json.Unmarshal([]byte(aJson),&aData)
		if err != nil {
			return LhjBetInfo{}
		}
	}

	return aData
}

/**
 * 创建新一局游戏
 *
 * return: error
 * return: int64
 */
func CreateNewGame() (err error,roundId int64) {
	now     := time.Now().Unix()

	o       := orm.NewOrm()
	table   := getTbLhjBetInfo()
	sql     := "INSERT INTO %s(`ctime`) VALUE(?)"
	sql      = fmt.Sprintf(sql,table)

	var res sql2.Result
	res,err  = o.Raw(sql,now).Exec()

	if err != nil {
		return err,0
	}

	roundId,err = res.LastInsertId()

	if err != nil {
		logs.Error("==CreateNewGame==",err)
		return err,0
	}

	delLhjGetLastRoundInfoCache()

	return
}

/**
 * 更新游戏信息
 */
func UpdateGameInfo(roundId int64 ,fields orm.Params) (bool,error) {

	tb  := getTbLhjBetInfo()

	if fields == nil {
		return false,errors.New("未指定要更新记录")
	}

	sql := " UPDATE %s SET "

	//todo 一定要注意顺序不能乱！！！！！！！！
	var values []interface{}
	for field,val := range fields {
		sql += field + " = ?, "
		values = append(values,val)
	}
	
	//最后一定要是round Id
	values   = append(values,roundId)

	sql		 = strings.Trim(sql,", ") //去掉首尾连接符

	sql      += " WHERE `round_id` = ? LIMIT 1"

	sql      = fmt.Sprintf(sql,tb) //替换库表和指定要更新的用户

	o        := orm.NewOrm()
	_,err    := o.Raw(sql,values...).Exec()

	if err != nil && err != orm.ErrNoRows{
		logs.Error("更新游戏数据失败",err,sql)
		return false, err
	}

	delLhjGetLastRoundInfoCache()

	return true,nil
}



/**
 * 水果机下注
 *
 * param: uint64 UserId
 * param: string betTxt
 * param: uint   add
 * return: error
 */
func LhjGameAddBet(UserId uint64,betTxt string,add uint) error {

	_,ok := config.LhjBetTxt[betTxt]

	if !ok {
		return errors.New("下注对象不存在")
	}

	lastRound := LhjGetLastRoundInfo()

	if lastRound.Status != 0 {
		return errors.New("已结束投注")
	}

	if add != 1 && add != 5 && add != 10 && add != 100{
		return errors.New("无效投注")
	}

	//先扣钱
	aType   := config.GoldTypeLogAddBet
	orderId := CUtil.CreateOrderId(UserId,1)
	_,err 	:= SetUserCoins(UserId,-(int(add)),aType,orderId,"",false)

	if err != nil {
		return err
	}

	//写押注记录
	err = RecordTbBetLog(UserId,lastRound.RoundId,betTxt,add)
	if err != nil {
		logs.Error("===RecordTbBetLog==",err)
		return err
	}

	//加入押注集合
	rKey   := GetRedisLhjGameAddBetKey(lastRound.RoundId,betTxt)
	CRedis := db.CRedis.Get()
	defer CRedis.Close()


	_,err   = redis.Int(CRedis.Do("ZINCRBY",rKey,add,UserId))
	if err != nil {
		logs.Error("加入下注集合失败",rKey,UserId,add,err)
		return  errors.New("下注失败（-100）")
	}

	//过期时间长一些 便于结算
	CRedis.Do("EXPIRE",rKey,600)

	//记录本局所有参与者
	TpKey := GetLhjGameTakePartInKey(lastRound.RoundId)
	CRedis.Do("SADD",TpKey,UserId)
	CRedis.Do("EXPIRE",TpKey,600)

	return nil
}


/**
 * 获取总下注额度
 *
 * param: int64 RoundId
 * return: map[string]int64
 */
func TotalBet(RoundId int64) (aList map[string]int64) {

	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	aList = make(map[string]int64)

	for betTxt := range config.LhjBetTxt {

		aList[betTxt] = 0

		var total int64
		var iter int
		var values []string

		rKey := GetRedisLhjGameAddBetKey(RoundId,betTxt)

		for {

			if arr, err := redis.Values(CRedis.Do("ZSCAN", rKey, iter)); err != nil {
				logs.Error("==TotalBet SSCAN ==",err)
				break
			} else {
				iter, _  = redis.Int(arr[0], nil)
				values,_ = redis.Strings(arr[1],nil)
			}

			//key val 在一维数组里面
			for index := range values {
				if index % 2 == 1 {
					aNum,_ := strconv.ParseInt(values[index],10,64)
					total += aNum
				}
			}

			aList[betTxt] = total

			if iter == 0  {
				break
			}
		}
	}

	return aList
}

/**
 * 保存最近的中奖记录
 *
 * param: int luckId
 * param: int64 luckTime
 */
func AddLhjLuckId(luckId int,luckTime int64) {

	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	rKey    := GetLhjGameLuckIdLogs()
	member  := strconv.Itoa(luckId) + "|" + strconv.FormatInt(luckTime,10)
	_,err   := redis.Int(CRedis.Do("ZADD",rKey,luckTime,member))
	if err != nil {
		logs.Error("AddLhjLuckId",err)

		return
	}

	//最大保存24
	count,_ := redis.Int(CRedis.Do("ZCARD",rKey))
	if count > LhjLuckIdLogsCount {
		_,err = redis.Int(CRedis.Do("ZREMRANGEBYRANK",rKey,0,count - LhjLuckIdLogsCount - 1))
		if err != nil {
			logs.Error("AddLhjLuckId REM",err)
		}
	}

	CRedis.Do("EXPIRE",rKey,86400)
}

/**
 * 获取最近中奖记录
 *
 * return: map[string]string
 */
func GetLhjLuckLogs() ( ret []LhjLuckLogsStruct) {

	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	rKey := GetLhjGameLuckIdLogs()

	aList,_ := redis.Strings(CRedis.Do("ZREVRANGE",rKey,0,LhjLuckIdLogsCount,"WITHSCORES"))

	var aLog LhjLuckLogsStruct
	if len(aList) > 0 {
		for index := range aList {
			if index % 2 == 0 {

				arr := strings.Split(aList[index],"|")

				if len(arr) != 2 {
					continue
				}

				aLog.LuckId,_ = strconv.Atoi(arr[0])

				if betName,ok := config.LhjBetList[aLog.LuckId]; ok {

					aLog.LuckTime,_ = strconv.ParseInt(aList[index+1],10,64)

					aLog.LuckName = betName

					ret = append(ret, aLog)
				}
			}
		}
	}

	return

}
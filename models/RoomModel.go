package models

import (
	"app/icu/CUtil"
	"app/icu/RedisKey"
	"app/icu/config"
	"app/icu/db"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"sort"
	"strconv"
	"time"
)

type TuHaoListStruct struct {
	Gold 	uint
	Nick 	string
	Avatar 	int
	Level   string
}

type TuHaoList []TuHaoListStruct

const (
	BankruptGold  = 200 //每次200破产补助
	BankruptTimes = 5   //每天5次
)

// BubbleStructList 排序规则
func (list TuHaoList) Len() int {
	return len(list)
}

//排序规则
func (list TuHaoList) Less(i, j int) bool {
	return  list[i].Gold >= list[j].Gold
}

func (list TuHaoList) Swap(i, j int) {
	var temp  = list[i]
	list[i]   = list[j]
	list[j]   = temp
}


/**
  土豪榜
 */
func redisTuHaoListKey(date string) string {
	return RedisKey.TuHaoListKey(date)
}

/**
	获取破产补助 返回破产次数
 */
func GetBankruptGold(UserId uint64) (times int,err error) {
	_,err = CUtil.ReqAnti(UserId,"GetBankruptGold",5,"EX")
	if err != nil {
		return 0,errors.New("请勿频繁操作")
	}

	aUser 	:= GetOneByUid(UserId)
	if aUser.Gold > 0 {
		return 0,errors.New("您未破产")
	}

	date 		:= CUtil.GetTheDate(time.Now().Unix(),"")
	UniqueKey 	:= "GetBankruptGold|" + date
	endTime 	:= CUtil.GetTodayEndUnixTime()

	times,err = CUtil.ReqAntiTimes(UserId,UniqueKey,1,endTime,"SET")
	if times > BankruptTimes {
		CUtil.ReqAntiTimes(UserId,UniqueKey,-1,endTime,"SET")
		return 0,errors.New("今日已领完全部补助")
	}

	aType   := config.GoldTypeLogBankrupt

	orderId := CUtil.CreateOrderId(UserId,1)

	_, err = SetUserCoins(UserId,BankruptGold,aType,orderId,"",false)
	if err != nil {
		return 0,errors.New("领取破产补助失败")
	}

	return times,nil
}

/**
 加入土豪榜并返回名次
 */
func AddTuHaoList(UserId uint64) (rank int) {
	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	date  	:= CUtil.GetTheDate(time.Now().Unix(),"")
	rKey  	:= redisTuHaoListKey(date)
	EndTime := CUtil.GetTodayEndUnixTime()

	aUser := GetOneByUid(UserId)

	//加入土豪榜
	CRedis.Do("ZADD",rKey,aUser.Gold,UserId)

	//设置过期时间
	CRedis.Do("EXPIREAT",rKey,EndTime)

	rank,_ = redis.Int(CRedis.Do("ZREVRANK",rKey,UserId))

	return rank
}

/**
 获取每日土豪行100名
 */
func GetTuHaoList() (aList TuHaoList,err error) {
	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	date  := CUtil.GetTheDate(time.Now().Unix(),"")
	rKey  := redisTuHaoListKey(date)

	var aMap map[string]string
	aMap,err = redis.StringMap(CRedis.Do("ZREVRANGE",rKey,0,100,"WITHSCORES"))
	if err != nil {
		logs.Error("获取土豪榜失败",err)
		return aList, errors.New("网络繁忙")
	}

	if len(aMap) == 0 {
		return aList,nil
	}

	var aRet TuHaoListStruct

	for aUid,aGold := range aMap {

		UserId,err := strconv.ParseUint(aUid,10,64)
		if err != nil {
			logs.Error("土豪榜用户ID值类型转换失败",aUid,err)
			continue
		}

		aUser 		:= GetOneByUid(UserId)
		Gold,_ 		:= strconv.Atoi(aGold)
		aRet.Gold 	 = uint(Gold)
		aRet.Avatar  = aUser.Avatar
		aRet.Nick    = aUser.Nick
		aRet.Level   = GetLevel(aRet.Gold)

		aList = append(aList,aRet)
	}

	sort.Sort(aList)

	return
}
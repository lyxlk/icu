/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/27
 * Time: 18:19
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
	"math"
	"strconv"
	"strings"
	"time"
)

//json标签意义是定义此结构体解析为json或序列化输出json时value字段对应的key值,如不想此字段被解析可将标签设为`json:"-"`
//orm的column标签意义是将orm查询结果解析到此结构体时每个结构体字段对应的数据表字段名
type UserInfo struct {
	UserId		uint64		`json:"user_id"        orm:"column(user_id)"`
	OpenUdId	string		`json:"openudid"       orm:"column(openudid)"`
	From		uint8		`json:"from"           orm:"column(from)"`
	Nonce		string		`json:"nonce"          orm:"column(nonce)"`
	Email		string		`json:"email"          orm:"column(email)"`
	Avatar		int			`json:"avatar"         orm:"column(avatar)"`
	Nick		string		`json:"nick"           orm:"column(nick)"`
	Age			uint8		`json:"age"            orm:"column(age)"`
	Sex			uint8		`json:"sex"            orm:"column(sex)"`
	Fail		int			`json:"fail"           orm:"column(fail)"`
	Win			int			`json:"win"            orm:"column(win)"`
	Draw		int			`json:"draw"           orm:"column(draw)"`
	Gold		uint		`json:"gold"           orm:"column(gold)"`
	IsRegister	uint8		`json:"isRegister"`
	RegTime		int64		`json:"reg_time"       orm:"column(reg_time)"`
}

const RegTimeByIp = 10 //每个IP一天限制10个账号

/**
 * 数据表
 *
 * return: string
 */
func getTbUser() string {
	return "`db_fight`.`t_user`"
}

/**
 * Redis Key
 *
 * param: uint64 UserId
 * return: string
 */
func getUserInfoKey(UserId uint64) string {
	return RedisKey.GetUserInfoKey(UserId)
}

func getUserInfoByOpenUdIdKey(openudid string,from uint8) string {
	return RedisKey.GetUserInfoByOpenUdIdKey(openudid,from)
}

/**
 * 清理用户信息缓存
 *
 * param: uint64 UserId
 * return: bool
 */
func delTbUserCache(UserId uint64) bool {
	rKey   := getUserInfoKey(UserId)
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
 * 获取基础用户信息
 *
 * param: uint64 UserId
 * return: UserInfo
 */
func getBaseByUid(UserId uint64) UserInfo {

	if UserId <= 0 {
		return UserInfo{}
	}

	var aData UserInfo

	rKey    := getUserInfoKey(UserId)
	CRedis  := db.CRedis.Get()
	defer CRedis.Close()

	aJson,err := redis.String(CRedis.Do("GET",rKey))

	if ( err != nil ) || ( aJson == "") {
		o       := orm.NewOrm()
		table   := getTbUser()

		sql     := "SELECT * FROM " + table + " WHERE `user_id`=? LIMIT 1"
		err     := o.Raw(sql,UserId).QueryRow(&aData)

		if err != nil {
			return  UserInfo{}
		}

		aJson,err := json.Marshal(aData)

		if err != nil {
			return UserInfo{}
		}

		_,err = redis.String(CRedis.Do("SET",rKey,string(aJson),"EX",600))
		if err != nil {
			logs.Error("getBaseByUid",err,rKey)
		}

	} else {
		err := json.Unmarshal([]byte(aJson),&aData)
		if err != nil {
			return UserInfo{}
		}
	}

	return aData
}

/**
 * 获取用户信息
 *
 * param: uint64 UserId
 * return: UserInfo
 */
func GetOneByUid(UserId uint64) UserInfo {
	baseInfo := getBaseByUid(UserId)

	if baseInfo.OpenUdId == "" {
		return UserInfo{}
	}

	return baseInfo
}

/**
 * UserId|Nonce|orderId 组装二维码信息
 *
 * param: uint64 UserId
 * return: string
 */
func EncodeUserQrCode(UserId uint64) string {
	aUser   := GetOneByUid(UserId)
	aTime   := time.Now().Nanosecond()
	aString := fmt.Sprintf("%d|%s|%d",UserId,aUser.Nonce,aTime)
	code    := CUtil.AesEncrypt(aString,config.CommSalt)
	return code
}

/**
 * 解析二维码数组
 *
 * param: string aString
 * return: uint64
 * return: string
 * return: int64
 */
func DecodeUserQrCode(aString string) (userId uint64,nonce string,aTime int64) {

	var err error

	origin := CUtil.AesDecrypt(aString,config.CommSalt)
	arr    := strings.Split(origin,"|")

	if len(arr) != 3 {
		return 0,"",0
	}

	userId,err = strconv.ParseUint(arr[0],10,64)
	if err != nil {
		return 0,"",0
	}

	nonce      = arr[1]
	aTime,_    = strconv.ParseInt(arr[2],10,64)

	return
}

/**
 * 根据openUdId获取用户信息
 *
 * param: string openUdid
 * param: uint8  from
 * return: UserInfo
 */
func GetOneByOpenUdId(openUdid string,from uint8) UserInfo {

	if openUdid == "" || from == 0 {
		return UserInfo{}
	}

	var aData UserInfo

	rKey := getUserInfoByOpenUdIdKey(openUdid,from)
	CRedis  := db.CRedis.Get()
	defer CRedis.Close()

	UserId,err := redis.Uint64(CRedis.Do("GET",rKey))

	if ( err != nil ) || ( UserId == 0) {
		o       := orm.NewOrm()
		table   := getTbUser()
		sql     := "SELECT `user_id` FROM %s WHERE `openudid` = ? AND `from`= ? LIMIT 1"
		sql      = fmt.Sprintf(sql,table)

		err     := o.Raw(sql,openUdid,from).QueryRow(&aData)

		if err != nil || aData.UserId == 0 {
			return UserInfo{}
		}

		UserId   = aData.UserId

		_,err = redis.String(CRedis.Do("SET",rKey, UserId ,"EX",600))
		if err != nil {
			logs.Error("GetOneByOpenUdId",err,rKey)
		}
	}

	aData = GetOneByUid(UserId)

	return aData
}


/**
 * 更新扩展表指定数据
 *
 * param: uint64     UserId
 * param: orm.Params Params
 */
func UpdateUserInfo(UserId uint64,fields orm.Params) (bool,error) {

	if fields == nil {
		return false,errors.New("未指定要更新记录")
	}

	table   := getTbUser()

	sql     := " UPDATE %s SET "

	//todo 一定要注意顺序不能乱！！！！！！！！
	var values []interface{}
	for field,val := range fields {

		//不能改资金
		if field == "gold" {
			continue
		}

		sql += field + " = ?, "
		values = append(values,val)
	}

	sql		 = strings.Trim(sql,", ") //去掉首尾连接符

	sql      += " WHERE `user_id` = ? LIMIT 1"
	values   = append(values,UserId)

	sql      = fmt.Sprintf(sql,table) //替换库表和指定要更新的用户

	o        := orm.NewOrm()
	_,err    := o.Raw(sql,values...).Exec()

	if err != nil && err != orm.ErrNoRows{
		logs.Error("更新扩展信息失败",err,sql)
		return false, err
	}

	delTbUserCache(UserId)

	return true,nil
}

/**
 * 同一IP每天注册次数
 */
func OptIpRegTimes(ip,opt string,num int) (int,error)  {

	rKey 	:= RedisKey.IpRegKey(ip)

	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	switch opt {
	case "GET" :
		return redis.Int(CRedis.Do("GET",rKey))
	case "SET" :
		ret,err := redis.Int(CRedis.Do("INCRBY",rKey,num))
		endTime := CUtil.GetTodayEndUnixTime()
		CRedis.Do("EXPIREAT",rKey,endTime)

		return ret,err
	}

	return 0,nil
}

/**
 * 系统注册
 *
 * param: string openudid
 * param: uint8  from
 * param: string nick
 */
func RegisterAUser(Ip string,openudid string,nonce string,from uint8,nick string,age uint8)  (aUser UserInfo,err error) {

	gold := 200

	if openudid == "" || from == 0 || Ip == ""{
		return aUser,errors.New("参数丢失、注册失败")
	}

	_, err = CUtil.ReqAnti(0,openudid,10,"EX")
	if err != nil {
		return UserInfo{},errors.New("请勿频繁重复操作")
	}

	var num int
	num,err = OptIpRegTimes(Ip,"SET",1)
	if err != nil {
		logs.Error("RegisterAUser err",err)
		return UserInfo{},errors.New("注册失败")
	}

	if num > RegTimeByIp {
		OptIpRegTimes(Ip,"SET",-1)

		return UserInfo{},errors.New("少年、您已注册较多账号，要节制啊~")
	}

	regTime := time.Now().Unix()
	table   := getTbUser()

	o       := orm.NewOrm()
	err     = o.Begin()
	if err != nil {
		logs.Error("注册事物开启失败",err)
		return UserInfo{},errors.New("注册事物开启失败")
	}

	sql     := "INSERT IGNORE INTO %s (`openudid`,`nonce`,`nick`,`age`,`from`,`gold`,`reg_time`) VALUE (?, ?, ?, ?, ?, ?, ?)"
	sql      = fmt.Sprintf(sql,table)

	ret,err2 := o.Raw(sql,openudid,nonce,nick,age,from,gold,regTime).Exec()
	if err2 != nil {
		o.Rollback()
		logs.Error("数据库异常注册失败",err2)
		return UserInfo{},errors.New("数据库异常,注册失败")
	}

	var lastId int64

	lastId,err = ret.LastInsertId()
	if err != nil {
		logs.Error("获取注册生成UserId失败",err)
		return UserInfo{},errors.New("获取注册生成UserId失败")
	}

	UserId := uint64(lastId)

	err = o.Commit()

	if err != nil {
		logs.Error("提交注册事物失败",err)
		return UserInfo{},errors.New("提交注册事物失败")
	}

	delTbUserCache(UserId)

	aUser = GetOneByUid(UserId)

	aUser.IsRegister = 1

	return aUser,nil
}

/**
 * 操作用户资产
 *
 * param: uint64 UserId
 * param: int    nums
 * param: int    aType
 * param: string orderId
 * param: string srcIp
 * param: bool   reset
 */
func SetUserCoins(UserId uint64,nums int,aType int,orderId string,srcIp string,reset bool) (int, error) {

	if nums >= 1000000 {
		return 99, errors.New("limit error")
	}

	aUser := GetOneByUid(UserId)
	if aUser.UserId == 0 {
		return 99, errors.New("用户不能存在")
	}

	o    := orm.NewOrm()

	err  := o.Begin()
	if err != nil {
		logs.Error("开启资产操作事物失败:",err)
		return 99 ,errors.New("开启资产操作事物失败")
	}

	tb   := getTbUser()

	if reset == true {
		sql   := fmt.Sprintf("UPDATE %s SET `gold`= ? WHERE `user_id`=? LIMIT 1",tb)
		_,err := o.Raw(sql,nums,UserId).Exec()

		if err != nil {
			o.Rollback()

			logs.Error("重置资产失败：",err.Error(),UserId)
			return 99,errors.New("重置资产失败")
		}


	} else {
		var  aParamsArr []interface{}

		aParamsArr = append(aParamsArr,nums)

		where := " 1=1 "

		if nums < 0 {
			absNum := uint(math.Abs(float64(nums)))

			if absNum > aUser.Gold {
				o.Rollback()
				return 99,errors.New("金币值不足")
			}

			where += " AND `gold` >= ? "
			aParamsArr = append(aParamsArr,absNum)
		}


		sql   := fmt.Sprintf("UPDATE %s SET `gold`= `gold` + ? WHERE %s AND `user_id`=? LIMIT 1",tb, where)
		aParamsArr = append(aParamsArr,UserId)

		_,err := o.Raw(sql,aParamsArr...).Exec()
		if err != nil {
			o.Rollback()

			logs.Error("金币值更新失败：",err.Error(),UserId)
			return 99, errors.New("金币值更新失败")
		}
	}

	//写总记录流水
	err   = RecordTbGoldFlow(orderId,UserId,aType,nums,srcIp,o)
	if err != nil {
		o.Rollback()
		logs.Error("云钻流水记录失败：",err.Error(),UserId)
		return 99 , errors.New("云钻流水记录失败(1)")
	}

	err = o.Commit()
	if err!=nil {
		logs.Error("金币事物提交失败：",err.Error(),UserId)
		return 99 , errors.New("金币事物提交失败")
	}

	//立即清理缓存
	delTbUserCache(UserId)

	return  0, nil
}


/**
 * 按金币获取等级
 *
 * param: int Gold
 * return: string
 */
func GetLevel(Gold uint) string {

	if Gold >=0 && Gold < 10000 {
		return "包身工"
	}

	if Gold >= 10000 && Gold < 25000 {
		return "短工"
	}

	if Gold >= 25000 && Gold < 40000 {
		return "长工"
	}

	if Gold >= 40000 && Gold < 80000 {
		return "佃户"
	}

	if Gold >= 80000 && Gold < 140000 {
		return "贫农"
	}

	if Gold >= 140000 && Gold < 250000 {
		return "渔夫"
	}

	if Gold >= 250000 && Gold < 365000 {
		return "猎人"
	}

	if Gold >= 365000 && Gold < 500000 {
		return "中农"
	}

	if Gold >= 500000 && Gold < 700000 {
		return "富农"
	}

	if Gold >= 700000 && Gold < 1000000 {
		return "掌柜"
	}

	if Gold >= 1000000 && Gold < 1500000 {
		return "商人"
	}

	if Gold >= 1500000 && Gold < 2200000 {
		return "衙役"
	}

	if Gold >= 2200000 && Gold < 3000000 {
		return "小财主"
	}

	if Gold >= 3000000 && Gold < 4000000 {
		return "大财主"
	}

	if Gold >= 4000000 && Gold < 5500000 {
		return "小地主"
	}

	if Gold >= 5500000 && Gold < 7700000 {
		return "大地主"
	}

	if Gold >= 7700000 && Gold < 10000000 {
		return "知县"
	}

	if Gold >= 10000000 && Gold < 14000000 {
		return "通判"
	}

	if Gold >= 14000000 && Gold < 20000000 {
		return "知府"
	}

	if Gold >= 20000000 && Gold < 30000000 {
		return "总督"
	}

	if Gold >= 30000000 && Gold < 45000000 {
		return "巡抚"
	}

	if Gold >= 45000000 && Gold < 70000000 {
		return "丞相"
	}

	if Gold >= 70000000 {
		return "帝王"
	}

	return "屌丝"
}
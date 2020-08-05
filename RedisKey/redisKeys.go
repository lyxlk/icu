package RedisKey

import (
	"app/icu/config"
	"fmt"
)

//需要遵守命名规则 【前缀】 + 【方法名】 + 【参数】
//Key前缀、防项目冲突
const PreFix string = "icu"

//并发频率限制
func ReqAntiConcurrency(UserId uint64,key string) string {
	return fmt.Sprintf("%s|ReqAntiConcurrency|%d|%s",PreFix,UserId,key)
}

//N秒内并发频率限制
func ReqAntiConcurrencyByTime(UserId uint64,key string) string {
	return fmt.Sprintf("%s|ReqAntiConcurrencyByTime|%d|%s",PreFix,UserId,key)
}


//session
func SessionKey(UserId uint64) string {
	return fmt.Sprintf("%s|SessionKey|%d",PreFix,UserId)
}

//用户信息
func GetUserInfoKey(UserId uint64) string {
	return fmt.Sprintf("%s|GetUserInfoKey|%d",PreFix,UserId)
}

//根据openudid获取用户信息
func GetUserInfoByOpenUdIdKey(openUdId string,from uint8) string {
	return fmt.Sprintf("%s|GetUserInfoByOpenUdIdKey|%s|%d",PreFix,openUdId,from)
}

//同一IP 注册次数
func IpRegKey(ip string) string  {
	return fmt.Sprintf("%s|IpRegKey|%s",PreFix,ip)
}

//聊天记录列表
func GetChatListKey(roomId config.RoomIdType) string  {
	return fmt.Sprintf("%s|GetChatListKey|%s",PreFix,roomId)
}

//老虎机最后一局信息
func LhjGetLastRoundInfoKey() string {
	return fmt.Sprintf("%s|LhjGetLastRoundInfoKey",PreFix)
}

//每一局所有下注集合
func LhjGameAddBetKey(roundId int64,betTxt string) string {
	return fmt.Sprintf("%s|LhjGameAddBetKey|%d|%s",PreFix,roundId,betTxt)
}

//获取某类资产流水总记录数
func GetAGoldFlowRecordCount(UserId uint64,aType int) string {
	return fmt.Sprintf("%s|GetAGoldFlowRecordCount|%d|%d",PreFix,UserId,aType)
}

//每局游戏参与用户
func LhjGameTakePartInKey(roundId int64) string {
	return fmt.Sprintf("%s|LhjGameTakePartInKey|%d",PreFix,roundId)
}

//每日土豪榜
func TuHaoListKey(date string) string {
	return fmt.Sprintf("%s|TuHaoListKey|%s",PreFix,date)
}

//老虎机最近中奖信息
func LhjGameLuckIdLogs() string {
	return fmt.Sprintf("%s|LhjGameLuckIdLogs",PreFix)
}
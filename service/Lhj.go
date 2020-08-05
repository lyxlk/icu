/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/6/23
 * Time: 17:13
 */
package service

import (
	"app/icu/CUtil"
	"app/icu/config"
	"app/icu/db"
	"app/icu/models"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"strconv"
	"time"
)

var EndBox = 1

func init() {
	go launchLhjGame()
}

func launchLhjGame() {

	//循环开启游戏
	for {

		//每局结果 重新播种 保证随机数不重复
		rand.Seed(time.Now().UnixNano())

		lastRound := models.LhjGetLastRoundInfo()

		if lastRound.RoundId == 0 || lastRound.Status == 2{

			logs.Info("====创建新一轮游戏....")

			models.CreateNewGame()

			lastRound = models.LhjGetLastRoundInfo()

			if lastRound.RoundId == 0 || lastRound.Status != 0 {
				logs.Error("====开启新一局游戏失败",lastRound)
				return
			}

			logs.Info("====新一轮游戏开始....",lastRound.RoundId)

			//状态通知
			go SendRoundStartToClient(lastRound.RoundId)
		}

		//投注结束时间
		betEndTime := lastRound.Ctime + int64(config.LhjGameConfig.AddTime)

		sleepTime  := betEndTime - time.Now().Unix()

		if sleepTime > 0 {
			logs.Info("====休眠等待押注结束...",lastRound.RoundId,sleepTime)

			time.Sleep(time.Second * time.Duration(sleepTime))
		}

		//结束 押注时间
		models.UpdateGameInfo(lastRound.RoundId,orm.Params{"status":1})

		logs.Info("====押注结束,开始结算",lastRound.RoundId)

		//开始结算
		betTxt,luckId := LhjCreateGameLuck()

		//todo .... 异步给客户端推送中奖效果
		go SendRoundRetToClient(lastRound.RoundId,betTxt,luckId)

		//休眠等待结算
		countEndTime  := lastRound.Ctime + int64(config.LhjGameConfig.AddTime + config.LhjGameConfig.CountTime)

		sleepTime = countEndTime - time.Now().Unix()
		if sleepTime > 0 {
			time.Sleep(time.Second * time.Duration(sleepTime))
		}

		//结算
		go LhjSendGameAward(lastRound.RoundId,betTxt,luckId)

		//休眠等待下一局开始
		WaitEndTime  := lastRound.Ctime + int64(config.LhjGameConfig.AddTime + config.LhjGameConfig.CountTime + config.LhjGameConfig.WaitTime)

		sleepTime = WaitEndTime - time.Now().Unix()
		if sleepTime > 0 {
			time.Sleep(time.Second * time.Duration(sleepTime))
		}
	}

}

/**
 新一轮开始
 */
func SendRoundStartToClient(roundId int64) {
	CRedis := db.CRedis.Get()
	defer CRedis.Close()


	status,aEvent := models.NewEvent(0,models.EventGameBetNewRound,"新一轮开始~",map[string]string{
		"roundId": strconv.FormatInt(roundId,10),
		"status" : "0",
	})

	if !status {
		return
	}

	//所有参与者
	for _,aRoom := range RoomMap {
		if aRoom == nil {
			continue
		}

		aRoom.Range(func(key, val interface{}) bool {

			AClient := val.(*AClient)

			if AClient != nil {
				AClient.WsSendMsg(1,"OK",aEvent)
			}

			return true
		})
	}
}

/**
 抽奖
 */
func LhjCreateGameLuck() (string,int) {

	betTxt := "CHA"

	err,luckBet := CUtil.GetLuckyKey(config.LhjBetTxtRate)
	if err != nil {
		logs.Error("抽奖失败",err)
		return betTxt,6
	}

	if rate,ok := config.RandIndexRate[luckBet];ok {
		err,luckId := CUtil.GetLuckyKey(rate)
		if err != nil {
			logs.Error("抽奖失败2",err)
			return betTxt,6
		}


		for key,val := range config.LhjBetTxt {
			if luckBet == val {
				betTxt = key
				break
			}
		}

		return betTxt,luckId

	} else {

		return betTxt,6
	}
}

/**
 * 发放奖励并通知
 *
 * param: int64  roundId
 * param: string betTxt
 */
func LhjSendGameAward(roundId int64,LuckBet string,LuckId int) {

	lastRound := models.LhjGetLastRoundInfo()

	if lastRound.RoundId != roundId {
		return
	}

	if lastRound.Status != 1 {
		return
	}

	rKey   := models.GetRedisLhjGameAddBetKey(roundId,LuckBet)
	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	//todo ------------------------ 更新数据库信息------------------------------
	endStatus 	:= 2
	endTime 	:= time.Now().Unix()
	params 		:= orm.Params{}
	params["status"] 	= endStatus
	params["resultid"] 	= LuckId
	params["endtime"] 	= endTime

	aList := models.TotalBet(roundId)

	for betText,val := range aList {
		params[betText + "_total"] = val
	}

	_,err := models.UpdateGameInfo(roundId,params)

	if err == nil {
		models.AddLhjLuckId(LuckId,endTime)
	}


	//todo -----------现象推送

	CurrentInfo := map[string]interface{}{
		"roundId"	: strconv.FormatInt(roundId,10),
		"LuckBet"	: LuckBet,
		"LuckId" 	: LuckId,
		"Status" 	: endStatus,
		"Msg"    	: "很遗憾本次没中奖，请再接再厉~",
		"Award"  	: "0",
		"LuckLogs"  : models.GetLhjLuckLogs(),
	}

	//结算
	var eventType models.EventType

	if bei,ok := config.LhjPrizeBei[LuckId];ok {

		//所有参与者 都没中奖
		for _,aRoom := range RoomMap {
			if aRoom == nil {
				continue
			}

			aRoom.Range(func(key, val interface{}) bool {

				AClient := val.(*AClient)

				if AClient != nil {

					if bei == 0 {

						eventType = models.EventGameBetNoWin

						CurrentInfo["Msg"] = "很遗憾本次没中奖，请再接再厉~"

					} else {
						//部分中奖
						betNum,_ := redis.Int(CRedis.Do("ZSCORE",rKey,AClient.UserId))

						if betNum <= 0 {

							eventType = models.EventGameBetNoWin
							CurrentInfo["Msg"] = "很遗憾本次没中奖，请再接再厉~"

						} else {

							eventType = models.EventGameBetWin

							award := betNum * bei

							CurrentInfo["Msg"]   = fmt.Sprintf("恭喜你本局获得[%d金币]奖励",award)
							CurrentInfo["Award"] = strconv.Itoa(award)

							//返奖
							aType   := config.GoldTypeLogAddBetWin

							orderId := CUtil.CreateOrderId(AClient.UserId,1)

							go models.SetUserCoins(AClient.UserId,award,aType,orderId,"",false)
						}
					}

					status,aEvent := models.NewEvent(0,eventType,"公布开奖结果中",CurrentInfo)
					if status {
						AClient.WsSendMsg(1,"OK",aEvent)
					}
				}

				return true
			})
		}
	}

}

/**
 周知本次抽奖结果
 */
func SendRoundRetToClient(roundId int64,LuckBet string,LuckId int) {
	CRedis := db.CRedis.Get()
	defer CRedis.Close()

	aExt := map[string]string{
		"roundId": strconv.FormatInt(roundId,10),
		"LuckBet": LuckBet,
		"LuckId" : strconv.Itoa(LuckId),
		"EndBox" : strconv.Itoa(EndBox),
		"status" : "1",
		"gold"   : "0",
	}

	//所有参与者
	for _,aRoom := range RoomMap {
		if aRoom == nil {
			continue
		}

		aRoom.Range(func(key, val interface{}) bool {

			AClient := val.(*AClient)

			if AClient != nil {

				aUser 	:= models.GetOneByUid(AClient.UserId)

				aExt["gold"] = strconv.FormatUint(uint64(aUser.Gold),10)

				status,aEvent := models.NewEvent(0,models.EventGameBetLuck,"开奖啦~",aExt)

				if status {

					AClient.WsSendMsg(1,"OK",aEvent)
				}
			}

			return true
		})
	}

	EndBox = LuckId

}

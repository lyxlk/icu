/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/5/26
 * Time: 18:42
 */
package v1

import (
	"app/icu/CUtil"
	"app/icu/config"
	"app/icu/models"
	"app/icu/service"
	"fmt"
	"sort"
)

type RoomController struct {
	BaseController
}

type GetUsersController struct {
	UserId		uint64		`json:"user_id"`
	Avatar		int			`json:"avatar"`
	Nick		string		`json:"nick"`
}

/**
 * todo 拉取房间列表 ....支持多房间群聊
 */
func (this *RoomController) GetList() {
	aList := config.RoomList.Data

	var kArr []string

	for key := range aList {
		kArr = append(kArr,string(key))
	}

	sort.Strings(kArr)

	var ret [] config.RoomInfoStruct

	for _,aKey := range kArr{

		_key := config.RoomIdType(aKey)

		if aList[_key].Type != 1 {
			continue
		}

		ret = append(ret,aList[_key])
	}

	this.JsonResponse(1,"OK",ret)
}

/**
 * 获取在房间里的所有部分活跃用户
 */
func (this *RoomController) GetUsers() {
	hasJoin,OldClient := service.HasJoinRoom(this.GetUserId(true))
	if !hasJoin || OldClient == nil {
		this.JsonResponse(1,"未进入任何聊天室",nil)
	}

	roomId := OldClient.RoomId

	var ret []GetUsersController

	if sMap,ok := service.RoomMap[roomId]; ok {
		sMap.Range(func(key, value interface{}) bool {

			client := value.(*service.AClient)
			aUser  := models.GetOneByUid(client.UserId)


			ret    = append(ret,GetUsersController{
				UserId	:	aUser.UserId,
				Nick	:	aUser.Nick,
				Avatar	:	aUser.Avatar,
			})

			if len(ret) > 30 {
				return false
			}

			return true
		})
	}

	this.JsonResponse(1,"OK",ret)

}

/**
 * 聊天历史记录
 */
func (this *RoomController) ChatLogs() {
	hasJoin,OldClient := service.HasJoinRoom(this.GetUserId(true))
	if !hasJoin  || OldClient == nil{
		this.JsonResponse(1,"未进入任何聊天室",nil)
	}

	page,err    := this.GetInt("page",1)
	if err != nil || page < 1{
		this.JsonResponse(-1,"参数错误",nil)
	}

	var pageSize uint = 10
	roomId := OldClient.RoomId

	//todo 先查询总记录
	err,count := models.GetChatLogListCount(roomId)
	if err != nil{
		this.JsonResponse(-1,err.Error(),nil)
	}
	pages := CUtil.GetPaginationPages(uint(count),pageSize)


	err , aData := models.GetChatLogList(roomId,uint(page),pageSize)
	if err != nil{
		this.JsonResponse(-1,err.Error(),nil)
	}

	this.JsonResponse(1,"OK",map[string]interface{}{
		"list"  : aData,
		"pages" : pages,
	})
}

/**
 * 获取当前一局信息
 */
func (this *RoomController) BetInfo() {

	var ret models.LhjBetInfoShow

	UserId      := this.GetUserId(true)
	aUser       := models.GetOneByUid(UserId)
	info 		:= models.LhjGetLastRoundInfo()
	ret.RoundId  = info.RoundId
	ret.Status   = info.Status
	ret.TopList  = info.TopList
	ret.Gold     = aUser.Gold
	ret.TotalBet = models.TotalBet(ret.RoundId)
	ret.MyBet    = models.GetMyBet(UserId,ret.RoundId)
	ret.EndBox   = service.EndBox

	ret.LuckLogs = models.GetLhjLuckLogs()

	this.JsonResponse(1,"OK",ret)
}

/**
  破产补助
 */
func (this *RoomController) Bankrupt() {
	UserId := this.GetUserId(true)
	times,err := models.GetBankruptGold(UserId)
	if err != nil{
		this.JsonResponse(-1,err.Error(),nil)
	}

	aUser := models.GetOneByUid(UserId)
	msg   := fmt.Sprintf("恭喜获得今日第%d次破产补助：%d 金币",times,models.BankruptGold)

	this.JsonResponse(1,msg,map[string]uint{
		"gold" : aUser.Gold,
	})
}

/**
 土豪榜
 */
func (this *RoomController) TuHaoList() {
	aList,err := models.GetTuHaoList()
	if err != nil{
		this.JsonResponse(-1,err.Error(),nil)
	}

	rank := models.AddTuHaoList(this.GetUserId(true))

	this.JsonResponse(1,"OK",map[string]interface{}{
		"list" : aList,
		"rank" : rank,
	})
}
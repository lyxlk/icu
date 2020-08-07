/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/3/31
 * Time: 19:21
 */
package service

import (
	"app/icu/CUtil"
	"app/icu/models"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
)

func (client *AClient) Request(message []byte) {

	var aJson ClientMsg

	err := json.Unmarshal(message,&aJson)
	if err != nil {
		logs.Info("message Unmarshal err, user_id[%d] err:%v | stringINfo: %s", client.UserId, err,message)
		return
	}

	//过滤html标签 空消息或命令字错了 不转发消息
	content := CUtil.TrimHtml(aJson.Content)

	switch aJson.Cmd {

	case models.EventSendMsg	: //聊天消息

		cSize := len([]rune(content))

		if cSize == 0 {
			return
		}

		if cSize > 150 {
			client.WsSendMsg(models.EventToast,"消息太长、请分多次发送",nil)
			return
		}

		client.SendMsg(aJson)

	case models.EventBetMoney	:
		client.AddBet(aJson)
	}

}


func (client *AClient) SendMsg (aJson ClientMsg) {
	size := len([]rune(aJson.Content))

	if size == 0 {
		return
	}

	if size > 150 {
		client.WsSendMsg(models.EventToast,"字数太长",nil)
		return
	}

	//并发检测
	UniqueKey := fmt.Sprintf("Request|%v",aJson.Cmd)

	_,err := CUtil.ReqAnti(client.UserId,UniqueKey, 50,"PX")
	if err != nil {
		client.WsSendMsg(models.EventToast,"请勿频繁操作",nil)
		return
	}

	//过滤铭感词
	content := models.SensitiveAdmin.Replace(aJson.Content,'*')

	//构建消息事件
	status,aEvent := models.NewEvent(client.UserId,aJson.Cmd,content,nil)

	if status {
		PublishEvent[client.RoomId] <- &aEvent

		//写入数据库
		go models.SaveChatLog(client.RoomId,client.UserId,content,aJson.Cmd)
	}
}

//聊天室发文件 图片/语音/视频
func (client *AClient) SendChatFile (aLink string,cmd models.EventType) {

	//并发检测
	UniqueKey := fmt.Sprintf("Request|%v",cmd)

	_,err := CUtil.ReqAnti(client.UserId,UniqueKey, 50,"PX")
	if err != nil {
		client.WsSendMsg(models.EventToast,"请勿频繁操作",nil)
		return
	}

	//构建消息事件
	status,aEvent := models.NewEvent(client.UserId,cmd,aLink,nil)

	if status {
		PublishEvent[client.RoomId] <- &aEvent

		//写入数据库
		go models.SaveChatLog(client.RoomId,client.UserId,aLink,cmd)
	}
}

/**
 * 下注
 *
 * param: ClientMsg aJson
 */
func (client *AClient) AddBet(aJson ClientMsg) {
	var BetInfo AddBetInfo
	err := json.Unmarshal([]byte(aJson.Content),&BetInfo)
	if err != nil {
		logs.Info("===下注数据异常===",err)
		client.WsSendMsg(models.EventToast,"下注数据异常",nil)
		return
	}

	err = models.LhjGameAddBet(client.UserId,BetInfo.Id,BetInfo.Add)
	if err != nil {
		client.WsSendMsg(models.EventToast,err.Error(),nil)
		return
	}


	var ret models.LhjBetInfoShow

	info 		:= models.LhjGetLastRoundInfo()
	ret.RoundId  = info.RoundId
	ret.Status   = info.Status
	ret.TopList  = info.TopList
	ret.TotalBet = models.TotalBet(ret.RoundId) //下发总下注数
	ret.EndBox   = EndBox


	//给所有参与者 下发下注信息
	for _,aRoom := range RoomMap {
		if aRoom == nil {
			continue
		}

		aRoom.Range(func(key, val interface{}) bool {

			AClient := val.(*AClient)

			if AClient != nil {

				//下发自己的信息
				ret.MyBet = models.GetMyBet(AClient.UserId,ret.RoundId)
				aUser    := models.GetOneByUid(AClient.UserId)
				ret.Gold  = aUser.Gold

				//下发总下注数
				status,aEvent := models.NewEvent(client.UserId,models.EventBetMoney,"下注信息",ret)

				if status {
					AClient.WsSendMsg(models.EventBroadcast,"OK",aEvent)
				}
			}

			return true
		})
	}
}
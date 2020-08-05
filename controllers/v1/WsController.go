/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/5/22
 * Time: 14:13
 */
package v1

import (
	"app/icu/CUtil"
	"app/icu/config"
	"app/icu/models"
	"app/icu/service"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type WsController struct {
	BaseController
}

var (

	UpGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool {

			if r.Method != "GET" {
				return false
			}

			return true
		},
	}
)

func (this *WsController) WsConnect() {

	ws,err := UpGrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)

	if _, ok := err.(websocket.HandshakeError); ok {
		logs.Error("upgrader err:%v", err)
		return
	}

	if err != nil {
		logs.Error("socket 连接失败",err)
		return
	}

	UserId   := this.GetUserId(false)

	sessid	 := this.Ctx.GetCookie("sessid")

	ThisClient := service.AClient{Ws:ws,UserId:UserId,RoomId:config.PublicRoomId,IsAlive:1}

	// 定义最后退出句柄
	// 例如 任何地方调用了 client.Ws.Close() 都会从 ReadMessage 抛出异常 从而结束本次协程
	defer ThisClient.LeaveRoom()


	status,_ := models.CheckLogin(UserId,sessid,"")

	if !status {
		cmd := models.EventNoLogin

		ThisClient.S2CMsg(cmd,"Session已过期,请重新登录",nil,models.EventExist,time.Second / 10)

		return
	}


	nums,_ := CUtil.ReqAntiTimes(UserId,"WsConnect",1,3600,"SET")
	if nums >  models.LoginTimesByHour  {

		CUtil.ReqAntiTimes(UserId,"WsConnect",-1,3600,"SET")

		cmd   := models.EventToast

		ttl,_ := CUtil.ReqAntiTimes(UserId,"WsConnect",0,0,"TTL")


		msg   := fmt.Sprintf("频繁登录退出，请 %ds 候再来",ttl)

		ThisClient.S2CMsg(cmd,msg,nil,models.EventExist,time.Second / 10)
	} else {

		hasJoin,oldClient := service.HasJoinRoom(UserId)

		if hasJoin {
			oldClient.S2CMsg(models.EventRepeatJoin,"账号已在别处登录",nil,models.EventExist,time.Second/5)
		}

	}

	// Join chat room.
	err = ThisClient.JoinRoom()
	if err != nil {
		//todo 从处理错误信息的房间退出
		ThisClient.S2CMsg(models.EventToast,err.Error(),nil,models.EventExist,time.Second/10)
		return
	}

	ThisClient.ReadMessage()

}


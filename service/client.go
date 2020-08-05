/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/3/31
 * Time: 20:17
 */
package service

import (
	"app/icu/CUtil"
	"app/icu/models"
	"bytes"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"time"
)

type AddBetInfo struct {
	Id  string
	Add uint
}

//客户端发送过来的消息结构
type ClientMsg struct {
	Cmd      models.EventType `json:"cmd"`
	Content  string `json:"content"`
}



//消息监听
func (client *AClient)ReadMessage() {

	client.Ws.SetReadLimit(MaxMessageSize)

	//客户端断网后最大可允许链接时长,比如欠费，进入电梯，开启飞行模式的外部断网
	//手机锁屏则直接触发链接关闭事件
	client.Ws.SetReadDeadline(time.Now().Add(PongWait))


	// 设置每个请求的时间，超过这个时间直接返回错误
	client.Ws.SetPongHandler(func(string) error {

		logs.Info("====SetPongHandler======")

		//写数据超时时间延长 n s
		err := client.Ws.SetWriteDeadline(time.Now().Add(WriteWait))

		if err != nil {
			logs.Info("SetWriteDeadline 心跳包超时 user_id[%d] err:%v", client.UserId, err)
		}

		//读数据超时时间延长 n s
		err = client.Ws.SetReadDeadline(time.Now().Add(PongWait))

		if err != nil {
			logs.Info("SetReadDeadline 心跳包超时 user_id[%d] err:%v", client.UserId, err)
		}

		return nil
	})

	//ping
	go client.Ping()

	// Message receive loop.
	for {

		// 一旦调用 Ws.Close() 立刻返回错误信息 直接退出
		_, message, err := client.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Error("websocket user_id[%d] unexpected close error: %v", client.UserId, err)
			}

			logs.Info("=========ReadMessage======",err)

			return
		}

		//数据过滤
		message = bytes.TrimSpace(bytes.Replace(message, NewLine, Space, -1))

		length := len(message)

		if length <= 0 || length > 1024{
			continue
		}

		client.Request(message)
	}
}


//用户加入房间
func (client *AClient) JoinRoom() error {

	if _ ,ok := RoomMap[client.RoomId] ; !ok {
		return errors.New("房间不存在或已关闭")
	}

	aUser := models.GetOneByUid(client.UserId)

	if aUser.UserId == 0 {
		return errors.New("用户不存在")
	}

	hasJoin,OldClient := HasJoinRoom(client.UserId)
	if hasJoin {
		return errors.New("您已在房间:" + string(OldClient.RoomId))
	}

	//关联用户链接
	_,ok := RoomMap[client.RoomId].LoadOrStore(client.UserId,client)
	if ok {
		return errors.New("请勿重复加入")
	}

	JoinRoomUser[client.RoomId] <- &Subscriber{client.UserId,client.Ws}

	return nil

}

//离开房间 最后的数据清理工作
func (client *AClient) LeaveRoom() (bool,error) {

	defer client.Ws.Close()

	hasJoin,oldClient := HasJoinRoom(client.UserId)

	if hasJoin && oldClient.RoomId != "" {

		RoomMap[oldClient.RoomId].Delete(client.UserId)

		PingMap[oldClient.RoomId].Delete(client.UserId)

		aUser := models.GetOneByUid(client.UserId)

		if aUser != (models.UserInfo{}){
			LeaveRoomUser[oldClient.RoomId] <- &client.UserId
		}
	}

	return true,nil
}

/**
 服务器给指定客户端下发的小
 */
func (client *AClient) S2CMsg(cmd int,msg string,aData interface{},MsgType int,waitTime time.Duration) {

	cIndex       := GetToClientChanIndex(client.UserId)

	ResponseBody := CUtil.FormatApiJson(cmd,msg,aData)

	aMsg         := SendToClientMsg{
		MsgType : MsgType,
		Client  : client,
		MsgBody : ResponseBody,
	}

	//刷入消息通道 待消息引擎发送
	SendToClientChan[client.RoomId][cIndex] <- &aMsg

	//休眠一会 等待消息收发完毕
	if waitTime > 0 {
		time.Sleep(waitTime)
	}
}

//发送一个Ping数据
func (client *AClient)WsSendPing(data []byte) {

	cIndex  := GetToClientChanIndex(client.UserId)

	aMdg    := SendToClientMsg{
		MsgType : websocket.PingMessage,
		Client  : client,
		MsgBody : data,
	}

	//刷入消息通道 待消息引擎发送
	SendToClientChan[client.RoomId][cIndex] <- &aMdg
}



//发送一条普通数据
func (client *AClient) WsSendMsg(cmd int, msg string ,data interface{}) {

	cIndex       := GetToClientChanIndex(client.UserId)

	ResponseBody := CUtil.FormatApiJson(cmd,msg,data)

	aMsg         := SendToClientMsg{
		MsgType : websocket.TextMessage,
		Client  : client,
		MsgBody : ResponseBody,
	}

	//刷入消息通道 待消息引擎发送
	SendToClientChan[client.RoomId][cIndex] <- &aMsg
}
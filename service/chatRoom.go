package service

import (
	"app/icu/config"
	"app/icu/models"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Subscriber struct {
	UserId uint64
	Conn *websocket.Conn //WebSocket users
}


type AClient struct {
	Ws 		*websocket.Conn
	UserId 	uint64
	RoomId 	config.RoomIdType
	IsAlive uint8
}

//发送到客户的 缓存通道数据，防止并发问题
type SendToClientMsg struct {
	MsgType   int
	Client    *AClient
	MsgBody   interface{}
}

//进出房间扩展消息
type JoinLeaveEvent struct {
	RoomUserCount int   `json:"room_user_count"` //聊天室当前人数
}

//送礼物奖品事件
type AwardInfo struct {
	Name   string   `json:"name"`
	Num    uint8    `json:"num"`
	Img    string   `json:"img"`
}

const (
	WriteWait      = 10 * time.Second
	MaxMessageSize = 512
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	ClientChanNum  = 10 //消息引擎数量
)

//全局变量
var (
	NewLine          = []byte{'\n'}
	Space            = []byte{' '}

	//加入聊天室用户
	JoinRoomUser     = make(map[config.RoomIdType]chan *Subscriber)

	//离开聊天室用户
	LeaveRoomUser    = make(map[config.RoomIdType]chan *uint64)

	//聊天室和用户
	RoomMap          = make(map[config.RoomIdType] *sync.Map)

	//消息实体
	PublishEvent     = make(map[config.RoomIdType]chan *models.EventInfo)

	// 每个聊天室开 ClientChanNum个 消息转发通道
	// 防止并发情况往同一个写数据 出现 panic: concurrent write to websocket connection
	SendToClientChan = make(map[config.RoomIdType]map[int]chan *SendToClientMsg)

	//ping 消息
	PingMap          = make(map[config.RoomIdType] *sync.Map)
)

func init()  {
	initRoomList()
	launchChatRoom()
}

//根据用户ID 确定消息发送通道索引号
func GetToClientChanIndex(UserId uint64) int {

	mod := UserId % ClientChanNum

	return int(mod)
}

//todo 初始化聊天室房间
func initRoomList()  {
	RList := config.RoomList.Data

	if len(RList) == 0 {
		panic("未配置聊天室")
	}

	for RoomId := range RList {
		JoinRoomUser[RoomId]     = make(chan *Subscriber)
		LeaveRoomUser[RoomId]    = make(chan *uint64)
		PublishEvent[RoomId]     = make(chan *models.EventInfo)
		SendToClientChan[RoomId] = make(map[int]chan *SendToClientMsg)
		RoomMap[RoomId]          = &sync.Map{}
		PingMap[RoomId]          = &sync.Map{}

		for index:=0; index<ClientChanNum; index++ {
			SendToClientChan[RoomId][index] = make(chan *SendToClientMsg)
		}
	}

	logs.Info("======聊天室初始化完成======")
}

//todo 启动聊天室
func launchChatRoom() {
	RoomList := config.RoomList.Data

	if len(RoomList) == 0 {
		panic("未配置聊天室")
	}

	for RoomId := range RoomList {

		// 聊天室消息生产者
		go roomInfoProducer(RoomId)

		//消费者
		go roomInfoConsumer(RoomId)

		//每个聊天室启动10个协程作为数据输出引擎,以防并发模式处理对客户端的消息转发
		for index:=0; index<ClientChanNum; index++ {
			go sendMsgEngine(RoomId,index)
		}
	}

	logs.Info("======聊天室消息收发引擎初始化完成======")
}

//todo 只允许初始化时执行！！！！  PublishEvent[RoomId] <- ****
func roomInfoProducer(RoomId config.RoomIdType) {
	for {
		select {
		case sub,ok := <-JoinRoomUser[RoomId] :
			if !ok {
				logs.Info("==JoinRoomInfoProducer==",sub)
				continue
			}

			JoinRoomBroadcast(RoomId,*sub)

		case uid,ok := <-LeaveRoomUser[RoomId] :
			if !ok {
				logs.Info("==LeaveRoomInfoProducer==",uid)
				continue
			}

			LeaveRoomBroadcast(RoomId,*uid)
		}
	}
}


//todo 只允许初始化时执行！！！！  <- PublishEvent[RoomId]
func roomInfoConsumer(RoomId config.RoomIdType)  {
	for {
		select {
		case event ,ok := <-PublishEvent[RoomId] :

			if !ok {
				logs.Info("=RoomInfoConsumer==",event)
				continue
			}

			PublishBroadcast(RoomId,event)
		}
	}
}


//todo 消息引擎 每个聊天室 ClientChanNum 个 只允许初始化时执行
func sendMsgEngine(RoomId config.RoomIdType,index int) {

	for {
		select {
		case msg := <-SendToClientChan[RoomId][index] :
			switch msg.MsgType {

			case models.EventExist : fallthrough
			case websocket.TextMessage :
				msgByte, err := json.Marshal(msg.MsgBody)

				if err != nil {
					logs.Error("send msg [%v] marsha1 err:%v", string(msgByte), err)
					continue
				}


				err = msg.Client.Ws.SetWriteDeadline(time.Now().Add(WriteWait))
				if err != nil {
					logs.Error("send msg SetWriteDeadline [%v] err:%v", string(msgByte), err)
					continue
				}

				w, err := msg.Client.Ws.NextWriter(websocket.TextMessage)
				if err != nil {
					cErr := msg.Client.Ws.Close()
					logs.Info("ERR NextWriter :",err,msg.Client.UserId,cErr)
					continue
				}

				_,err = w.Write(msgByte)
				if err != nil {
					logs.Error("Write msg [%v] err: %v",string(msgByte),err)
				}

				if err := w.Close(); err != nil {
					cErr := msg.Client.Ws.Close()
					logs.Info("ERR Close :",err,msg.Client.UserId,cErr)
					continue
				}

				if msg.MsgType == models.EventExist {
					logs.Info("logout",&msg.Client.UserId)
					msg.Client.Ws.Close()
				}

			case websocket.PingMessage :

				if err := msg.Client.Ws.WriteMessage(websocket.PingMessage, msg.MsgBody.([]byte)); err != nil {

					logs.Info("心跳异常 user_id[%d] err:%v",msg.Client.UserId, err)

					msg.Client.IsAlive = 0

					msg.Client.Ws.Close()
				}
			}
		}
	}
}


//全网广播
func PublishBroadcast(RoomId config.RoomIdType, event *models.EventInfo) {

	aRoom := RoomMap[RoomId]

	switch event.Type {
	case models.EventSendMsg       : fallthrough
	case models.EventBroadcast     : fallthrough
	case models.EventJoin          : fallthrough
	case models.EventClientSendAwd : fallthrough
	case models.EventGameBetNoWin  : fallthrough
	case models.EventSendChatImg   : fallthrough
	case models.EventSendChatAudio : fallthrough
	case models.EventSendChatVedio : fallthrough
	case models.EventLeave         :
		aRoom.Range(func(key, val interface{}) bool {

			client := val.(*AClient)

			//离开房间消息 不用发给自己
			if (event.Type == models.EventLeave) && (client.UserId == event.UserId) {
				return true
			}

			client.WsSendMsg(1,"OK",event)

			return true
		})

	}
}


//检测用户是否加入过房间
func HasJoinRoom(UserId uint64) (ret bool,client *AClient) {
	for _,syncMap := range RoomMap {
		val,ok := syncMap.Load(UserId)

		if !ok {
			continue
		}

		return true,val.(*AClient)
	}

	return false,nil
}

//获取某个聊天室总用户数
func GetRoomUserCount(RoomId config.RoomIdType) int {
	 if sMap,ok := RoomMap[RoomId]; !ok {
		 return 0
	 } else {
	 	 var count = 0
		 sMap.Range(func(key, value interface{}) bool {
			 count++
			 return true
		 })

	 	 return count
	 }
}

//加入房间广播事件
func JoinRoomBroadcast(RoomId config.RoomIdType,sub Subscriber) {

	//构建加入事件
	count        := GetRoomUserCount(RoomId)
	JoinRoomUser := JoinLeaveEvent{RoomUserCount:count}

	status,event := models.NewEvent(sub.UserId, models.EventJoin, "登录了系统", JoinRoomUser)
	if !status {
		return
	}

	PublishEvent[RoomId] <- &event

}

//离开聊天室
func LeaveRoomBroadcast(RoomId config.RoomIdType,UserId uint64) {
	//构建离开事件
	count         := GetRoomUserCount(RoomId)
	LeaveRoomUser := JoinLeaveEvent{RoomUserCount:count}
	status,event  := models.NewEvent(UserId, models.EventLeave, "离开了系统", LeaveRoomUser)
	if !status {
		return
	}

	PublishEvent[RoomId] <- &event
}


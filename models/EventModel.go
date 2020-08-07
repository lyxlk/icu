package models

import (
	"github.com/importcjj/sensitive"
	"time"
)

type EventType int32

// 消息类型客户端和服务器公用命令字 0 - 20000 范围
// 客户端专用命令字 范围 100000 - 10000000
// 服务器专用命令字 范围 10000000 -
const (
	EventExist                  = -10
	EventToast                  = -1
	EventNoLogin                = -100
	EventBroadcast              = 1
	EventJoin                   = 10000
	EventLeave                  = 10001
	EventSendMsg                = 10002
	EventRepeatJoin             = 10003

	EventClientSendAwd          = 100000
	EventBetMoney             	= 100001

	EventGameBetNewRound        = 10000000
	EventGameBetNoWin           = 10000001
	EventGameBetWin             = 10000002
	EventGameBetLuck            = 10000003
	EventSendChatImg            = 10000004
	EventSendChatAudio          = 10000005
	EventSendChatVedio          = 10000006

)

var (

	// 定义服务器响应事件

	CmdMap = map[EventType]string{
		EventExist                    : "强制退出聊天室",
		EventToast                    : "公共错误信息",
		EventNoLogin                  : "Session已过期",
		EventBroadcast                : "广播消息",
		EventJoin                     : "进入聊天室",
		EventLeave                    : "离开聊天室",
		EventSendMsg                  : "发送消息",
		EventRepeatJoin               : "重复加入聊天室",

		EventClientSendAwd            : "送礼物",
		EventGameBetNewRound          : "新一轮开始",
		EventBetMoney                 : "押注",
		EventGameBetNoWin             : "未中奖",
		EventGameBetWin               : "押注中奖",
		EventGameBetLuck              : "开奖",
		EventSendChatImg              : "发聊天图片",
		EventSendChatAudio            : "发语音链接",
		EventSendChatVedio            : "发视频链接",
	}

	//铭感词管理员
	SensitiveAdmin *sensitive.Filter
)

//消息结构
type EventInfo struct{
	Type       EventType      `json:"type"`        // 消息事件类型
	UserId     uint64         `json:"user_id"`     //发消息用户
	UserLevel  uint8          `json:"user_level"`  //发消息用户
	Nick       string         `json:"nick"`       //发消息用户昵称
	Avatar     int         	  `json:"avatar"`     //发消息用户头像
	Time       int64          `json:"time"`       //发消息时间
	Content    string         `json:"content"`   //消息内容
	Ext        interface{}    `json:"ext"`       //扩展信息 备用字段
}

/**
 * 创建一个消息结构实体
 *
 * param: uint64      UserId
 * param: EventType   cmd
 * param: string      content
 * param: interface{} ext
 */
func NewEvent(UserId uint64,cmd EventType,content string,ext interface{})  (bool,EventInfo) {

	_,ok := CmdMap[cmd]
	if !ok {
		return false,EventInfo{}
	}

	//过滤H5标签
	//content = CUtil.TrimHtml(content)

	aUser  := GetOneByUid(UserId)
	cTime  := time.Now().Unix()

	return true, EventInfo {
		Type       : cmd,
		UserId     : aUser.UserId,
		Nick       : aUser.Nick,
		Avatar     : aUser.Avatar,
		Time       : cTime,
		Content    : content,
		Ext        : ext,
	}
}
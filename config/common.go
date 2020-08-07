/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/27
 * Time: 17:52
 */
package config

import (
	"sync"
)

type RoomIdType string

type RoomInfoStruct struct {
	Rid			RoomIdType 	`json:"rid"`
	Slogan 		string 		`json:"slogan"`
	Player 		uint8  		`json:"player"`
	Viewer 		uint8  		`json:"viewer"`
	Background 	string		`json:"background"`
	Type	 	uint8		`json:"type"`
}

type RoomListStruct struct{
	Data map[RoomIdType]RoomInfoStruct
	Lock sync.RWMutex
}


const (

	TimeZone 					= "Asia/Shanghai"

	BaseYmdHis					= "2006-01-02 15:04:05"
	BaseYmd						= "2006-01-02"
	BaseYmdHisNoFix				= "20060102 150405"
	BaseYmdNoFix				= "20060102"
	BaseYm						= "2006-01"
	BaseYmNoFix					= "200601"
	SessionExpire				= 3600 //session 有效期
	DefaultNick 				= "农码倥偬用户"
	QrCodeSize 					= 10240   //二维码文件大小限制 10kB
	UploadSize 					= 2097152 //聊天图片大小限制 2M

	DefaultAge 					= 18 //默认年龄
	AvatarNums 					= 36 //系统头像数量

	PublicRoomId  RoomIdType    = "R000"

	GoldTypeLogAddBet			= 10 //下注
	GoldTypeLogAddBetWin		= 11 //下注赢了
	GoldTypeLogBankrupt 		= 12 //下注赢了
)


var (

	OnlyUploadImgExt = []string{".jpg",".png",".gif"}

	CommSalt = []byte("abcdE123&^*^&*@&") //公共可逆加密盐值 注意长度符合 aes.NewCipher 要求

	//创建订单号标识
	OrderActIds = map[string]int {
		"Reg" : 1, //注册
		"Qr"  : 2, //生成二维码
	}

	UserFrom = map[string]uint8 {
		"qrcode"   		 : 1,
	}

	//登录方式
	LoginControl = map[uint8]string {
		1 : "qrcode", //二维码登录
	}

	//免权限验证URL
	NoCheckMap = map[string]string {
		"/fight/login/index" 		: "Login",
		"/fight/login/reg" 			: "Reg",
		"/fight/login/auto" 		: "Auto",
		"/fight/login/out" 		    : "LogOut",
		"/fight/room/list" 			: "RoomList",
		"/fight/ws/server"	 		: "WS",
	}

	//房间信息 R000
	RoomList = RoomListStruct{
		Data: map[RoomIdType] RoomInfoStruct {

			PublicRoomId : {
				Rid			:	"R000",
				Slogan		:	"在线江湖",
				Player		:	0,
				Viewer		: 	200,
				Background	:	"",
				Type		:   0,//群聊房间，1对战房间
			},

			"R001" : {
				Rid			:	"R001",
				Slogan		:	"老头老太来摔跤🤺",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#bce8f1",
				Type		:   1,//群聊房间，1对战房间
			},

			"R002" : {
				Rid			:	"R002",
				Slogan		:	"直男不打基佬😍",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#b1d4ef",
				Type		:   1,//群聊房间，1对战房间
			},

			"R003" : {
				Rid			:	"R003",
				Slogan		:	"不成功 便成鬼😎😎",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#00CCCC",
				Type		:   1,//群聊房间，1对战房间
			},

			"R004" : {
				Rid			:	"R004",
				Slogan		:	"聚是一坨屎，散是满天稀💨",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#FFCC99",
				Type		:   1,//群聊房间，1对战房间
			},

			"R005" : {
				Rid			:	"R005",
				Slogan		:	"士可杀，你侮辱我时小心点🤣",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#FFCCCC",
				Type		:   1,//群聊房间，1对战房间
			},

			"R006" : {
				Rid			:	"R006",
				Slogan		:	"拿小皮鞭抽我啊~🤞🤞🤞",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#CCCCFF",
				Type		:   1,//群聊房间，1对战房间
			},

			"R007" : {
				Rid			:	"R007",
				Slogan		:	"比武招亲,基佬也行🍳",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#99CC99",
				Type		:   1,//群聊房间，1对战房间
			},

			"R008" : {
				Rid			:	"R008",
				Slogan		:	"左青龙，右白虎，老牛在腰间",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#66CCCC",
				Type		:   1,//群聊房间，1对战房间
			},

			"R009" : {
				Rid			:	"R009",
				Slogan		:	"来了老弟😘😘😘😘",
				Player		:	2,
				Viewer		: 	100,
				Background	:	"#FFCCFF",
				Type		:   1,//群聊房间，1对战房间
			},
		},

		Lock:sync.RWMutex{},
	}

	//下注字段 0.871概率 0.129啥都没有
	LhjBetTxt = map[string]int{
		"lhj_bet_txt_bar" 			: 1, // 0.001
		"lhj_bet_txt_seven" 		: 2, // 0.01  * 20
		"lhj_bet_txt_star" 			: 3, // 0.1   * 20
		"lhj_bet_txt_watermelons" 	: 4, // 0.01  * 30
		"lhj_bet_txt_alarm" 		: 5, // 0.01  * 5
		"lhj_bet_txt_coconut" 		: 6, // 0.15  * 10
		"lhj_bet_txt_orange" 		: 7, // 0.15  * 10
		"lhj_bet_txt_apple" 		: 8, // 0.44  * 5
	}

	//下注字段中奖概率
	LhjBetTxtRate = map[int]float64 {
		0 : 0.1,
		1 : 0.001,
		2 : 0.014,
		3 : 0.015,
		4 : 0.05,
		5 : 0.15,
		6 : 0.25,
		7 : 0.25,
		8 : 0.17,
	}

	//中奖后再抽位置
	RandIndexRate = map[int]map[int]float64 {
		0 : {
            6  : 0.5, //不中奖
            18 : 0.5,
		},
		1 : {
			24 : 0.1,
			23 : 0.9,//小bar
		},
		2 : {
			12 : 0.2,
			11 : 0.8,//小77
		},
		3 : {
			17 : 0.8,//小start
			16 : 0.2,
		},
		4 : {
			5 : 0.9,//小西瓜
			4 : 0.1,
		},
		5 : {
			22 : 0.2,
			20 : 0.6,//小铃铛
			10 : 0.2,
		},
		6 : {
			15 :0.1,
			14 :0.8,//小木瓜
			3 : 0.1,
		},
		7 : {
			21 : 0.25,
			9  : 0.25,
			8  : 0.50, //小橘子
		},
		8 : {
			1  : 0.2,
			2  : 0.5,//小苹果
			7  : 0.1,
			13 : 0.1,
			19 : 0.1,
		},
	}

	//奖励倍数
	LhjPrizeBei = map[int]int{
		24 : 100, //大BAR
		23 : 50,  //小BAR
		22 : 20,  //大铃铛
		21 : 10,  //大橘子
		20 : 2,   //小铃铛
		19 : 5,   //大苹果
		18 : 0,   //CHA
		17 : 2,   //小star
		16 : 30,  //大star
		15 : 10,  //大木瓜
		14 : 2,   //小木瓜
		13 : 5,   //大苹果
		12 : 40,  //大77
		11 : 2,   //小77
		10 : 20,  //大铃铛
		9  : 10,  //大橘子
		8  : 2,   //小橘子
		7  : 5,   //大苹果
		6  : 0,   //CHA
		5  : 2,   //小西瓜
		4  : 25,  //大西瓜
		3  : 10,  //大木瓜
		2  : 2,   //小苹果
		1  : 5,   //大苹果
	}

	LhjBetList = map[int]string{
		24 : "b_bar", //大BAR
		23 : "s_bar",  //小BAR
		22 : "b_alarm",  //大铃铛
		21 : "b_orange",  //大橘子
		20 : "s_alarm",   //小铃铛
		19 : "b_apple",   //大苹果
		18 : "cha",   //CHA
		17 : "s_star",   //小star
		16 : "b_star",  //大star
		15 : "b_coconut",  //大木瓜
		14 : "s_coconut",   //小木瓜
		13 : "b_apple",   //大苹果
		12 : "b_77",  //大77
		11 : "s_77",   //小77
		10 : "b_alarm",  //大铃铛
		9  : "b_orange",  //大橘子
		8  : "s_orange",   //小橘子
		7  : "b_apple",   //大苹果
		6  : "cha",   //CHA
		5  : "s_watermelons",   //小西瓜
		4  : "b_watermelons",  //大西瓜
		3  : "b_coconut",  //大木瓜
		2  : "s_apple",   //小苹果
		1  : "b_apple",   //大苹果
	}

	LhjGameConfig = struct {
		AddTime int
		CountTime int
		WaitTime int
	}{
		AddTime   : 25, //下注时间
		CountTime : 10, //结算时间
		WaitTime  : 5, //下一局开启时间
	}

	CreateOrderIdActs = map[string]int {
		"SetCoin" 	   : 1,  //加减资产
		"SpreadLink"   : 2,  //渠道拉新
	}
)


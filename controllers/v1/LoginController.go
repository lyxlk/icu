/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/27
 * Time: 17:31
 */
package v1

import (
	"app/icu/CUtil"
	"app/icu/config"
	"app/icu/models"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/tuotoo/qrcode"
	"time"
)

type LoginController struct {
	BaseController
}

//登录信息
type LoginInfo struct {
	UserId			uint64		`json:"user_id"`
	IsRegister		uint8		`json:"isRegister"`
	RegTime			string		`json:"reg_time"`
	Avatar			int			`json:"avatar"`
	Age				uint8		`json:"age"`
	Sex				uint8		`json:"sex"`
	From			uint8		`json:"from"`
	Nick			string		`json:"nick"`
	Qrcode  		string		`json:"qrcode"`
	Fail			int			`json:"fail"`
	Win				int			`json:"win"`
	Draw			int			`json:"draw"`
	Gold			uint		`json:"gold"`
	Level			string		`json:"level"`

}

func (this *LoginController) afterLogin(UserId uint64,IsRegister uint8,initSess bool) {

	if UserId == 0 {
		this.JsonResponse(-1,"账号不存在登录失败",nil)
	}

	nums,_ := CUtil.ReqAntiTimes(UserId,"WsConnect",0,0,"GET")
	if nums > models.LoginTimesByHour {
		ttl,_ := CUtil.ReqAntiTimes(UserId,"WsConnect",0,0,"TTL")
		msg   := fmt.Sprintf("频繁登录退出，请 %ds 候再来",ttl)
		this.JsonResponse(-1,msg,nil)
	}

	UserInfo := models.GetOneByUid(UserId)

	if UserInfo.UserId == 0 {
		this.JsonResponse(-1,"账号不存在登录失败(2)",nil)
	}

	if initSess {
		var aMap map[string]string
		sessid,err := models.InitUserLoginTarget(UserId,aMap)
		if err != nil {
			this.JsonResponse(-1,err.Error(),nil)
		}

		this.Ctx.SetCookie("sessid",sessid,3600,"/")
	}


	//每日土豪榜
	models.AddTuHaoList(UserId)

	//todo 组装可下发用户信息
	var loginInfo LoginInfo

	regDate 				:= CUtil.GetTheDate(UserInfo.RegTime,"-")
	regClock				:= CUtil.GetTheClock(UserInfo.RegTime,":")

	loginInfo.UserId		= UserInfo.UserId
	loginInfo.IsRegister	= IsRegister
	loginInfo.RegTime	    = fmt.Sprintf("%s %s",regDate,regClock)
	loginInfo.From	    	= UserInfo.From
	loginInfo.Sex		    = UserInfo.Sex
	loginInfo.Age		    = UserInfo.Age
	loginInfo.Avatar		= UserInfo.Avatar
	loginInfo.Nick			= UserInfo.Nick
	loginInfo.Fail			= UserInfo.Fail
	loginInfo.Win			= UserInfo.Win
	loginInfo.Draw			= UserInfo.Draw
	loginInfo.Gold			= UserInfo.Gold
	loginInfo.Level			= models.GetLevel(UserInfo.Gold)
	loginInfo.Qrcode		= models.EncodeUserQrCode(UserId)

	this.JsonResponse(1,"OK",loginInfo)
}

/**
 * 一键注册
 */
func (this *LoginController) Reg()  {

	Ip      := this.GetUserIp()

	timeUid := uint64(time.Now().Unix())

	openid  := CUtil.CreateOrderId(timeUid,config.OrderActIds["Reg"])
	openid   = CUtil.MD5(openid)

	nonce   := string(CUtil.CreateNonceBt(10,0))
	sub     := nonce[0 : 4]
	nick    := fmt.Sprintf("%s%s",config.DefaultNick,sub)

	aUser, err := models.RegisterAUser(Ip,openid,nonce,config.UserFrom["qrcode"],nick,config.DefaultAge)

	if err != nil {
		this.JsonResponse(-1,err.Error(),nil)
	}

	this.afterLogin(aUser.UserId,aUser.IsRegister,true)
}

/**
 * 自动登录校验
 */
func (this *LoginController) Auto() {
	UserId  := this.GetUserId(false)

	sessid	:= this.Ctx.GetCookie("sessid")

	srcIp 	:= this.GetUserIp()

	status,_ := models.CheckLogin(UserId,sessid,srcIp)
	if !status {
		this.JsonResponse(-101,"请登录~",nil)
	}

	this.afterLogin(UserId,0,false)
}

/**
 * 登录
 */
func (this *LoginController) Index()  {

	f, h, err  := this.GetFile("qrcode")
	if err != nil {
		this.JsonResponse(-1,"请使用二维码登陆",nil)
	}

	if  h.Size > config.QrCodeSize {
		this.JsonResponse(-1,"登录用二维码文件过大",nil)
	}

	defer f.Close()

	qrMatrix, err := qrcode.Decode(f)
	if err != nil {
		logs.Error("登录用二维码解析失败",err)
		this.JsonResponse(-1,"登录用二维码解析失败",nil)
	}

	UserId,nonce,_ := models.DecodeUserQrCode(qrMatrix.Content)

	aUser := models.GetOneByUid(UserId)
	if aUser.Nonce != nonce {
		this.JsonResponse(-1,"二维码已失效",nil)
	}

	this.afterLogin(UserId,0,true)
}

/**
 退出系统
 */
func (this *LoginController) LouOut()  {

	err := models.LogOut(this.GetUserId(true))

	if err != nil {
		this.JsonResponse(-1,err.Error(),nil)
	}

	this.JsonResponse(1,"OK",nil)

}
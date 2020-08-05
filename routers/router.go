// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"app/icu/CUtil"
	"app/icu/controllers/v1"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func init() {

	//静态文件 需要设置完整路径
	staticPath := CUtil.GetAppPath("static")
	logs.Critical("===设置静态资源目录完成===",staticPath)

	beego.SetStaticPath("/",staticPath)

	//权限校验
	beego.InsertFilter("/*",beego.BeforeExec, v1.BaseAuth)

	//初始化 namespace
	ns := beego.NewNamespace("/fight",
		beego.NSNamespace("/login",
			beego.NSRouter("/index",&v1.LoginController{},"post:Index"),
			beego.NSRouter("/auto",&v1.LoginController{},"post:Auto"),
			beego.NSRouter("/reg",&v1.LoginController{},"post:Reg"),
			beego.NSRouter("/out",&v1.LoginController{},"post:LouOut"),
		),

		beego.NSNamespace("/home",
			beego.NSRouter("/avatar",&v1.HomeController{},"post:RandAvatar"),
			beego.NSRouter("/modify",&v1.HomeController{},"post:Modify"),
		),

		beego.NSNamespace("/room",
			beego.NSRouter("/bet-info",&v1.RoomController{},"post:BetInfo"),
			beego.NSRouter("/users",&v1.RoomController{},"post:GetUsers"),
			beego.NSRouter("/chat-logs",&v1.RoomController{},"post:ChatLogs"),
			beego.NSRouter("/bankrupt",&v1.RoomController{},"post:Bankrupt"),
			beego.NSRouter("/tuhao",&v1.RoomController{},"post:TuHaoList"),
		),


		beego.NSNamespace("/ws",
			beego.NSRouter("/server",&v1.WsController{},"get:WsConnect"),
		),
	)

	//注册 namespace
	beego.AddNamespace(ns)
}
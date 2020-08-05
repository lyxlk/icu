package main

import  (
	_ "app/icu/SerIni"
	"app/icu/controllers/v1"
	_ "app/icu/crontab"
	_ "app/icu/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
)

func main() {

	defer toolbox.StopTask()

	httpAddr := beego.AppConfig.String("httpaddr")
	httpPort := beego.AppConfig.String("httpport")
	listen   := httpAddr + ":" + httpPort

	logs.Critical("监听地址：%s", listen)

	if beego.BConfig.RunMode == beego.PROD {
		beego.BeeLogger.DelLogger("console")//2
	}

	beego.ErrorController(&v1.ErrorController{})

	beego.Run(listen)
}
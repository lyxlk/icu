/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/5/14
 * Time: 15:53
 */
package crontab

import (
	"app/icu/CUtil"
	"app/icu/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
)

/**
 * 建表计划任务
 *
 * param: string           spec : https://beego.me/docs/module/toolbox.md#spec-%E8%AF%A6%E8%A7%A3
 * param: string           name : 计划任务名称
 * param: toolbox.TaskFunc task : 具体计划
 */
func registerATask(name string,spec string,task toolbox.TaskFunc,debug bool) {

	tk := toolbox.NewTask(name,spec,task)

	// 测试开启运行
	if debug {
		err := tk.Run()
		if err != nil {
			logs.Error(name,err)
			return
		}
	}

	toolbox.AddTask(name,tk)

}

/**
 * 任务配置列表
 */
func gotoTaskLists() {

	//建表
	registerATask("createDbTb"," 0 0 03 * * *", func() error {

		models.CreateChatLogTb()

		_ = models.CreateTbTGoldFlow()

		_ = models.CreateTbBetLog()

		return nil
	},false)



	//凌晨三点清理旧表
	registerATask("ClearLogTb", "0 0 03 * * *", func() error {

		_ = models.CleanTbTGoldFlow()

		_ = models.CleanTbBetLog()

		return nil
	},false)

}


/**
 * 计划任务 : https://beego.me/docs/module/toolbox.md
 * 初始化相关
 *
 */
func init () {

	localIP,err := CUtil.GetLocalIP()

	if err != nil {
		logs.Critical("========= 无效IP地址不加载计划任务",err)
		return
	}

	//只有主服务才启动计划任务脚本
	masterServer := beego.AppConfig.String("masterServer")

	if beego.BConfig.RunMode == beego.PROD {
		if masterServer != localIP{
			logs.Critical("========== 非主服务器停止加载计划任务",err,localIP)
			return
		}
	}

	gotoTaskLists()

	toolbox.StartTask() // 按计划时间执行

	logs.Critical("======= 计划任务加载完成 =============")

}


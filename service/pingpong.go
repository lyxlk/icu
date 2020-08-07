/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/3/31
 * Time: 19:20
 */
package service

import (
	"github.com/astaxie/beego/logs"
	"time"
)

//心跳检测
func (client *AClient) Ping()  {

	//设置可以ping的默认状态
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {

		case <-ticker.C :

			if client.IsAlive == 0 {
				logs.Info("========Not Alive=========",client.UserId)
				return
			}

			client.WsSendPing([]byte("PING"))
		}
	}
}

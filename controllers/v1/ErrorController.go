/**
 * Created by IntelliJ IDEA.
 * User: Kevin
 * Date: 2020/4/29
 * Time: 11:32
 */
package v1

// 定义错误控制器
type ErrorController struct {
	BaseController
}


// 定义404错误, 调用例子: this.Abort("404")
func (this *ErrorController) Error404() {
	this.JsonResponse(-1,"页面未找到(404)",map[string]string{
		"url" : this.Ctx.Request.URL.Path,
	})
}

// 定义500错误, 调用例子: this.Abort("500")
func (this *ErrorController) Error500() {
	this.JsonResponse(-1,"服务器内部错误(500)",nil)
}

func (this *ErrorController) Error502() {
	this.JsonResponse(-1,"502 Bad Gateway",nil)
}

// 定义db错误， 调用例子: this.Abort("Db")
func (this *ErrorController) ErrorDb() {
	this.JsonResponse(-1,"数据库异常(500)",nil)
}


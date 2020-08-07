package v1

import (
	"app/icu/CUtil"
	"app/icu/config"
	"app/icu/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"io"
	"math/rand"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

type BaseController struct {
	beego.Controller
}

/**
 * 上下文分析
 *
 * param: *context.Context ctx
 * param: string           aKey
 * return: string
 */
func analysisContext(ctx *context.Context,aKey string) string {
	switch aKey {
	case "GetIPV4" :

		remoteAddr := ctx.Request.RemoteAddr

		if IPV4 := ctx.Request.Header.Get("X-Real-IP") ; IPV4 != "" {
			remoteAddr = IPV4
		} else if IPV4 = ctx.Request.Header.Get("X-Forwarded-For"); IPV4 != ""  {
			remoteAddr = IPV4
		} else {
			remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
		}

		if remoteAddr == "::1" {
			remoteAddr = "127.0.0.1"
		}

		address := net.ParseIP(remoteAddr)
		if address == nil {
			return ""
		}

		return address.String()

	case "GetUrlPath" :
		url := ctx.Request.URL.Path
		return url
	}

	return ""
}


//中间件session 校验
func BaseAuth (ctx *context.Context) {

	//全局播种
	rand.Seed(time.Now().UnixNano())

	var err error
	path   := analysisContext(ctx,"GetUrlPath")

	UidStr := ctx.GetCookie("user_id")
	if UidStr == "" {
		UidStr = "0"
	}

	UserId,err  := strconv.ParseUint(UidStr,10,64)

	if err != nil {
		ret     := CUtil.FormatApiJson(-1,"无效的用户ID",nil)
		aJson,_ := json.Marshal(ret)
		ctx.ResponseWriter.Write(aJson)
		return
	}


	//todo ------------- 免权限验证URL-------------------------

	_,ok := config.NoCheckMap[path]
	if ok {
		return
	}

	srcIp		:= analysisContext(ctx,"GetIPV4")

	currentKey  := ctx.GetCookie("sessid")

	hasLogin,_  := models.CheckLogin(UserId,currentKey,srcIp)

	if !hasLogin {
		ret     := CUtil.FormatApiJson(-100,"SESSION 已过期",nil)
		aJson,_ := json.Marshal(ret)
		ctx.ResponseWriter.Write(aJson)
		return
	}

	return
}

/**
 *	统一Json输出
 */
func (this *BaseController) JsonResponse (iRet int,sMsg string, data interface{}) {
	this.Data["json"] = CUtil.FormatApiJson(iRet,sMsg,data)
	this.ServeJSON()
	this.StopRun()
}

/**
 * 获取系统当前操作用户ID
 *
 * param: bool stop 出现异常是否终止
 * return: uint64
 */
func (this *BaseController) GetUserId(stop bool) uint64 {
	uidString := this.Ctx.GetCookie("user_id")
	UserId ,err := strconv.ParseUint(uidString,10,64)
	if err != nil {
		if stop {
			this.JsonResponse(-1,err.Error(),nil)
		} else {
			return 0
		}
	}

	return UserId
}

/**
 * 获取远程用户IP
 *
 * return: string
 */
func (this *BaseController) GetUserIp() string {
	return analysisContext(this.Ctx,"GetIPV4")
}


/**
	文件上传
	fileName : file data in file upload field named as key.
	FileDir  : 存放路径
	return: error
	return: string link 相对目录文件路径
 */
func (this *BaseController) UpLoadFileToServer(key string,FileDir string,maxSize int64) (err error,link string) {

	file, handler, err := this.GetFile(key)
	if err != nil {
		return err,""
	}

	defer file.Close()

	if  handler.Size > maxSize {
		ret  := float64(maxSize / 1024 /1024)
		err  := fmt.Sprintf("请上传小于%.1fM图片", ret)
		return errors.New(err),""
	}

	rootPath  := CUtil.GetAppPath("static")

	uploadDir := "/" + FileDir + "/"

	dirExist  := CUtil.CheckFileIsExist(rootPath + uploadDir)

	if !dirExist {
		err := os.Mkdir(rootPath + uploadDir, os.ModePerm) //创建文件夹
		if err != nil {

			return err,""
		}
	}

	now          := time.Now().UnixNano() / 1e6

	fileExt 	 := path.Ext(handler.Filename)

	newFileName  := fmt.Sprintf("%d%s",now,fileExt)

	newFile, err := os.OpenFile(rootPath + uploadDir + newFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err,""
	}

	_,err = io.Copy(newFile, file)

	if err != nil {
		return err , ""
	}

	basePathFile := uploadDir + newFileName

	return nil,basePathFile
}
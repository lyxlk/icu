package SerIni

import (
	"app/icu/CUtil"
	"app/icu/db"
	"app/icu/models"
	"bufio"
	"fmt"
	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/cache/redis"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/importcjj/sensitive"
	"os"
)


func init() {
	InitEnv()
	db.InitRedis()
	InitMysql()
	InitSensitive()
}

//环境初始化
//本地开发 请将SerIni目录下env文件内容设置为0
//开发服  请将SerIni目录下env文件内容设置为 1
//外网服务器请勿在SerIni目录下放置env文件
func InitEnv() {

	var envName string
	var logLevel int

	envFile := CUtil.GetAppPath("SerIni") + "/env"
	//没有env文件则连接正式环境
	if !utils.FileExists(envFile) {
		envFile  = ""
		envName  = beego.PROD
		logLevel = logs.LevelError
	} else {
		//本地开发服和公共开发服判断
		envName = beego.DEV

		//调试查询日志 https://beego.me/docs/mvc/model/overview.md
		orm.Debug = true

		logLevel = logs.LevelInformational
	}

	logs.SetLevel(logLevel)

	beego.BConfig.RunMode = envName

	//设置打印函数及行号
	logs.EnableFuncCallDepth(true)

	logPath := CUtil.GetAppPath("logs")

	logFile := `{"filename":"`+logPath+"/"+envName+`.log"}`

	logs.SetLogger(logs.AdapterFile,logFile)

	logs.Critical("当前初始化环境:%s,环境文件标识：%s,日志文件：%s", envName,envFile,logFile)
}


//todo 创建数据库连接池
func InitMysql() {
	mysqlUser        := beego.AppConfig.String("mysqlUser")
	mysqlPass        := beego.AppConfig.String("mysqlPass")
	mysqlUrls        := beego.AppConfig.String("mysqlUrls")
	mysqlPort        := beego.AppConfig.String("mysqlPort")
	mysqlDbName      := beego.AppConfig.String("mysqlDbName")
	mysqlMaxIdle,err := beego.AppConfig.Int("mysqlMaxIdle")

	if err != nil {
		panic("mysqlMaxIdle ERR:" + err.Error())
	}

	mysqlMaxOpen,err := beego.AppConfig.Int("mysqlMaxOpen")

	if err != nil {
		panic("mysqlMaxOpen ERR:" + err.Error())
	}

	orm.RegisterDriver("mysql",orm.DRMySQL)//注册数据库，数据库为mysql

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4",mysqlUser,mysqlPass,mysqlUrls,mysqlPort,mysqlDbName)

	orm.RegisterDataBase("default", "mysql", dsn)

	//大空闲连接
	orm.SetMaxIdleConns("default", mysqlMaxIdle)

	//最大数据库连接
	orm.SetMaxOpenConns("default", mysqlMaxOpen)

	logs.Info("===Mysql连接池 初始化完成===")
}


//敏感词导入
func InitSensitive() {

	models.SensitiveAdmin = sensitive.New()

	aFile := CUtil.GetAppPath("SerIni") + "/sensitive.txt"

	file, err := os.Open(aFile)
	if err != nil {
		logs.Error("Cannot open text file", aFile, err)
		return
	}

	defer func() {
		file.Close()
		logs.Critical("===敏感词库 加载完成===")
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		models.SensitiveAdmin.AddWord(line)
	}

	if err := scanner.Err(); err != nil {
		logs.Error("Cannot scanner text file", aFile, err)
		return
	}
}
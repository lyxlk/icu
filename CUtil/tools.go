package CUtil

import (
	"app/icu/config"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RspBody struct {
	Ret int `json:"iRet"`
	Msg string `json:"sMsg"`
	Data interface{} `json:"data"`
}

/**
 * 获取本机IP
 *
 * return: string
 * return: error
 */
func GetLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}

	err = errors.New("未找到内网IP")
	return
}

/**
 * 系统分页
 *
 * param: uint page
 * param: uint pageSize
 */
func Pagination(page,pageSize uint) (uint,uint) {

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 1
	}

	offset := pageSize * (page - 1)

	return offset,pageSize
}

/**
 * 系统分页获取总页数
 *
 * param: uint count
 * param: uint pageSize
 * return: uint
 */
func GetPaginationPages(count uint,pageSize uint) uint {
	if pageSize == 0 || count == 0{
		return 1
	}

	pages := math.Ceil(float64(count) / float64(pageSize))

	return uint(pages)
}

/**
 * 类 PHP func_get_args
 */
func FuncGetArgs(args...interface{}) string {

	aString := ""

	for _,val := range args {
		aString += fmt.Sprintf("|%v",val)
	}

	aString = strings.Trim(aString,"|")

	return aString
}

/**
 * 生成唯一订单号
 *
 * param: int64 UserId
 * param: int   actId
 * return: string
 */
func CreateOrderId(UserId uint64,actId int) string {
	msInt       := time.Now().UnixNano() / 1e6 //获取毫秒

	uidStr      := fmt.Sprintf("%010d",UserId) //补齐10位

	actIdStr    := fmt.Sprintf("%02d",actId) //行为ID

	random      := RandInt64(1000,10000) //区间 [1000,9999]

	return fmt.Sprintf("%v_%v_%v_%v",msInt,uidStr,actIdStr,random)
}



//获取指定目录路径
func GetAppPath(dirName string) string {
	appPath := os.Getenv("GOPATH")
	appName := beego.BConfig.AppName

	if dirName == "" {
		return ""
	}

	basePath := filepath.Join(appPath, "src", "app", appName, dirName)

	return basePath
}

/**
 * 标准化系统要输出的 json 格式
 *
 * param: int         iRet
 * param: string      sMsg
 * param: interface{} data
 * return: RspBody
 */
func FormatApiJson(iRet int,sMsg string, data interface{}) RspBody {
	var ResponseBody  RspBody
	ResponseBody.Ret  = iRet
	ResponseBody.Msg  = sMsg
	ResponseBody.Data = data
	return ResponseBody
}


//过滤html标签
func TrimHtml(src string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")
	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")
	return strings.TrimSpace(src)
}

//生成32位MD5
func MD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

//随机整数 返回区间：[min, max) min max 必须大于0
func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	//交由BaseController v1.BaseAuth 控制
	//rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

//随机整数 返回区间：[min, max)  min max 必须大于0
func RandInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	//交由BaseController v1.BaseAuth 控制
	//rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

/**
 * 获取今天凌晨0点时间戳
 *
 * return: int64
 */
func GetTodayStartUnixTime() int64 {
	location, _ := time.LoadLocation("Asia/Shanghai")
	timeStr 	:= time.Now().Format(config.BaseYmd)
	t, _ 		:= time.ParseInLocation(config.BaseYmdHis, timeStr + " 00:00:00", location)
	return t.Unix()
}

/**
 * 获取今天23:59:59时间戳
 *
 * return: int64
 */
func GetTodayEndUnixTime() int64  {
	location, _ := time.LoadLocation("Asia/Shanghai")
	timeStr 	:= time.Now().Format(config.BaseYmd)
	t, _ 		:= time.ParseInLocation(config.BaseYmdHis, timeStr + " 23:59:59", location)
	return t.Unix()
}

/**
 * todo 获取今天起前后N天的时间戳
 *
 * param: int64 day
 * param: int8  opt : 0 按当前时间戳算起 ; 1 按凌晨零点 ; 2 今天23点59分59秒
 * return: int64
 */
func GetTodayBeforeOrAfterDayUnixTime(day int64, opt int8) int64 {
	var st int64
	switch opt {
	case 1 :
		st = GetTodayStartUnixTime()
	case 2 :
		st = GetTodayEndUnixTime()
	default:
		st = time.Now().Unix()
	}

	st  += day * 86400

	return st
}

/**
 * todo 获取今天起前后N天的日期
 *
 * param: int64 day
 * param: int8  opt : 0 按当前时间戳算起 ; 1 按凌晨零点 ; 2 今天23点59分59秒
 * param: uint  format : 0 : 年月日时分秒，1 ：年月日
 * return: string
 */
func GetTodayBeforeOrAfterDayDateTime(day int64, opt int8,format uint) string {
	var st int64
	switch opt {
	case 1 :
		st = GetTodayStartUnixTime()
	case 2 :
		st = GetTodayEndUnixTime()
	default:
		st = time.Now().Unix()
	}

	st  += day * 86400


	var base string
	switch format {
	case 0 : base  = config.BaseYmdHis
	case 1 : base  = config.BaseYmd
	case 2 : base  = config.BaseYmdHisNoFix
	case 3 : base  = config.BaseYmdNoFix
	}

	return time.Unix(st, 0).Format(base)
}

/**
 * todo 获取本月起 前后第N个月日期
 *
 * param: int  m 		需要加减月分数
 * param: int8 opt		0 从本月1号开始计算，1从今天开始计算
 * param: uint format
 * return: string
 */
func GetTodayBeforeOrAfterMonthDateTime(m int,opt int8,format uint ) string {
	year, month, day 	:= time.Now().Date()
	location, _ 		:= time.LoadLocation("Asia/Shanghai")

	var thisMonth time.Time

	switch opt {
	case 1 :
		thisMonth 		= time.Date(year, month, day, 0, 0, 0, 0, location)	 //从今天开始
	default:
		thisMonth 		= time.Date(year, month, 1, 0, 0, 0, 0, location)  //从本月1号开始
	}


	var base string
	switch format {
	case 0 : base  = config.BaseYmNoFix
	case 1 : base  = config.BaseYm
	case 2 : base  = config.BaseYmdNoFix
	case 3 : base  = config.BaseYmd
	}

	date := thisMonth.AddDate(0, m, 0).Format(base)

	return date
}

/**
 * todo 格式化输出某年某月某日
 *
 * param: string fix
 * return: string
 */
func GetTheDate(second int64,fix string) string {

	now := time.Unix(second,0)

	year,month,day := now.Format("2006"), now.Format("01"), now.Format("02")

	return fmt.Sprintf("%s%s%s%s%s",year,fix,month,fix,day)
}

/**
 * todo 格式化某时某分某秒
 *
 * param: string fix
 * return: string
 */
func GetTheClock(second int64,fix string) string {

	now := time.Unix(second,0)

	hour, min, sec := now.Format("15"),now.Format("04"),now.Format("05")

	return fmt.Sprintf("%s%s%s%s%s",hour,fix,min,fix,sec)
}


/**
 * 任意日期转时间戳
 *
 * param: string dateTime
 * param: uint8  format 0 : Ymd ; 1 : Ymd His ; 2 : Y-m-d ; 3 : Y-m-d H:i:s
 * return: int64
 */
func StrToTime(dateTime string,format uint8) int64 {
	location, _ := time.LoadLocation("Asia/Shanghai")

	var layout string
	switch format {
	case 0 :
		layout = config.BaseYmdNoFix
	case 1 :
		layout = config.BaseYmdHisNoFix
	case 2 :
		layout = config.BaseYmd
	case 3 :
		layout = config.BaseYmdHis
	default:
		panic("无效格式")
	}

	ts,_ := time.ParseInLocation(layout,dateTime,location)
	timestamp := ts.Unix()

	return timestamp
}

/**
 * 按字典升降序排列 map[string]string
 *
 * param: map[string]string aMap
 * param: string            opt  : ASC / DESC
 * return: []string
 */
func MapKSort(aMap map[string]string,opt string) []string {
	var  keysArr []string

	if aMap == nil {
		return keysArr
	}

	for k := range aMap {
		keysArr = append(keysArr,k)
	}


	if opt == "DESC" {
		sort.Sort(sort.Reverse(sort.StringSlice(keysArr)))
	} else {
		sort.Strings(keysArr)
	}

	/*for _, k := range keysArr {
		fmt.Println("Key:", k, "Value:", aMap[k])
	}*/

	return keysArr
}


/**
 * 随机字符串
 *
 * param: int size
 * param: int kind 0 : 纯数字 ;  1 : 小写字母 ; 2 : 大写字母 ; 3 : 数字、大小写字母
 * return: []byte
 *
 * 使用时  string([]byte)
 */
func CreateNonceBt(size int, kind int) []byte {

	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)

	isAll := kind > 2 || kind < 0

	rand.Seed(time.Now().UnixNano())

	for i :=0; i < size; i++ {
		if isAll { // random ikind
			ikind = rand.Intn(3)
		}

		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}

	return result
}

/**
 * 发送POST请求
 *
 * param: string url
 * param: string post
 */
func CurlPost (url string , post string,ct string) ([]byte,error) {

	req, err := http.NewRequest("POST",url,strings.NewReader(post))

	if err != nil {
		return nil,err
	}

	req.Header.Set("Content-Type", ct)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Send Req ERR:%s", err.Error())
		return nil,errors.New(errMsg)
	}
	defer resp.Body.Close()


	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		errMsg := fmt.Sprintf("IO ReadAll ERR:%s", err.Error())

		return nil,errors.New(errMsg)
	}

	StatusCode := resp.StatusCode

	if StatusCode != 200 {
		errMsg := "网络异常("+  strconv.Itoa(StatusCode)  +")"
		return nil,errors.New(errMsg)
	}

	return data,nil
}


// 该函数比较两个版本号是否相等，是否大于或小于的关系
// 返回值：0表示v1与v2相等；1表示v1大于v2；2表示v1小于v2
func VersionCompareV1(v1, v2 string) int {
	// 替换一些常见的版本符号
	replaceMap := map[string]string{"V":"","v": "", "-": ".",}
	//keywords := {"alpha,beta,rc,p"}
	for k, v := range replaceMap {
		if strings.Contains(v1, k) {
			strings.Replace(v1, k, v, -1)
		}
		if strings.Contains(v2, k) {
			strings.Replace(v2, k, v, -1)
		}
	}
	ver1 := strings.Split(v1, ".")
	ver2 := strings.Split(v2, ".")
	// 找出v1和v2哪一个最短
	var shorter int
	if len(ver1) > len(ver2) {
		shorter = len(ver2)
	} else {
		shorter = len(ver1)
	}
	// 循环比较
	for i := 0; i < shorter; i++ {
		if ver1[i] == ver2[i] {
			if shorter-1 == i {
				if len(ver1) == len(ver2) {
					return 0
				} else {
					// @todo check for keywords
					if len(ver1) > len(ver2) {
						return 1
					} else {
						return 2
					}
				}
			}
		} else if ver1[i] > ver2[i] {
			return 1
		} else {
			return 2
		}
	}
	return -1
}

/**
 * 版本号比较
 */
func VersionCompareV2(v1, v2 , operator string) bool {
	com := VersionCompareV1(v1,v2)
	switch operator {
	case "==":
		if com == 0 {
			return true
		}
	case "<":
		if com == 2 {
			return true
		}
	case ">":
		if com == 1 {
			return true
		}
	case "<=":
		if com == 0 || com == 2 {
			return true
		}
	case ">=":
		if com == 0 || com == 1{
			return true
		}
	}
	return false
}

/**
 * 按照概率抽奖
 * aMap = map[int]float64 { 1=>0.5, 2=>0.3, 3=>0.2 } (最大精度0.001%)
 * param: map[int]float64 aMap
 */
func GetLuckyKey(aMap map[int]float64) (error,int) {
	var max	   				= decimal.NewFromFloat(100000.0)  //放大10w倍
	var check   			= decimal.NewFromFloat(0.000000)
	var decimalOneFloat 	= decimal.NewFromFloat(1.000000)
	var tEnd   				= decimal.NewFromInt(0)
	var decimalOneInt 		= decimal.NewFromInt(1)
	var tmp 				= make(map[int]map[string]int64)

	for key,val := range aMap {
		tmp[key]	 = make(map[string]int64)

		valDecimal	:= decimal.NewFromFloat(val)
		check		 = check.Add(valDecimal)

		magnify		:= valDecimal.Mul(max)  //放大10w倍

		tmp[key]["rateBegin"] = tEnd.Add(decimalOneInt).IntPart()
		tmp[key]["rateEnd"]	  = tEnd.Add(magnify).IntPart()

		tEnd = tEnd.Add(magnify)
	}

	if check.Cmp(decimalOneFloat) != 0 {

		ret ,_ := check.Float64()

		// 'b' (-ddddp±ddd，二进制指数) 'e' (-d.dddde±dd，十进制指数) 'E' (-d.ddddE±dd，十进制指数)
		// 'f' (-ddd.dddd，没有指数) 'g' ('e':大指数，'f':其它情况) 'G' ('E':大指数，'f':其它情况)
		return errors.New("概率不等于1:" + strconv.FormatFloat(ret,'f', -1, 64)),0
	}


	aRandNum := RandInt64(1,100001) //区间 [1,10000]

	for key,val := range tmp {

		if aRandNum >= val["rateBegin"] && aRandNum <= val["rateEnd"] {
			return nil,key
		}
	}

	return errors.New("随机概率失败"),0
}


//补码
func pKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//去码
func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}


/**
 * 可逆加密
 *
 * param: string orig
 * param: []byte key 详见 aes.NewCipher
 * return: string
 */
func AesEncrypt(orig string, key []byte) string {
	// 转成字节数组
	origData := []byte(orig)

	// 分组秘钥
	block, err  := aes.NewCipher(key)
	if err != nil {
		logs.Error("NewCipher ERR :",err.Error())
		return ""
	}

	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = pKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)

	return base64.StdEncoding.EncodeToString(cryted)

}

/**
 * 可逆解密
 *
 * param: string cryted
 * param: []byte key  详见 aes.NewCipher
 * return: string
 */
func AesDecrypt(cryted string, key []byte) string {
	// 转成字节数组
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		logs.Error("crytedByte ERR :",err.Error())
		return ""
	}

	// 分组秘钥
	block, err := aes.NewCipher(key)
	if err != nil {
		logs.Error("NewCipher ERR :",err.Error())
		return ""
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = pKCS7UnPadding(orig)
	return string(orig)
}

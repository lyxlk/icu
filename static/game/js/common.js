
let LhjGameInstance = {};

const AvatarPath = "/avatars/";

const BetImgPath = "/game/images/fruit/";

const ReConnectTime = 6; //重连次数

//消息时间命令字
let EventCmd = {
    sendMsg : 10002
};


/**
 * 获取用户头像
 * @param avatar
 * @returns {string}
 * @constructor
 */
let GetUserAvatar = function (avatar) {
    return AvatarPath + avatar + ".png";
};


/**
 * 获取押注名称的图片
 * @param betName
 * @returns {string}
 * @constructor
 */
let GetBetImgById = function (betName) {

    return BetImgPath + betName + ".png";
};


//ajax 请求
let adxRequest = function(url, callback, data, method, loading, dataType) {
    let index = 0;
    if(loading !== false) {
        index = layer.msg('加载中');
    }

    $.ajax({
        url: url,
        type: method || 'get',
        dataType: dataType || 'json',
        data: data,
        success: function(data){
            if(parseInt(data.iRet) === -100) {

                window.location.href = "/";

            } else {
                layer.close(index);
                callback(data);
            }

        },
        error: function(xhr, status, msg) {
            let html = '';
            if(msg.indexOf("Invalid JSON")==0) {
                html += "网络连接超时，请稍后再试。";
            } else if(xhr.status==302) {
                html += "没有足够权限访问全部信息";
            } else {
                html  = '<span style="color:#ff0000;font-weight: bolder">HTTP&nbsp;&nbsp;&nbsp;ERROR</span><br />';
                if(xhr.status) {
                    html += '<span style="font-weight: bolder">Message : '+xhr.status+'&nbsp;&nbsp;'+msg+'</span>';
                } else {
                    html += '<span style="font-weight: bolder">网络异常，请重新刷新页面</span>';
                }
            }

            index = layer.open({
                content: html
                ,btn: '我知道了'
                ,time: 2 //2秒后自动关闭
                ,style: 'font-size: 0.25rem;'
            });
        }
    });
};


let TimeAdd0 = function (m){
    m = parseInt(m);

    return m < 10 ? '0' + m : m
};

let UnixTimeFormat = function (unixtime){
    unixtime = parseInt(unixtime);

    let time = new Date(unixtime);
    let y = time.getFullYear();
    let m = time.getMonth()+1;
    let d = time.getDate();
    let h = time.getHours();
    let mm = time.getMinutes();
    let s = time.getSeconds();
    return y + '-' + TimeAdd0(m) + '-' + TimeAdd0(d) +' ' + TimeAdd0(h) + ':' + TimeAdd0(mm) + ':' + TimeAdd0(s);
};


function isMobile() {
    let userAgentInfo = navigator.userAgent;

    let mobileAgents = [ "Android", "iPhone", "SymbianOS", "Windows Phone", "iPad","iPod"];

    let mobile_flag = false;

    //根据userAgent判断是否是手机
    for (let v = 0; v < mobileAgents.length; v++) {
        if (userAgentInfo.indexOf(mobileAgents[v]) > 0) {
            mobile_flag = true;
            break;
        }
    }

    let screen_width = window.screen.width;
    let screen_height = window.screen.height;

    //根据屏幕分辨率判断是否是手机
    if(screen_width < 500 && screen_height < 800){
        mobile_flag = true;
    }

    return mobile_flag;
};
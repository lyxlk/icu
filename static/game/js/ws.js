let websocket = {} ; //单列模式

let connectTimes = 0;

/**
 * 消息到服务器
 * @param Cmd  统一服务器命令字
 * @param Content 要发送的内容 string
 * @returns {boolean}
 * @constructor
 */
let SendMsg = function(Cmd=0,Content='') {

    let user_id  = parseInt($.cookie("user_id"));

    if (user_id <= 0) {
        return false;
    }

    if (typeof (websocket[user_id]) == "undefined") {
        layer.alert("socket服务异常,请刷新重试或加群反馈！");
        return false
    }

    let cond = {Cmd:Cmd,Content:Content};

    websocket[user_id].send(JSON.stringify(cond));

    return true;
};

/**
 * ws初始化
 * @returns {boolean}
 * @constructor
 */
let InitWsConn = function() {

    let user_id  = parseInt($.cookie("user_id"));

    if (user_id <= 0) {
        return false;
    }


    let scheme   = "https:" === document.location.protocol ? "wss": "ws";

    let host     = window.location.host;

    let wsServer = scheme + "://" + host + "/fight/ws/server";

    if (typeof (websocket[user_id]) !== "undefined") {
        console.log("ws已链接");
        return false
    }

    websocket[user_id] = new WebSocket(wsServer); //创建WebSocket对象

    //已经建立连接
    websocket[user_id].onopen = function (evt) {
        HallMsg('<span class="layui-badge-dot layui-bg-orange"></span> <span>---成功连接Socket服务器--- </span>');
    };

    //已经关闭连接
    websocket[user_id].onclose = function (evt) {

        connectTimes ++ ;

        if (connectTimes > ReConnectTime) {
            layer.alert(ReConnectTime + "次重连失败，请检查网络或远程服务是否正常");

            return
        }

        HallMsg('<span class="layui-badge-dot layui-bg-orange"></span> <span style="color: #ff0000">---已断开远程socket服务器,10s尝试重连---</span>');

        delete ( websocket[user_id] );
        setTimeout(function () {
            InitWsConn()
        },10000)
    };

    //收到服务器消息，使用evt.data提取
    websocket[user_id].onmessage = function (evt) {

        let aJson = eval('('+ evt.data +')');

        CommMessage(aJson)
    };

    //产生异常
    websocket[user_id].onerror = function (evt) {
        console.log(evt.data);
    };

   /* setInterval(function () {
        websocket[user_id].send("ping");
    },2000);*/

    return true
};

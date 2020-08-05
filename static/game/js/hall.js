/**
 * 设置游戏消息提醒
 * @param opt
 * @returns {jQuery|*|*}
 * @constructor
 */
let OptGameMsg = function(opt) {
    opt = parseInt(opt);
    switch (opt) {
        case 0 :
        case 1 :
            $.cookie("GameMsgTips",opt,{ expires: 30, path: '/' });
            break;
    }

    let ret = $.cookie("GameMsgTips");

    return parseInt(ret);
};

/**
 * 只有点了游戏页面才把状态设置为1
 */
$(".mainBody .GameGuess").click(function () {
    OptGameMsg(1)
}).siblings().click(function () {
    OptGameMsg(0)
});

//滚动到底部
let scrollToEnd = function (opt){
    opt = parseInt(opt);
    switch (opt) {

        case 1 : //左侧系统消息
            $("#SysMsgList").scrollTop($("#msgListUl")[0].scrollHeight);
            break;
        case 2 : //大厅聊天室
            $(".ChatRoomDiv .pnl-msgs").scrollTop($(".ChatRoomDiv .pnl-msgs")[0].scrollHeight);
            break;
        case 3 : //右侧历史聊天记录
            $(".ChatRoomDiv .chatRoomLogsList").scrollTop($(".ChatRoomDiv .chatRoomLogsList")[0].scrollHeight);
            break
    }
};


/**
 * 所有大厅 左侧消息
 * @param msg
 * @constructor
 */
let HallMsg = function (msg) {

    let len = $('#SysMsgList .msgList').length;
    if(len > 20) {
        $("#SysMsgList #msgListUl").empty();
    }

    let html  = '<li class="msgList">' ;
    html +=  msg;
    html += '</li>';

    $("#SysMsgList #msgListUl").append(html);

    scrollToEnd(1);
};

/**
 * 渲染登录/退出用户
 * @param data
 */
let renderLoginUser = function(data) {
    let html = '<div style="display: inline-block;line-height: 30px">';
    /*  html += '<img src="/avatars/'+data.avatar+'.png" class="layui-nav-img" style="width: 30px;height: 30px;float:left">';*/
    html += '<span class="layui-badge-dot layui-bg-orange"></span> ';
    html += '<span>[ '+data.nick+' ] '+ data.content+'</span>';
    html += '<div>';
    HallMsg(html)
};

/**
 * 渲染聊天消息
 * @param user_id
 * @param nick
 * @param avatar
 * @param unitTime
 * @param content
 * @returns {string}
 * @constructor
 */
let RenderMsg = function (user_id,nick,avatar,unitTime,content) {
    let localUid = parseInt($.cookie("user_id"));

    let contentInfo = contentEmojiParse(content);


    let whosMsg = "";
    if(localUid === user_id) {
        whosMsg = "layim-chat-mine";
    }

    let avatarPath  = GetUserAvatar(avatar);
    let time        = UnixTimeFormat(unitTime * 1000);

    let html    = '<li class="' + whosMsg + '">\n' +
        '               <div class="layim-chat-user">\n' +
        '                   <img src="' + avatarPath +'">\n' +
        '                   <cite><i style="padding:0">'+time+'</i> '+nick+' </cite>\n' +
        '               </div>\n' +
        '               <div class="layim-chat-text" style="max-width: 200px;white-space:normal; word-break:break-all;">'+contentInfo+'</div>\n' +
        '          </li>';

    return html
};

/**
 * 聊天内容转换
 * @param content
 * @returns {*|*|*}
 */
let contentEmojiParse = function(content) {
    return $.emojiParse({
        content: content,
        emojis: [{type: 'qq', path: '/game/images/qq/', code: ':'}, {
            path: '/game/images/tieba/',
            code: ';',
            type: 'tieba'
        }, {path: '/game/images/emoji/', code: ',', type: 'emoji'}]
    });
};

/**
 * 普通房间消息
 * @param data
 * @constructor
 */
let Room000Msg = function(data) {

    let content = contentEmojiParse(data.content);

    let len = $(".ChatRoomDiv .pnl-msgs li").length;
    if(len > 20) {
        $(".ChatRoomDiv .pnl-msgs li").remove();
    }

    let msg = RenderMsg(data.user_id,data.nick,data.avatar,data.time,content);

    $(".ChatRoomDiv .pnl-msgs ul").append(msg);

    scrollToEnd(2);
};


/**
 * 用户加入大厅聊天室
 * @param userId
 * @param avatar
 * @param nick
 * @constructor
 */
let RenderInRoomUser = function (userId,avatar,nick) {

    //先清理再追加
    CleanInRoomUser(userId);

    let img  = GetUserAvatar(avatar);

    let html = ' <li data-uid="' + userId + '" class="pull-left">\n' +
        '           <img src="'+img+'" style="width: 50px;height: 50px;" class="layui-nav-img">\n' +
        '           <div class="code">'+nick+'</div>\n' +
        '        </li>';

    $(".inRoomUsers ul").prepend(html);

    //消息提醒
    if(favicon_flag) {
        clearInterval(favicon_flag_intval);
        let index = 1;
        favicon_flag_intval = setInterval(function () {
            let href = favicon;

            if((++index) % 2) {
                href = favicon2;
                index = index > 10000 ? 0 : index;
            }

            $("#favicon").attr('href',href + "?_=" + Math.random());

        },300);
    }
};

/**
 * 清理房间用户
 * @param userId
 * @constructor
 */
let CleanInRoomUser = function(userId) {
    $(".inRoomUsers li[data-uid='"+userId+"']").remove();
};

/**
 * 展示游戏进度信息
 * @param roundId
 * @param status
 * @constructor
 */
let SetAdvInfo = function(roundId,status) {

    status  = parseInt(status);
    roundId = parseInt(roundId);

    let advInfo = "第 " + roundId + " 局 : ";
    if (status === 0) {
        advInfo += "投注中...";
    } else if( status === 1) {
        advInfo += "开奖中..."
    } else {
        advInfo += "待下一局开始"
    }

    $("#adv").text(advInfo)
};

/**
 * 摇奖
 * @param data
 * @constructor
 */
let SetBetLuckRet = function(data) {

    LaoHuJiGameRunning(data.ext.EndBox,data.ext.LuckId,data.ext.gold);

    SetAdvInfo(data.ext.roundId,data.ext.status)

};

//游戏押注数据
let GameBetInfo = function(data) {

    //下注数据
    LhjGameInstance.InitBetInfoList(data);

    //剩余金币数
    LhjGameInstance.ResetGold(data.gold);

};

//未中奖处理
let ShowLuckNoWinRet = function(data) {

    let ret = OptGameMsg(3);

    if(ret) {
        layer.msg(data.ext.Msg);
    }


    SetAdvInfo(data.ext.roundId,data.ext.status);

    //清空下注界面
    LhjGameInstance.ResetBetInfoList();

    //刷新历史中奖信息
    HistoryAwardLogs(data.ext.LuckLogs);

};

//中奖处理
let ShowLuckWinRet = function(data) {

    let html = "<span style='color: #FF0000;font-weight: bolder'>"+data.ext.Msg+"</span>";

    let ret = OptGameMsg(3);

    if(ret) {
        layer.msg(html, {icon: 6});
    }


    SetAdvInfo(data.ext.roundId,data.ext.status);

    //获得奖励
    LhjGameInstance._money = parseInt(data.ext.Award);
    LhjGameInstance.$('lhj_ben_txt_money').innerHTML = LhjGameInstance._money;

    //清空下注界面
    LhjGameInstance.ResetBetInfoList();

    //刷新历史中奖信息
    HistoryAwardLogs(data.ext.LuckLogs);
};

/**
 * 展示历史中奖记录
 * @param aList
 * @constructor
 */
let HistoryAwardLogs = function(aList) {

    $("#awardLogs dl").empty();

    if(aList.length > 0) {
        $.each(aList,function (index,aData) {
            let src = GetBetImgById(aData.luck_name);
            let html   = '<dd style="margin: 2px;float: left;border: 1px solid #cfcfcf" >';
            html       +=   '<a>';
            html       +=       '<img style="width: 50px;height: 50px;" src="' + src + '">';
            html       +=   '</a>';
            html       += '</dd>';

            $("#awardLogs dl").append(html);
        });
    }
}

/**
 * 统一消息处理
 * @param aJson
 * @constructor
 */
let CommMessage = function (aJson) {

    if(parseInt(aJson.iRet) !== 1) {
        layer.alert(aJson.sMsg);
        return
    }

    let data = aJson.data;

    switch (data.type) {
        case 10000 : //登录系统

            renderLoginUser(data);

            RenderInRoomUser(data.user_id,data.avatar,data.nick);

            break;

        case 10001 : //断线
            CleanInRoomUser(data.user_id);
            break;

        case 100001 : //下注
            GameBetInfo(data.ext);
            break;

        case 10000000 ://新一轮开始
            SetAdvInfo(data.ext.roundId,data.ext.status);
            break;
        case 10002 : //广播各类消息
            Room000Msg(data);
            break;

        case 10000001 : //未中奖
            ShowLuckNoWinRet(data);
            break;

        case 10000002 : //中奖
            ShowLuckWinRet(data);
            break;

        case 10000003 : //摇奖
            SetBetLuckRet(data);
            break;

    }
};


$(document).ready(function() {

    //打开emoji表情后自动滚动到编辑器底部
    $(document).on("click",".emojionearea-filter",function () {
        setTimeout(function () {
            $(".layui-body").animate({
                scrollTop: $('.layui-body').get(0).scrollHeight
            }, 50);

        },50)
    });

    //发消息
    $(document).on("click",'#sendMsg',function () {

        let msg = $("#talkMsg").val().trim();

        let size = msg.length;
        if(size === 0) {
            return
        }

        //太长字符拦截
        if(size > 150) {
            layer.alert("消息太长、请分多次发送");
            return
        }

        SendMsg(EventCmd.sendMsg,msg);

        $("#talkMsg").val('');
    });


    $(document).keyup(function(){
        if(event.keyCode==13){
            //执行函数
            $("#sendMsg").click();
        }
    });


    //进入大厅聊天室
    $(document).on('click',".GameGuess",function () {

        adxRequest("/fight/room/bet-info",function (data) {
            if (parseInt(data.iRet) !== 1) {
                layer.alert(data.sMsg);
                return ;
            }

            let ret = data.data;

            $.cookie("RoundId",ret.round_id);

            //押注界面
            GameBetInfo(ret);

            //游戏进度
            SetAdvInfo(ret.round_id,ret.status);

            //历史中奖记录
            HistoryAwardLogs(ret.luck_logs)

        },null,"POST")
    });


    //进入大厅聊天室
    $(document).on('click',".CommRoomChat",function () {

        adxRequest("/fight/room/users",function (data) {
            if (parseInt(data.iRet) !== 1) {
                layer.alert(data.sMsg);
                return ;
            }

            $.each(data.data,function (kev,val) {
                RenderInRoomUser(val.user_id,val.avatar,val.nick);
            });


            //重置聊天室高度
            let chatRoomHeight = parseInt($('.ChatRoomDiv').height()) -  parseInt($('.editorDiv').height()) - 10;

            if(chatRoomHeight > 0) {
                $('.pnl-msgs').css({'height':chatRoomHeight + "px"});
            }


        },null,"POST")
    });


    //大厅聊天室 历史记录
    $(document).on('click',".chatRoomLogs",function () {

        $(".ChatRoomDiv .chatRoomLogsList ul>li").remove();

        let chatRoomLogsPage = 1;

        $.cookie("chatRoomLogsPage",chatRoomLogsPage);

        RenderChatRoomLogsList(chatRoomLogsPage);

        //首次加载滚到最下面
        setTimeout(function () {
            scrollToEnd(3);
        },200);
    });

    //历史记录 点击更多
    $(document).on('click','.MoreChatRoomLogs',function () {
        let chatRoomLogsPage = parseInt($.cookie("chatRoomLogsPage"));
        RenderChatRoomLogsList(chatRoomLogsPage)
    });


    //清屏
    $(document).on('click','#cleanPnlMsg',function () {
        $(".ChatRoomDiv .pnl-msgs li").remove();
    });


    (function () {

        //https://github.com/li914/emoji_jQuery
        $.Lemoji({
            emojiInput: '#talkMsg',
            emojiBtn: '.emojiIcon',
            position: 'RIGHTTOP',
            length: 8,
            emojis: {
                qq: {path: '/game/images/qq/', code: ':', name: 'QQ表情'},
                tieba: {path: '/game/images/tieba', code: ';', name: "贴吧表情"},
                emoji: {path: '/game/images/emoji', code: ',', name: 'Emoji表情'}
            }
        });

    })();

});
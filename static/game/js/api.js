
/**
 * 聊天记录
 * @param page
 * @constructor
 */
let RenderChatRoomLogsList = function(page) {

    let chatRoomLogsPage = parseInt($.cookie("chatRoomLogsPage"));
    let cond = {};
    cond.page = page;

    adxRequest("/fight/room/chat-logs",function (data) {

        //点击更多按钮清除
        $(".ChatRoomDiv .MoreChatRoomLogs").remove();

        if (parseInt(data.iRet) !== 1) {
            layer.alert(data.sMsg);
            return ;
        }
        
        let aList  = data.data.list;

        if (aList.length > 0 ) {
            $.each(aList,function (kev,val) {

                let msg = RenderMsg(val.user_id,val.nick,val.avatar,val.time,val.content,val.cmd);

                $(".ChatRoomDiv .chatRoomLogsList ul").prepend(msg);

            });

            if ( chatRoomLogsPage < data.data.pages ) {
                chatRoomLogsPage ++ ;
                $.cookie("chatRoomLogsPage",chatRoomLogsPage);

                let moreEm = '<li class="layim-chat-system MoreChatRoomLogs" style="font-weight: bolder;color: #ff0000;display: none;">\n' +
                    '            <span style="cursor: pointer">加载更多</span>\n' +
                    '         </li>';

                $(".ChatRoomDiv .chatRoomLogsList ul").prepend(moreEm);

                $(".ChatRoomDiv .MoreChatRoomLogs").fadeIn(1000)
            }
        }

    },cond,"POST")
};
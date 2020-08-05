
function wLaoHuJi(id,gold) {
    this.frameid = id;
    this._doc = document;
    this._config = {
        cardwidth: 60,
        cardheight: 60,
        betcardwidth:97,
        betcardheight:97,
        margin:5,
        runboxlength:4,
        runloop:4,
        maxbet:999999
    };
    this._piecelist = [];
    this._multitype = {
        "b_bar":100,
        "s_bar":50,
        "b_seven": 40,
        "b_star":30,
        "b_watermelons":25,
        "b_alarm":20,
        "b_coconut":10,
        "b_orange":10,
        "b_apple":5,
        "small":2,
        "cha":0
    };
    this._piecelistposition = {};
    this._piecelistmulti = {};
    this._piecelisttype = {};

    this._money = 0; //得分
    this._total = parseInt(gold);           //默认分
    this._startbox = 1;         //上次结果，此次的起点
    this._endbox = 1;          //这是这次的结果
    this._jumpnum = 1;        //这些需要算出来
    this._currentshowlist = [1];

    //状态值
    this._isfirstbet = true;
    this._isrun = false;


    //定时器
    this._interval = null;

    this._mainDiv = null;
    this.frame = {
        "piece": {"bg":null,"run":null},
        "bet": null
    };
    this._self = this;
    var i;
    for (i = 1; i < 25; i++) {
        this._piecelist.push(i);
    }
}

wLaoHuJi.prototype.$ = function(id) {

    if (!this._doc) {
        this._doc = document;
    }

    if (id && typeof (id) === "string") {
        return this._doc.getElementById(id);
    }

    return null;
};


wLaoHuJi.prototype.isArray = function(obj) {
    return Object.prototype.toString.call(obj) === '[object Array]';
}

wLaoHuJi.prototype.in_Array = function(arr, e) {
    for (var i = 0; i < arr.length; i++) {
        if (arr[i] == e)
            return true;
    }
    return false;
}

//生成随机数
wLaoHuJi.prototype.rand = function(min, max) {
    return parseInt(Math.random() * (max - min + 1) + min);
}

wLaoHuJi.prototype._getWinNum = function() {

    var r = this.rand(1, 100);
    var b = 2;
    if (r == 100) {
        //大Bar 1/100
        b = 24;
    }
    else if (r > 97) {
        //小Bar 2/100
        b = 23;
    }
    else if (r >= 95 && r <= 97) {
        //大77 3/100
        b = 12;
    }
    else if (r >= 90 && r <= 94) {
        //大星星 5/100
        b = 16;
    }
    else if (r >= 84 && r <= 89) {
        //大西瓜  6/100
        b = 4;
    }
    else if (r >= 76 && r <= 83) {
        //大铃铛  8/100
        var ar = this.rand(1, 2);
        switch (ar) {
            case 1:
                b = 22;
                break;
            case 2:
                b = 10;
                break;
        }
    }
    else if (r >= 68 && r <= 75) {
        //大椰子1  8/100
        var ar = this.rand(1, 2);
        switch (ar) {
            case 1:
                b = 3;
                break;
            case 2:
                b = 15;
                break;
        }
    }
    else if (r >= 59 && r <= 68) {
        //大桔子1  10/100
        var ar = this.rand(1, 2);
        switch (ar) {
            case 1:
                b = 21;
                break;
            case 2:
                b = 9;
                break;
        }
    }
    else if (r >= 44 && r <= 58) {
        //苹果  15/100
        var ar = this.rand(1, 4);
        switch (ar) {
            case 1:
                b = 1;
                break;
            case 2:
                b = 7;
                break;
            case 3:
                b = 13;
                break;
            case 4:
                b = 19;
                break;
        }
    }
    else if (r >= 36 && r <= 43) {
        //CHA  8/100
        var ar = this.rand(1, 2);
        switch (ar) {
            case 1:
                b = 6;
                break;
            case 2:
                b = 18;
                break;
        }
    }
    else {
        //小东西
        var ar = this.rand(1, 7);
        switch (ar) {
            case 1:
                b = 2;
                break;
            case 2:
                b = 5;
                break;
            case 3:
                b = 8;
                break;
            case 4:
                b = 11;
                break;
            case 5:
                b = 14;
                break;
            case 6:
                b = 17;
                break;
            case 7:
                b = 20;
                break;
        }
    }
    return b;
};


wLaoHuJi.prototype._getpieceinfo = function(i, j) {
    switch (i + "-" + j) {
        case "0-0":
            return { "type":"orange","css": "b_orange", "list": 21, "multi": "b_orange" };
        case "0-1":
            return { "type": "alarm", "css": "b_alarm", "list": 22, "multi": "b_alarm" };
        case "0-2":
            return { "type": "bar", "css": "s_bar", "list": 23, "multi": "s_bar" };
        case "0-3":
            return { "type": "bar", "css": "b_bar", "list": 24, "multi": "b_bar" };
        case "0-4":
            return { "type": "apple", "css": "b_apple", "list": 1, "multi": "b_apple" };
        case "0-5":
            return { "type": "apple", "css": "s_apple", "list": 2, "multi": "small" };
        case "0-6":
            return { "type": "coconut", "css": "b_coconut", "list": 3, "multi": "b_coconut" };
        case "1-0":
            return { "type": "alarm", "css": "s_alarm", "list": 20, "multi": "small" };
        case "2-0":
            return { "type": "apple", "css": "b_apple", "list": 19, "multi": "b_apple" };
        case "3-0":
            return { "type": "cha", "css": "cha", "list": 18, "multi": "cha" };
        case "4-0":
            return { "type": "star", "css": "s_star", "list": 17, "multi": "small" };
        case "5-0":
            return { "type": "star", "css": "b_star", "list": 16, "multi": "b_star" };
        case "6-0":
            return { "type": "coconut", "css": "b_coconut", "list": 15, "multi": "b_coconut" };
        case "6-1":
            return { "type": "coconut", "css": "s_coconut", "list": 14, "multi": "small" };
        case "6-2":
            return { "type": "apple", "css": "b_apple", "list": 13, "multi": "b_apple" };
        case "6-3":
            return { "type": "seven", "css": "b_seven", "list": 12, "multi": "b_seven" };
        case "6-4":
            return { "type": "seven", "css": "s_seven", "list": 11, "multi": "small" };
        case "6-5":
            return { "type": "alarm", "css": "b_alarm", "list": 10, "multi": "b_alarm" };
        case "6-6":
            return { "type": "orange", "css": "b_orange", "list": 9, "multi": "b_orange" };
        case "5-6":
            return { "type": "orange", "css": "s_orange", "list": 8, "multi": "small" };
        case "4-6":
            return { "type": "apple", "css": "b_apple", "list": 7, "multi": "b_apple" };
        case "3-6":
            return { "type": "cha", "css": "cha", "list": 6, "multi": "cha" };
        case "2-6":
            return { "type": "watermelons", "css": "s_watermelons", "list": 5, "multi": "small" };
        case "1-6":
            return { "type": "watermelons", "css": "b_watermelons", "list": 4, "multi": "b_watermelons" };
        default:
            return { "type": "", "css": "", "list": 0, "multi": "" };
    }
};

//显示单个或多个灯
wLaoHuJi.prototype.showbox = function(index) {
    var i, len, tpleft, tptop, box = '';
    if (typeof (index) === 'number' && index > 0 && index < 25) {
        tpleft = this._piecelistposition[index].left;
        tptop = this._piecelistposition[index].top;
        box = "<div style='position:absolute;width:" + this._config.cardwidth + "px;height:" + this._config.cardheight + "px;left:" + tpleft + "px;top:" + tptop + "px;' class='win_piece'></div>";
        this.frame.piece.run.innerHTML = box;
    } else if (this.isArray(index) && index.length > 0) {
        len = index.length;
        for (i = 0; i < len; i++) {
            tpleft = this._piecelistposition[index[i]].left;
            tptop = this._piecelistposition[index[i]].top;
            box += "<div style='position:absolute;width:" + this._config.cardwidth + "px;height:" + this._config.cardheight + "px;left:" + tpleft + "px;top:" + tptop + "px;' class='win_piece'></div>";
        }
        this.frame.piece.run.innerHTML = box;
    }
}

//每次改变需要显示的box，返回速度
wLaoHuJi.prototype.changeshowlist = function(jumpindex) {
    var i,
        len = this._currentshowlist.length,
        jumpmax = this._jumpnum;
    switch (jumpindex) {
        case 0:
            var v = this._currentshowlist[0] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 400;
        case 1:
            if (len == 1) {
                var v = this._currentshowlist[0] + 1;
                v = v > 24 ? v - 24 : v;
                this._currentshowlist.push(v);
            }
            return 350;
        case 2:
            if (len == 2) {
                var v = this._currentshowlist[1] + 1;
                v = v > 24 ? v - 24 : v;
                this._currentshowlist.push(v);
            }
            return 300;
        case 3:
            if (len == 3) {
                var v = this._currentshowlist[2] + 1;
                v = v > 24 ? v - 24 : v;
                this._currentshowlist.push(v);
            }
            return 200;
        case jumpmax - 1:
            var v = this._currentshowlist[0] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 800;
        case jumpmax - 2:
            var v = this._currentshowlist[0] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 700;
        case jumpmax - 3:
            var v = this._currentshowlist[0] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 600;
        case jumpmax - 4:
            var v = this._currentshowlist[0] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 400;
        case jumpmax - 5:
            var v = this._currentshowlist[0] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 300;
        case jumpmax - 6:
            var v = this._currentshowlist[1] + 1;
            this._currentshowlist.length = 0;
            v = v > 24 ? v - 24 : v;
            this._currentshowlist[0] = v;
            return 200;
        case jumpmax - 7:
            var v1 = this._currentshowlist[1] + 1;
            var v2 = this._currentshowlist[2] + 1;
            this._currentshowlist.length = 0;
            v1 = v1 > 24 ? v1 - 24 : v1;
            v2 = v2 > 24 ? v2 - 24 : v2;
            this._currentshowlist[0] = v1;
            this._currentshowlist[1] = v2;
            return 100;
        case jumpmax - 8:
            var v1 = this._currentshowlist[1] + 1;
            var v2 = this._currentshowlist[2] + 1;
            var v3 = this._currentshowlist[3] + 1;
            this._currentshowlist.length = 0;
            v1 = v1 > 24 ? v1 - 24 : v1;
            v2 = v2 > 24 ? v2 - 24 : v2;
            v3 = v3 > 24 ? v3 - 24 : v3;
            this._currentshowlist[0] = v1;
            this._currentshowlist[1] = v2;
            this._currentshowlist[2] = v3;
            return 50;
        default:
            for (i = 0; i < len; i++) {
                this._currentshowlist[i]++;
                if (this._currentshowlist[i] > 24) {
                    this._currentshowlist[i] -= 24;
                }
            }
            return 30;
    }
}

wLaoHuJi.prototype.run = function() {
    var self = this._self,
        time = 500,
        inx = 0,
        shownum = 1,
        boxmax = this._config.runboxlength,
        runloopnum = 0,
        jumpmax = this._jumpnum,
        jumpindex = 0,
        timer = null;
    function timerdo() {
        time = self.changeshowlist(jumpindex);
        self.showbox(self._currentshowlist);
        jumpindex++;
        if (jumpindex >= jumpmax) {
            clearTimeout(timer);
            self._startbox = self._endbox;
            setTimeout(function() { self.result(); }, 200);
        }
        else {
            timer = setTimeout(timerdo, time);
        }
    }
    timerdo();
}

wLaoHuJi.prototype.result = function() {

    //统一由服务器下发结果
    /*var winbox = this._endbox;
    var type = this._piecelisttype[winbox];
    if (type == "cha" || type == "") {
        //未得奖，或别的奖，暂无
    } else {
        var betid = "lhj_bet_txt_" + type;
        var betnum = parseInt(this.$(betid).value);
        if (betnum > 0) {
            this._money = parseInt(this._piecelistmulti[winbox]) * betnum;
            this.$('lhj_ben_txt_money').innerHTML = this._money;
        }
    }*/
    this._isfirstbet = true;
    this._isrun = false;
}

wLaoHuJi.prototype.init = function() {
    var i, j, piecewidth, pieceheight, betwidth, betheight, tnhtml, piecehtml, bethtml, advwidth, advheight, advleft, advtop, bsheight, self = this._self;
    this._mainDiv = this.$(this.frameid);
    if (!this._mainDiv) {
        layer.alert("不存在,别瞎搞");
        return;
    }
    piecewidth = (this._config.cardwidth + this._config.margin) * 7 + this._config.margin + 100;
    pieceheight = (this._config.cardheight + this._config.margin) * 7 + this._config.margin + 2;
    betwidth = (this._config.betcardwidth + this._config.margin) * 4 + 64;
    betheight = pieceheight;
    this._mainDiv.style.border = "1px solid green";
    this._mainDiv.style.width = (piecewidth + betwidth + 5) + "px";
    this._mainDiv.style.height = pieceheight + "px";
    this._mainDiv.style.margin = "10px auto";
    this._mainDiv.style.padding = "0";
    let mainhtml = "<table cellpadding='0' cellspacing='0' style='-moz-user-select:none;margin:0;padding:0;border:0'><tr><td><div style = 'position: relative;top:0px;left:0px;width:" + piecewidth + "px;height:" + pieceheight + "px;'><div id='lhj_piece_bg' style='position:absolute;top:0,left:0;z-index:101;'></div><div id='lhj_piece_run' style='position:absolute;top:0,left:0;z-index:102;'></div></div></td><td valign='top'><div id='lhj_bet'></div></td></tr></table>";
    this._mainDiv.innerHTML = mainhtml;
    this.frame.piece.bg = this.$('lhj_piece_bg'); //this._mainDiv.childNodes[0].childNodes[0].childNodes[0];
    this.frame.piece.run = this.$('lhj_piece_run');
    this.frame.bet = this.$('lhj_bet'); //this._mainDiv.childNodes[0].childNodes[0].childNodes[1];
    //初始化piece
    piecehtml = [];
    for (i = 0; i < 7; i++) {
        for (j = 0; j < 7; j++) {
            if (i == 0 || j == 0 || i == 6 || j == 6) {
                var tpleft = j * (this._config.cardwidth + this._config.margin) + this._config.margin;
                var tptop = i * (this._config.cardheight + this._config.margin) + this._config.margin;
                var tpinfo = this._getpieceinfo(i, j);
                piecehtml.push("<div id='" + i + "-" + j + "' style='position:absolute;width:" + this._config.cardwidth + "px;height:" + this._config.cardheight + "px;left:" + tpleft + "px;top:" + tptop + "px;' class='piece " + tpinfo.css + "'></div>");
                this._piecelistposition[tpinfo.list] = { left: tpleft, top: tptop };
                this._piecelistmulti[tpinfo.list] = this._multitype[tpinfo.multi];
                this._piecelisttype[tpinfo.list] = tpinfo.type;
            }
        }
    }
    advwidth = (this._config.cardwidth + this._config.margin) * 5 - this._config.margin;
    advheight = (this._config.cardheight + this._config.margin) * 5 - this._config.margin;
    advleft = this._config.cardwidth + this._config.margin * 2;
    advtop = this._config.cardheight + this._config.margin * 2;
    piecehtml.push("<div id='adv' style='position:absolute;border:1px red solid;width:" + advwidth + "px;height:" + (advheight-10) + "px;z-index:102;left:" + advleft + "px;top:" + advtop + "px;text-align: center;padding-top: 10px;color: #ff0000;font-weight: bolder;font-size: 16px'></div>");
    this.frame.piece.bg.innerHTML = piecehtml.join('');

    //初始化bet
    bethtml = [];
    bsheight = betheight - this._config.betcardheight * 2 - 118 - 112;
    bethtml.push("<div style='width:" + (betwidth - 8) + "px;height:63px;margin:8px 3px;border:0px dotted red;'>");
    bethtml.push("<span class='lhj_span'>赢得金币<br /><span class='lhj_input lhj_money_input' id='lhj_ben_txt_money' />0</span></span>");
    bethtml.push("<span class='lhj_span'>总金币<br /><span class='lhj_input lhj_money_input' id='lhj_ben_txt_total' />0</span></span>");
    bethtml.push("</div>");
    bethtml.push("<div style='width:" + (betwidth - 8) + "px;height:" + bsheight + "px;margin:2px 5px 3px 3px;border:1px dotted #ffda76;'>");
    bethtml.push("<div id='GameMsg' style='color: red;line-height: 70px;margin: 5px 20px ;display: none;word-wrap:break-word; word-break:break-all;'>恭喜 [专业偷充电桩] 获得本局最大收益:10000金币</div>");
    bethtml.push("</div>");
    bethtml.push("<div style='width:" + (betwidth - 8) + "px;'>");
    bethtml.push("<table cellpadding='0' cellspacing='0' class='lhj_bet_table'>");
    bethtml.push("<tr>");
    bethtml.push("<td><div class='bet bar' id='lhj_bet_bar' title='奖励值：大100小50' ></div></td>");
    bethtml.push("<td><div class='bet seven' id='lhj_bet_seven' title='奖励值：大40小2'></div></td>");
    bethtml.push("<td><div class='bet star' id='lhj_bet_star' title='奖励值：大30小2'></div></td>");
    bethtml.push("<td><div class='bet watermelons' id='lhj_bet_watermelons' title='奖励值：大25小2'></div></td>");
    bethtml.push("</tr>");
    bethtml.push("<tr>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_bar' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_bar' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_seven' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_seven' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_star' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_star' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_watermelons' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_watermelons' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("</tr>");
    bethtml.push("<tr>");
    bethtml.push("<td><div class='bet alarm'  id='lhj_bet_alarm' title='奖励值：大20小2'></div></td>");
    bethtml.push("<td><div class='bet coconut'  id='lhj_bet_coconut' title='奖励值：大10小2'></div></td>");
    bethtml.push("<td><div class='bet orange' id='lhj_bet_orange' title='奖励值：大10小2'></div></td>");
    bethtml.push("<td><div class='bet apple' id='lhj_bet_apple' title='奖励值：大5小2'></div></td>");
    bethtml.push("</tr>");
    bethtml.push("<tr>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_alarm' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_alarm' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_coconut' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_coconut' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_orange' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_orange' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("<td>共投：<input type='text' id='total_lhj_bet_txt_apple' class='lhj_input lhj_bet_input' value='0' readOnly />已投：<input type='text' id='lhj_bet_txt_apple' class='lhj_input lhj_bet_input' value='0' readOnly /></td>");
    bethtml.push("</tr>");
    bethtml.push("</table>");
    bethtml.push("</div>");
    bethtml.push("<div style='width:" + betwidth + "px;text-align:right;'>");
    bethtml.push("<span id='bankrupt' style='color: #FF0000;cursor: pointer'>破产补助</span>&nbsp;&nbsp;");
    bethtml.push("押注倍数：<select id='betBei'><option value='1'>1倍</option><option value='5'>5倍</option><option value='10'>10倍</option><option value='100'>100倍</option></select>");
    bethtml.push("</div>");
    this.frame.bet.innerHTML = bethtml.join('');

    //清空下注面板
    wLaoHuJi.prototype.ResetBetInfoList = function () {

        self.$('total_lhj_bet_txt_bar').value = 0;
        self.$('total_lhj_bet_txt_seven').value = 0;
        self.$('total_lhj_bet_txt_star').value = 0;
        self.$('total_lhj_bet_txt_watermelons').value = 0;
        self.$('total_lhj_bet_txt_alarm').value = 0;
        self.$('total_lhj_bet_txt_coconut').value = 0;
        self.$('total_lhj_bet_txt_orange').value = 0;
        self.$('total_lhj_bet_txt_apple').value = 0;

        self.$('lhj_bet_txt_bar').value = 0;
        self.$('lhj_bet_txt_seven').value = 0;
        self.$('lhj_bet_txt_star').value = 0;
        self.$('lhj_bet_txt_watermelons').value = 0;
        self.$('lhj_bet_txt_alarm').value = 0;
        self.$('lhj_bet_txt_coconut').value = 0;
        self.$('lhj_bet_txt_orange').value = 0;
        self.$('lhj_bet_txt_apple').value = 0;
    };


    //初始化下注数据
    wLaoHuJi.prototype.InitBetInfoList = function (data) {

        //全体用户下注数据
        $.each(data.total_bet,function (key,val) {
            self.$("total_"+key).value = val;
        });

        //自己下注数据
        $.each(data.my_bet,function (key,val) {
            self.$(key).value = val
        });

    };

    //设置金币
    wLaoHuJi.prototype.ResetGold = function(gold) {

        self._money = 0;

        self.$('lhj_ben_txt_money').innerHTML =  0;

        self._total = parseInt(gold);

        self.$('lhj_ben_txt_total').innerHTML =  self._total;
    };

    let betmoeny = function(id) {
        if (self._isrun) {
            return;
        }

        //先将得分转过来
        self._total += self._money;
        self._money = 0;
        self.$('lhj_ben_txt_total').innerHTML = self._total;
        self.$('lhj_ben_txt_money').innerHTML = self._money;

        if (self._isfirstbet) {
            wLaoHuJi.prototype.ResetBetInfoList();
            self._isfirstbet = false;
        }

        var tpv = parseInt(self.$(id).value);

        let bei = parseInt(self.$('betBei').value);

        if (self._total - bei >= 0 && tpv < self._config.maxbet) {

            //ws 下注消息推送
            var obj = {id:id,add:bei};
            SendMsg(100001,JSON.stringify(obj))

            /*self._total -= bei;
            self.$(id).value = tpv + bei;
            self.$('lhj_ben_txt_total').innerHTML = self._total;*/
        } else {
            layer.alert("金币不足!")
        }
    }

    this.$('lhj_bet_bar').onclick = function() {
        betmoeny('lhj_bet_txt_bar');
    }
    this.$('lhj_bet_seven').onclick = function() {
        betmoeny('lhj_bet_txt_seven');
    }
    this.$('lhj_bet_star').onclick = function() {
        betmoeny('lhj_bet_txt_star');
    }
    this.$('lhj_bet_watermelons').onclick = function() {
        betmoeny('lhj_bet_txt_watermelons');
    }
    this.$('lhj_bet_alarm').onclick = function() {
        betmoeny('lhj_bet_txt_alarm');
    }
    this.$('lhj_bet_coconut').onclick = function() {
        betmoeny('lhj_bet_txt_coconut');
    }
    this.$('lhj_bet_orange').onclick = function() {
        betmoeny('lhj_bet_txt_orange');
    }
    this.$('lhj_bet_apple').onclick = function() {
        betmoeny('lhj_bet_txt_apple');
    }

    wLaoHuJi.prototype.GameRunning = function (endbox) {
        var betcount = parseInt(self.$('lhj_bet_txt_bar').value);
        betcount += parseInt(self.$('lhj_bet_txt_seven').value);
        betcount += parseInt(self.$('lhj_bet_txt_star').value);
        betcount += parseInt(self.$('lhj_bet_txt_watermelons').value);
        betcount += parseInt(self.$('lhj_bet_txt_alarm').value);
        betcount += parseInt(self.$('lhj_bet_txt_coconut').value);
        betcount += parseInt(self.$('lhj_bet_txt_orange').value);
        betcount += parseInt(self.$('lhj_bet_txt_apple').value);

        //先将得分转过来
        self._total += self._money;
        self._money = 0;
        self.$('lhj_ben_txt_total').innerHTML = self._total;
        self.$('lhj_ben_txt_money').innerHTML = self._money;
        //随机得到中奖结果

        //判断分还够不够下注
        if (self._isfirstbet) {
            if (self._total < betcount) {
                layer.alert("金币不足！");
                return;
            }
            self._total -= betcount;
            self.$('lhj_ben_txt_total').innerHTML = self._total;
        }

        this.disabled = true;
        self._isrun = true;
        //self._endbox = self._getWinNum();
        self._endbox = endbox ;
        //算出运行步数
        var step = (self._endbox - self._startbox) > 0 ? self._endbox - self._startbox : 24 - self._startbox + self._endbox;
        self._jumpnum = 24 * 4 + step; //这些需要算出来
        self.run();
    };

    //初始化默认总分
    this.$('lhj_ben_txt_total').innerHTML = this._total;


    wLaoHuJi.prototype.SetGameMsg = function(msg) {
        $("#GameMsg").fadeOut(100,function () {
            $("#GameMsg").text(msg);
            $("#GameMsg").fadeIn();
        });
    };
};

//初始化游戏
let InitwLaoHuJiGame = function(gold) {

    if (Object.keys(LhjGameInstance).length == 0) {
        LhjGameInstance = new wLaoHuJi("game",gold);
        LhjGameInstance.init();
    } else {
        LhjGameInstance.ResetGold(gold);
    }

    return LhjGameInstance
};

//新开局游戏
let LaoHuJiGameRunning = function (cbox,endbox,gold) {

    LhjGameInstance.ResetGold(gold);

    cbox                             = parseInt(cbox);
    endbox                           = parseInt(endbox);
    LhjGameInstance._startbox        = cbox;
    LhjGameInstance._currentshowlist = [cbox];

    LhjGameInstance.GameRunning(endbox)
};
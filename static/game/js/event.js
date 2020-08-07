
$(document).ready(function () {


    let RandAvatar   = 0 ;

    //绑定form表单 https://www.layui.com/doc/base/faq.html#form
    layui.use('form', function(){
        let form = layui.form;
        form.render();
    });

    //设置右上角头像
    let setRightTopAvatar = function(index) {
        let avatar = GetUserAvatar(index) ;
        $('#UserAvatar img').attr("src",avatar);
    };

    let InitUserInfo = function(data) {

        let ret = data.data;

        $.cookie("user_id",ret.user_id,{ expires: 30, path: '/' });
        $.cookie("sessid",ret.sessid,{ expires: 30, path: '/' });

        setRightTopAvatar(ret.avatar);

        let avatar = GetUserAvatar(ret.avatar);

        $('.userInfo').show();
        $('#LogOut').show();

        $('#UserLogin').hide();
        $('#UserReg').hide();
        $('#UserAvatar .loginTips').hide();

        $(".UserInfoForm .user_id").text(ret.user_id);
        $(".UserInfoForm .nick").val(ret.nick);
        $(".UserInfoForm .sex").val(ret.sex);
        $(".UserInfoForm .age").val(ret.age);
        $(".UserInfoForm .avatar").attr("src",avatar);
        $(".UserInfoForm .reg_time").text(ret.reg_time);
        $(".UserInfoForm .MyGolds").text(ret.gold);
        $(".UserInfoForm .MyLevel").text(ret.level);
        //使用canvas生成
        $('.UserInfoForm .qrcode').empty().qrcode({
            render: "canvas",
            width: 100,
            height: 100,
            text: ret.qrcode
        });

        LhjGameInstance.ResetGold(ret.gold);

        //socket
        InitWsConn();
    };

    let GoToReg = function() {
        adxRequest("/fight/login/reg",function (data) {
            if (parseInt(data.iRet) !== 1) {
                layer.alert(data.sMsg);
                return ;
            }

            InitUserInfo(data);

            if(parseInt(data.data.isRegister) === 1) {
                setTimeout(function () {
                    $('.userInfo').click();

                    //自动下载二维码
                    $('.saveLoginQrCode').click();

                    let html = '<span style="color: #ff0000;font-weight: bolder"> "登陆二维码" 是登陆唯一凭证！首次登陆时系统已自动下载，请妥善保管！</span>';

                    layer.alert(html)
                },500);
            }

        },null,"POST")
    };

    //每日土豪榜单
    $(document).on("click",".TuHaoList",function () {
        adxRequest("/fight/room/tuhao",function (data) {
            if (parseInt(data.iRet) != 1) {
                layer.alert(data.sMsg);
                return ;
            }

            let aList = data.data.list;
            if (aList.length > 0) {
                $(".TuHaoListMap dl").empty();

                $.each(aList,function (rank,aUser) {
                    let avatar = GetUserAvatar(aUser.Avatar);
                    let html   = '<dd>';
                    html       +=   '<a>';
                    html       +=       '<img src="' + avatar + '" title="'+aUser.Nick+'">';
                    html       +=       '<cite>' + aUser.Nick + '</cite>';
                    html       +=       '<i title="金币：'+aUser.Gold+'">'+aUser.Level+' '+aUser.Gold+'</i>';
                    html       +=   '</a>';
                    html       += '</dd>';

                    $(".TuHaoListMap dl").append(html);
                });
            }


        },null,"POST")
    });

    //破产补助 bankrupt
    $(document).on("click","#bankrupt",function () {
        adxRequest("/fight/room/bankrupt",function (data) {
            if (parseInt(data.iRet) != 1) {
                layer.alert(data.sMsg);
                return ;
            }

            LhjGameInstance.ResetGold(data.data.gold);
            layer.alert(data.sMsg)

        },null,"POST")
    });

    //注册
    $("#UserReg").on("click",function () {

        let user_id = parseInt($.cookie("user_id"));

        if (user_id > 0) {
            let index  = layer.confirm('您已注册过账号，确定重新注册吗？', {
                btn: ['确定','放弃'] //按钮
            }, function(){

                layer.close(index);

                GoToReg();

                return false

            }, function(){

                layer.close(index);

                return false;
            });
        } else {
            GoToReg();
        }
    });

    //登陆
    $("#UserLogin").on("click",function () {

        let html = '<input type="file" id="uploadFile" >';

        let index = layer.confirm( html, {
            title:"请上传服务器生成的二维码",
            btn: ['登录','一键生成新号'] //按钮
        }, function(){

            let formData = new FormData();

            formData.append("qrcode",$('#uploadFile')[0].files[0]);

            $.ajax({
                url:"/fight/login/index",
                type:"post",
                data:formData,
                processData: false, // 告诉jQuery不要去处理发送的数据
                contentType: false, // 告诉jQuery不要去设置Content-Type请求头
                success:function(data){

                    if (parseInt(data.iRet) !== 1) {
                        layer.alert(data.sMsg);
                        return ;
                    }

                    $.cookie("user_id",data.data.user_id,{ expires: 30, path: '/' });
                    $.cookie("sessid",data.data.sessid,{ expires: 30, path: '/' });

                    window.location.href = "/"
                },
                dataType:"json"
            })

        }, function(){
            layer.close(index);
            $("#UserReg").click();
        });
    });

    //获取随机头像
    $(".UserInfoForm .avatar").on("click",function () {
        adxRequest("/fight/home/avatar",function (data) {
            if (parseInt(data.iRet) != 1) {
                layer.alert(data.sMsg);
                return ;
            }

            RandAvatar = data.data.avatar;

            let avatar = GetUserAvatar(RandAvatar)  + "?_=" + Math.random() ;

            $(".UserInfoForm .avatar").attr("src",avatar);

        },null,"POST")
    });

    //修改用户信息
    $(".modifyUserInfo").on("click",function () {

        let cond = {};

        cond.nick   = $(".UserInfoForm .nick").val().trim();
        cond.avatar = RandAvatar;
        cond.sex    = $(".UserInfoForm .sex option:selected").val();
        cond.age    = $(".UserInfoForm .age").val();

        adxRequest("/fight/home/modify",function (data) {
            if (parseInt(data.iRet) != 1) {
                layer.alert(data.sMsg);
                return ;
            }

            setRightTopAvatar(RandAvatar);

            layer.alert(data.sMsg)

        },cond,"POST")
    });

    //保存登陆用二维码
    $(".saveLoginQrCode").on("click",function () {

        //将Jquery对象转化为JavaScript对象

        let canvas = $(".UserInfoForm .qrcode canvas").get(0);

        //将画布转化为base64格式

        let url = canvas.toDataURL('image/png');

        //JavaScript初始化下载方法

        let a = document.createElement('a');

        document.body.appendChild(a);

        a.href = url;

        let nick   = $(".UserInfoForm .nick").val().trim();
        a.download = "【" + nick + '】登陆二维码.png';

        //执行下载
        a.click();
    });


    //退出
    $("#LogOut").on("click",function () {
        adxRequest("/fight/login/out",function (data) {
            if (parseInt(data.iRet) !== 1) {
                layer.alert(data.sMsg);
                return ;
            }

            $.cookie("logout",1);

            window.location.reload()

        },null,"POST")
    });

    //登陆校验
    let CheckLogin = function (jump) {

        let aCookie  = $.cookie("logout");
        let isLogout = parseInt(aCookie);

        $.cookie("logout",0);

        if(isLogout === 0 || typeof (aCookie) === "undefined") {
            adxRequest("/fight/login/auto",function (data) {

                let iRet = parseInt(data.iRet);
                if (iRet === 1) {
                    //登录初始化
                    InitUserInfo(data);

                    //初始化面板押注数据
                    if(jump === true) {
                        $('.GameGuess').click()
                    }

                } else {

                    $("#UserLogin").click();
                }

            },null,"POST")
        }
    };

    //点击个人信息面板
    $(document).on('click','.userInfo',function () {
        CheckLogin(false)
    });

    $(".UploadImgIcon").click(function(){
        $("#upload").trigger('click');
    });

    $("#upload").change(function(){
        //formdata对象，用来模拟表单
        let formData = new FormData($('#uploadform')[0]);
        $("#uploadform")[0].reset();

        $.ajax({
            url:"/fight/home/upload",
            type:"post",
            data:formData,
            processData: false, // 告诉jQuery不要去处理发送的数据
            contentType: false, // 告诉jQuery不要去设置Content-Type请求头
            success:function(data){
                if (parseInt(data.iRet) !== 1) {
                    layer.alert(data.sMsg);
                    return ;
                }
            },
            dataType:"json"
        })
    });

    //播放音频
    $(document).on('click','.playAudio',function () {

        let othis = $(this);

        var audioData = othis.data('audio')
            ,audio = audioData || document.createElement('audio')
            ,pause = function(){
            audio.pause();
            othis.removeAttr('status');
            othis.find('i').html('&#xe652;');
        };
        if(othis.data('error')){
            return layer.msg('播放音频源异常');
        }
        if(!audio.play){
            return layer.msg('您的浏览器不支持audio');
        }
        if(othis.attr('status')){
            pause();
        } else {
            audioData || (audio.src = othis.data('src'));
            audio.play();
            othis.attr('status', 'pause');
            othis.data('audio', audio);
            othis.find('i').html('&#xe651;');
            //播放结束
            audio.onended = function(){
                pause();
            };
            //播放异常
            audio.onerror = function(){
                layer.msg('播放音频源异常');
                othis.data('error', true);
                pause();
            };
        }

    });

    $(".UploadAudioIcon").click(function () {
        layer.prompt({title: '请输入网络音频地址', formType: 3}, function(pass, index){
            layer.close(index);
            var cond = {};

            cond.link = pass;

            adxRequest("/fight/home/audio",function (data) {
                if (parseInt(data.iRet) !== 1) {
                    layer.alert(data.sMsg);
                    return ;
                }
            },cond,"POST")
        });
    });

    $(document).on('click','.playVideo',function () {

        let videoData = $(this).data('src')
            ,video = document.createElement('video');
        if(!video.play){
            return layer.msg('您的浏览器不支持video');
        }

        layer.open({
            type: 1
            ,title: '播放视频-网络视频可能有跨域限制而无法播放'
            ,area: ['460px', '300px']
            ,maxmin: true
            ,shade: false
            ,content: '<div style="background-color: #000; height: 100%;"><video style="position: absolute; width: 100%; height: 100%;" src="'+ videoData +'" loop="loop" autoplay="autoplay"></video></div>'
        });

    });

    $(".UploadVideoIcon").click(function () {
        layer.prompt({title: '请输入网络视频地址', formType: 3}, function(pass, index){
            layer.close(index);
            var cond = {};

            cond.link = pass;

            adxRequest("/fight/home/video",function (data) {
                if (parseInt(data.iRet) !== 1) {
                    layer.alert(data.sMsg);
                    return ;
                }
            },cond,"POST")
        });
    });


    (function () {

        let  mobile_flag = isMobile(); // true为PC端，false为手机端
        if(mobile_flag){
            layer.alert("暂不支持移动端，请在PC端体验",function () {
                window.location.href = "/"
            })
        } else {
            InitwLaoHuJiGame(0);

            CheckLogin(true);
        }
    })();
});
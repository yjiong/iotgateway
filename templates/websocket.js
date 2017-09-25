(function (window, undefined) {
    $(function () {
        var socket, $win = $('body');
        function format(txt,compress/*是否为压缩模式*/){/* 格式化JSON源码(对象转换为JSON文本) */  
                var indentChar = '    ';   
                if(/^\s*$/.test(txt)){   
                    alert('数据为空,无法格式化! ');   
                    return;   
                }   
                try{var data=eval('('+txt+')');}   
                catch(e){   
                    alert('数据源语法错误,格式化失败! 错误信息: '+e.description,'err');   
                    return;   
                };   
                var draw=[],last=false,This=this,line=compress?'':'\n',nodeCount=0,maxDepth=0;   
                   
                var notify=function(name,value,isLast,indent/*缩进*/,formObj){   
                    nodeCount++;/*节点计数*/  
                    for (var i=0,tab='';i<indent;i++ )tab+=indentChar;/* 缩进HTML */  
                    tab=compress?'':tab;/*压缩模式忽略缩进*/  
                    maxDepth=++indent;/*缩进递增并记录*/  
                    if(value&&value.constructor==Array){/*处理数组*/  
                        draw.push(tab+(formObj?('"'+name+'":'):'')+'['+line);/*缩进'[' 然后换行*/  
                        for (var i=0;i<value.length;i++)   
                            notify(i,value[i],i==value.length-1,indent,false);   
                        draw.push(tab+']'+(isLast?line:(','+line)));/*缩进']'换行,若非尾元素则添加逗号*/  
                    }else   if(value&&typeof value=='object'){/*处理对象*/  
                            draw.push(tab+(formObj?('"'+name+'":'):'')+'{'+line);/*缩进'{' 然后换行*/  
                            var len=0,i=0;   
                            for(var key in value)len++;   
                            for(var key in value)notify(key,value[key],++i==len,indent,true);   
                            draw.push(tab+'}'+(isLast?line:(','+line)));/*缩进'}'换行,若非尾元素则添加逗号*/  
                        }else{   
                                if(typeof value=='string')value='"'+value+'"';   
                                draw.push(tab+(formObj?('"'+name+'":'):'')+value+(isLast?'':',')+line);   
                        };   
                };   
                var isLast=true,indent=0;   
                notify('',data,isLast,indent,false);   
                return draw.join('');   
            }  
        showmessage = function (msg, type) {
            var datetime = new Date();
            var tiemstr = datetime.getHours() + ':' + datetime.getMinutes() + ':' + datetime.getSeconds() + '.' + datetime.getMilliseconds();
            if (type) {
                var $p = $('<div>').appendTo($win.find('#div_msg'));
                var $type = $('<span>').text('[' + tiemstr + ']' + type + '：').appendTo($p);
                var $msg = $('<span>').addClass('thumbnail').css({ 'margin-bottom': '5px' }).text(msg).appendTo($p);
            } else {
                var $center = $('<center>').text(msg + '(' + tiemstr + ')').css({ 'font-size': '12px' }).appendTo($win.find('#div_msg'));
            }
        };

        $win.find('#refresh_clearcache').click(function () {
            $.yszrefresh();
        });
        
        var url = "ws" + document.location.href.substring(4) + "message";
        socket = new WebSocket(url);
        socket.onmessage = function (eve) {
	    showmessage('onmessage');
//		    var jsobj = JSON.paser(eve.data);
//		    var st = JSON.stringify(jsobj,null,4);
            showmessage(eve.data, 'receive');
        };
        
        socket.onopen = function (event) {
            showmessage('连接成功');
        };
        socket.onclose = function (event) {
            showmessage('断开连接');
        };
/**************************************
        $win.find('#btn_conn').attr('disabled', false);
       $win.find('#btn_close').attr('disabled', true);

        $win.find('#btn_conn').click(function () {
            $win.find('#btn_conn').attr('disabled', true);
            $win.find('#btn_close').attr('disabled', false);
            //var url = $win.find('#inp_url').val();
            var url = "ws://127.0.0.1:8000/message";
            // 创建一个Socket实例
            socket = new WebSocket(url);
            showmessage('开始连接');
            // 打开Socket 
            socket.onopen = function (event) {
                // 发送一个初始化消息
                showmessage('连接成功');
            };
            // 监听消息
            socket.onmessage = function (eve) {
                showmessage(eve.data, 'receive');
            };
            // 监听Socket的关闭
            socket.onclose = function (event) {
                showmessage('断开连接');
                $win.find('#btn_conn').attr('disabled', false);
                $win.find('#btn_close').attr('disabled', true);
            };
        });*********************************/
        $win.find('#btn_close').click(function () {
            if (socket) {
                socket.close();
            }
        });
        $win.find('#btn_send').click(function () {
            var msg = $win.find('#inp_send').val();
            if (socket && msg) {
                socket.send(msg);
                showmessage(msg, 'send');
                $win.find('#inp_send').val('');
            }
        });
        $win.find('#inp_send').keyup(function () {
            if (event.ctrlKey && event.keyCode == 13) {
                $win.find('#btn_send').trigger('click');
            }
        });

        $win.find('#btn_clear').click(function () {
            $win.find('#div_msg').empty();
        }); 
    });
})(window);

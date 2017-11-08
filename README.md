
### iot-gateway与iot server 之间的报文约定:
1. topic
    * server publish到gateway的topic为:gateway_name/server_name.比如gateway的id是iot-20170001,server是iot-server,那么topic为:iot-20170001/iot-server
    * server subscribe自gateway的topic为:server_name/gateway_name,即为上述的:iot-server/iot-20170001, 对于同一个服务器来说,iot-server只有一个,那么订阅topic可以是:iot-server/#,这样就可以接收到所有gateway推送的消息.
2. 消息类型
    * request消息:无论是server和gateway,主动发起的都为request类型的消息.
    * response消息:gateway响应server发来的request的消息.注:server无须响应gateway的request,比如自动循环上报给server的request消息.
## 消息举例
### 1. 帮助命令 help:
```
        发送的消息命令如下:
        {
            "request": {
                "cmd": "help",
            	"data":"ModbusTcp"
        	}
        }
        解释:
        data字段是具体的设备.如果没有data字段,返回的应答是gataway的帮助信息,如果有具体的data设备,  
            将是具体的设备的帮助信息.
```
### 2. server获取gateway配置信息命令 init/get.do:
```
            发送的消息命令如下:
            {
                "msgtype": "request",
                "request": {
                    "cmd": "init/get.do",
            	    "requestid":"1234567890-0987",
                    "return": [
                        "requestid",
                        "cmd"
                    ],
                    "timestamp": 1470901793
                }
            }
            必要key:request,cmd
            其他key可有可无.
            解释:
                requestid:是server生成的唯一的guid,与return同时使用,server用来判断gateway应答的消息是  
                    该请求的应答消息.
                return:此段内包含的key,就是需要gateway应答时携带返回的,一般用做判断用,按需要使用.
                timestamp:时间戳.
                
             接收到的应答一般如下:
            {
              "header" : {
                "from" : {
                  "_devid" : "iot_gate",
                  "_model" : "TG120",
                  "_runstate" : "127",
                  "_version" : "1.0"
                },
                "msgtype" : "response"
              },
              "request" : {
                "cmd" : "init/get.do",
                "requestid" : "1234567890-0987"
              },
              "response" : {
                "cmd" : "init/get.do",
                "data" : {
                  "_client_id" : "iot_gate",
                  "_client_ip" : "192.168.64.1",
                  "_interval" : "10",
                  "_keepalive" : "60",
                  "_password" : "passwd",
                  "_server_ip" : "127.0.0.1",
                  "_server_name" : "server_name"
                  "_server_port" : "1883",
                  "_username" : "username",
                  "_will" : "1"
                },
                "statuscode" : 0,
                "timestamp" : 1507606025
              }
            }
            解释:
            request:就是上面return要求返回的东西,cmd和requestid,如果命令没return,这个字段就没有.
            response:这里的内容就是针对请求命令应答的具体数据.
            statuscode:gateway执行命令成功值为0,否则为非0
```
### 3. server发送配置gateway命令 init/set.do:
```
        发送的消息命令如下:
        {
                "msgtype": "request",
          	    "request": {
                    "cmd": "init/set.do",
                    "requestid": "1234567890-0987",
                    "data": {
                        "_client_gateway": "192.168.64.1",
                        "_client_ip": "192.168.64.1",
                        "_client_netmask": "255.255.255.0",
                        "_interface_inet": "static",
                        "_password": "passwd",
                        "_server_ip": "111.222.11.22",
                        "_server_name": "server_name",
                        "_server_port": "1883",
                        "_username": "username"
                    },
                    "return": [
                        "requestid"
                    ],
                    "timestamp": 1470901793
            }
        }
        解释:
            此消息配置gateway的ip和broker信息
            _interface_inet:值为static或dhcp
            _server_ip:指broker的地址
            _server_name:为iot-server的名字,和topic有关.
            命令成功会重启gateway,应用新的配置.
        
```
### 4. 设置自动上报间隔时间命令 manager/set_interval.do:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "manager/set_interval.do",
                "requestid": "1234567890-0987",
                "data":
                    {
                        "_interval":0
                    },
                "timestamp": 1470901793
            }
        }
        解释:
            _interval为int类型.值为0代表禁止自动上报数据,其他值为自动上报的间隔时间,单位秒.

```
### 5. 获取gateway支持的设备列表命令 manager/get_suppot_devlist:
```
        发送的消息命令如下:
        {
           "msgtype": "request",
                "request": {
                "cmd": "manager/get_suppot_devlist",
            	"requestid":"1234567890-0987",
                 "timestamp": 1470901793
            }
        }

```
### 5. 获取gateway通讯接口命令 manager/list_commif.do:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "manager/list_commif.do",
            	"requestid":"1234567890-0987",
                "return": [
                    "requestid"
                ],
                "timestamp": 1470901793
            }
        }
        解释:
        这里的通讯接口的含义一般是指物理通信接口,类似rs485,rs232,和具体的gateway硬件相关.
```
### 6. 设置更新gateway通讯接口命令 manager/update_commif.do:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
                "request": {
                "cmd": "manager/update_commif.do",
            	"requestid":"1234567890-0987",
                "data": {
        			"rs485-1":"/dev/ttyS0",
        			"rs485-2":"/dev/ttyS1",
        			"rs485-3":"/dev/ttyS2"
                 },
                "return": [
                    "requestid"
                ],
                "timestamp": 1470901793
            }
        }
        解释:
        此命令一般不用,gateway的通讯接出厂已经设置完成,供调试人员用.
```
### 7. 设置gateway系统时间 manager/set_system_time:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "manager/set_system_time",
            	"requestid":"1234567890-0987",
                    "data": {
                        "date": "12/02/2017",
                        "time": "15:57:30"
                    },
                "timestamp": 1470901793
            }
        }
        解释:
        如果gateway能连接公网,则会自动与时间服务器同步时间,如不连接公网,则用此命令校时.
```
### 8. 添加更新gateway下的设备 manager/dev/update.do:
```
        发送的消息命令如下:
        {
                "msgtype": "request",
                "request": {
                    "cmd": "manager/dev/update.do",
                    "requestid": "1234567890-0987",
                    "data":    {
                        "_devid": "xmz-00",
                        "_conn":{
                        	"devaddr":"1",
                        	"commif":"rs485-1",
                        	"BaudRate":9600,
                        	"DataBits":8,
                        	"Parity":"N",
                        	"StopBits" :1
                             },
                        "_type":"RSBAS"
                        },
                    "return": [
                    "requestid",
                    "cmd"
                    ],
                "timestamp": 1470901793
            }
        }

        解释:
        data域内重要的字段,带"_"开头的,是必须要的
        _devid是系统给设备定义的唯一的id,
        _conn内是该设备所必须要的通信参数和配置参数,可以通过设备的帮助文件或设备的
        文档获知.
        _type是设备的类型.
        例子添加了一个RSBAS类型的设备,id是xmz-00,此设备的物理通信地址是1.
```
### 9. 删除gateway下的设备 manager/dev/delete.do:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "manager/dev/delete.do",
                "requestid": "1234567890-0987",
                "data":  {
                    "_devid":"xmz-00"
                    },
                "return": [
                "requestid",
                "cmd"
                ],
                "timestamp": 1470901793
            }
       }
       解释:
       删除了id为"xmz-00"的设备.
```
### 10. 读取gateway下的设备列表 manager/dev/list.do:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "manager/dev/list.do",
        	    "requestid":"1234567890-0987",
                "return": [
                    "requestid",
                    "cmd"
                ],
                "timestamp": 1470901793
            }
        }
        解释:
        读取gateway下的具体设备的列表和设备接口配置信息.
```
### 11. 读取设备参数 do/getvar:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "do/getvar",
                "requestid": "1234567890-0987",
                "data": {
                    "_devid": "xmz-00",
            		"starting_address":0,
            		"quantity":2
                    } ,
                "return": [
                    "requestid",
                    "cmd"
                ],
                "timestamp": 1470901793
            }
        }
        解释:
        读取设备的实时数据,data内的参数根据不同设备类型而不同
```
### 12. 操作设备(写设备变量数据) do/setvar:
```
        发送的消息命令如下:
        {
            "msgtype": "request",
            "request": {
                "cmd": "do/setvar",
                "requestid": "1234567890-0987",
                "data": {
                    "_devid" : "modebusrtu-01",
                    "Function_code" : 16,
                    "Quantity" : 10,
                    "Starting_address" : 0,
                    "value" : [1,2,3,4,5,6,7,8,9,10]
                    } ,
                "return": [
                    "requestid",
                    "cmd"
                ],
                "timestamp": 1470901793
            }
        }
        解释:
        写设备变量数据,data内的参数根据不同设备类型而不同,这个例子操作了一个modbusrtu
        设备,使用功能码16,设备的起始地址0,写10个寄存器,值为1,2,3,4,5,6,7,8,9,10.
```
### gateway在线离线的判断
* mqtt有个retain 标记,意思的服务器会保留这个标记的消息,与willmessage配合起来后,就能判断是否在线.
```     
        当gateway一上线,就会推送下面的消息:
        {
            {
              "header" : {
                "from" : {
                  "_devid" : "iot_gate",
                  "_model" : "TG120",
                  "_runstate" : "127",
                  "_version" : "1.0"
                },
            "msgtype": "update"
          },
          "request": {
            "timestamp": 1507971399,
            "cmd": "push/state.do",
            "data": "1",
            "statuscode": 0
          }
        }
        解释:data=1代表是在线,当gateway离线了,服务器将会把gateway的willmessage推送,此时data=0.
```




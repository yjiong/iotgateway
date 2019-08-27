# IOTD JSON API 
## 基本格式

本 API 的传输消息使用 JSON 编码，共有 2 种消息，一种是请求消息 (返回中不包含 `statuscode` 数据项)，另一种是应答消息（返回中包含 `statuscode` 数据项)。

消息中的数据项名及命令名约定都用小写形式（数据项 data 内的项名可不受该约定限制）。

这两种消息都统一使用如下 JSON 格式：

```json
{
    // 即本消息使用的 API 协议名及版本
    "api": "IOTD/0.9",

    // 消息发送者的编号
    "sender": "消息发送者的标识或 ID",

    // 消息发送者的设备模型名
    // 例如:
    //      当网关为发送者时，值比如可为 "GW-XXXX",
    //      当服务器为发送者时，值可为 "IOTD Server" 等
    "model": "GW-485",

    // 在请求消息中时，该值为请求的命令名，
    // 在应答消息中时，该值为其对应请求消息中的命令名
    "cmd": "help",

    // 在请求消息中时，该值为发送方生成的一个序列号，值为 uint 型。
    // 在应答消息中时，该值为其对应请求消息中的 seq 值。
    // 一般可利用该值来匹配请求包的应答包。
    // 同时作为一种简单的校验机制，约定：
    //     由 IOTD 服务器主动发送的请求消息中，序列号都为偶数，
    //     而由网关主动发送的请求消息中，序列号都为奇数。
    // 该项为可选项，当命令无需应答时，可以不添加该数据项，
    // 例如自动上报命令 push /dev/vars 等中可以不用添加该项。
    "seq": 12345,

    // 本消息产生时的时间戳
    "tstamp": 15222222,

    // 如果是应答消息，还需要返回 `statuscode`，
    // 本版本中去除了 `mtype` 数据项，因此需要根据是否包含 `statuscode` 数据项
    //来判断前是请求消息包，还是应答消息包。
    // 值参考 HTTP 返回的状态码，例如 200 为成功。
    // 成功时，信息在 `data` 中返回，失败时，错误信息在 `error` 中返回
    // 一条消息中 `data` 数据项和 `error` 数据项两者只能有一个
    "statuscode": 200,

    // 如果是网关发送的消息包，还需要返回 `ctag` (表示config tag 或 change tag)， 值为字符串。
    // 当网关中的设备有更新（比如直接通过调试工具添加或删除过设备等），该值会相应更新。
    // 服务端程序一般根据该值判断网关的连接设备是否有更新。
    "ctag": "127",


    // 成功时，信息在 `data` 中返回
    // 消息的数据部分，为可选项，值类型不定，可以是 int, str, array, dict 等。
    "data": {},

    // 失败时，错误信息在 `error` 中返回
    "error" : "Device not found"
}
```

## 消息命令

### 1. 帮助命令 "help"

该命令消息由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端发起的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "help",
    "seq": 12340,
    "tstamp": 1511316809
}
```

data 项的值为字符串时，可指定一个具体的设备，用来返回指定设备的帮助信息;当值是一个字符串数组时，表示指定一组设备，用来返回一组设备的帮助信息; 如果没有 data 项，或者求值为假（例如空字符串），则返回网关的帮助信息。


网关返回的应答举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXX",
    "cmd": "help",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": "help 返回的各种帮助信息"
}
```

### 2. 服务端获取网关配置信息命令 "get /sys/info"

该命令消息由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端发起的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "get /sys/info",
    "seq": 12340,
    "tstamp": 1511316809
}
```

网关发送的应答举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "get /sys/info",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": {
        "devid": "go_iot_gate",
        "ip": "192.168.64.1",
        "gateway": "网络的网关地址", 
        "netmask": "网络掩码", 
        "iface": "网络通信接口名",
        "inet": "static或dhcp", 
        "dns": "域名服务器",
        "net_status": "online",

        "proto": "MQTT/3.1.1",
        "mqtt_svr_ip": "mqtt服务地址", 
        "mqtt_svr_port": "mqtt服务 端口", 
        "mqtt_topic": "mqtt接收topic,比如things", 
        "mqtt_usr": "mqtt登录用户名",
        "mqtt_pwd": "mqtt登录密码", 
        "mqtt_will": "1",

        "interval": "10",
        "keepalive": "60"
    }
}
```

### 3. 服务端配置网关命令 "set /sys/info"

该命令消息由 IOTD 服务端发起请求，网关无应答。

IOTD 服务端发起的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "set /sys/info",
    "seq": 12340, // 可选项
    "tstamp": 1511316809,
    "data": {
        "ip": "192.168.64.1",
        "gateway": "网络的网关地址", 
        "netmask": "网络掩码", 
        "iface": "网络通信接口名",
        "inet": "static或dhcp", 

        "mqtt_svr_ip": "mqtt服务地址", 
        "mqtt_svr_port": "mqtt服务 端口", 
        "mqtt_topic": "mqtt接收topic,比如things", 
        "mqtt_usr": "mqtt登录用户名",
        "mqtt_pwd": "mqtt登录密码", 
        "mqtt_will": "1",

        "keepalive": 60  // 可选项
    }
}
```

命令成功后网关会自动重启，并应用新的配置。该请求没有应答。


### 4. 设置自动上报间隔时间命令 "set /sys/interval"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "set /sys/interval",
    "seq": 12340,
    "tstamp": 1511316809,
    "data": {
        "interval":10
    }
}
```

data 数据项为 int 类型，值为 0 代表禁止自动上报，其他值为自动上报的间隔时间， 单位为秒。

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "set /sys/interval",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data":"设置成功",
}
```

若设置成功，则应答消息中的 `data` 项值和设置时的值相同 。


### 5. 获取网关支持的设备列表命令 "list /dev/supported"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "list /dev/supported",
    "seq": 12340,
    "tstamp": 1511316809
}
```

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "list /dev/supported",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": [
        "DTSD422",
        "FUJITSU",
        "ModbusRtu",
        "ModbusTcp",
        "TC100R8",
        "QDSL_SM510",
        "RSBAS",
        "TEST_GO",
        "XP_YSHY"
    ]
}
```

### 6. 获取网关通讯接口命令 "list /sys/commif"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "list /sys/commif",
    "seq": 12340,
    "tstamp": 1511316809
}
```

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "list /sys/commif",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": {
        "rs485-1": "/dev/ttyS0",
        "rs485-2": "/dev/ttyS1",
        "rs485-3": "/dev/ttyS2"
    }
}
```

这里的通讯接口的含义一般是指物理通信接口，类似 rs485， rs232 和具体的网关硬件相关。

### 7. 更新网关通讯接口命令 "set /sys/commif"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "set /sys/commif",
    "seq": 12340,
    "tstamp": 1511316809,
    "data": {
        "rs485-1": "/dev/ttyS0",
        "rs485-2": "/dev/ttyS1",
        "rs485-3": "/dev/ttyS2"
    }
}
```

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "set /sys/commif",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": "設置通信接口成功"
}
```

设置成功后，返回的 `data` 项值与设置值相同。

此命令一般不用，网关的通讯接口在出厂时就已经设置完成。仅供调试使用。

### 8. 设置网关系统时间 "set /sys/time"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "set /sys/time",
    "seq": 12340,
    "tstamp": 1511316809,
    "data": "2017-12-19T02:07:28.492Z"
}
```

时间值格式 "2017-12-19T02:07:28.492Z" 遵循 [RFC 3339](https://www.ietf.org/rfc/rfc3339.txt)，也就是 ISO 时间格式。时间值 "2017-12-19T02:07:28.492Z" 转换成北京时间是 "2017-12-19 10:07:28 492毫秒"，因为北京时间与 ISO 时间有 8 小时偏移。

在 JavaScript 中，`new Date(Date.now()).toISOString()` 直接能返回 RFC 3339 格式的当前时间字符串。


如果网关能连接公网，则会自动与时间服务器同步时间，若不能连接公网，则用此命令校时。

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "set /sys/time",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": "系统时间设置成功"
}
```

设置成功后，返回的 `data` 项值与设置值相同。


### 9. 添加和更新连接到网关的设备 "update /dev/item"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "add /dev/item",
    "seq": 12340,
    "tstamp": 1511316809,
    "data": {
        "devid": "xmz-00",
        "conn":{
            "devaddr":"1",
            "commif":"rs485-1",
            "BaudRate":9600,
            "DataBits":8,
            "Parity":"N",
            "StopBits" :1
             },
        "type":"RSBAS"
    }
}
```

+ `devid`: 系统给设备定义的唯一 id
+ `conn`: 该设备所必须要的通信参数和配置参数，可以通过设备的帮助文件或文档获知
+ `type`: 设备的类型

本例中添加了一个 RSBAS 类型的设备，id 值是 'xmz-00'，此设备的物理通信地址是 "1"。

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "add /dev/item",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": {
        "设备 : xmz-00更新成功"
    }
}
```


### 10. 删除连接到网关的设备 "del /dev/item"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "del /dev/item",
    "seq": 12340,
    "tstamp": 1511316809,
    "data": {
        "devid": "xmz-00"
    }
}
```


+ `devid`: 系统给设备定义的唯一 id

本例中删除了一个 id 值是 'xmz-00' 设备。

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "del /dev/item",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": {
        "设备 : xmz-00删除成功"
    }
}
```



### 11. 读取网关下的设备列表 "list /dev/items"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "list /dev/items",
    "seq": 12340,
    "tstamp": 1511316809
}
```

网关的应答包举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "GW-XXXX",
    "cmd": "list /dev/items",
    "seq": 12340,
    "tstamp": 1511316809,
    "statuscode": 200,
    "ctag": "127",
    "data": [
        {
          "conn": {
              "commif": "rs485-1",
              "devaddr": "8"
          },
          "devid": "test-dev",
          "type": "TEST_GO"
      }
    ]
}
```


该命令用来读取网关下的具体设备的列表和设备接口配置信息。

### 12. 读取设备参数的实时值 "get /dev/var"

该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
        "api": "IOTD/0.9",
        "sender": "SRV001",
        "model": "IOTD 001",
        "cmd": "get /dev/var",
        "seq": 12340,
        "tstamp": 1511316809,
        "data": {
            "devid": "xmz-00",
            "starting_address":0,
            "quantity":2
        }
    }
}
```

读取设备的实时数据，`data` 内的参数根据不同设备类型而不同。


网关的应答举例：

```json
```

### 13. 设置设备参数的值(写设备变量数据) "set /dev/var"


该命令由 IOTD 服务端发起请求，网关进行应答。

IOTD 服务端的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "set /dev/var",
    "seq": 12340,
    "tstamp": 1511316809,
    "data": {
        "devid": "modebusrtu-01",
        "Function_code": 16,
        "Quantity": 10,
        "Starting_address": 0,
        "value": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    }
}
```

写设备变量数据时，`data` 内的参数根据不同设备类型而不同，本例操作了一个 modbusrtu 设备，使用功能码 16，设备的起始地址 0，写 10 个寄存器， 值为 1,2,3,4,5,6,7,8,9,10。


网关的应答举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "go_iot_gate",
    "model": "MQTT-GATEWAY",
    "cmd": "set /dev/var",
    "seq": 12340,
    "ctag": "11",
    "statuscode": 200,
    "data": {
        "devid": "modebusrtu-01"
    },
    "tstamp":1514191772
}
```


### 14. 网关程序在线更新命令 "upgrade"

该命令消息由 IOTD 服务端发起请求，网关无应答。

IOTD 服务端发起的请求举例：

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "upgrade",
    "seq": 12340,
    "tstamp": 1511316809,
    "data":{
        // 数据体部分的编码类型，是一个可选项，
        // 可根据 `enctype` 判断是否有编码，值为：
        //      plain: 无编码，
        //      gzip:　GNU zip 编码
        //      compress: 用 Unix 的文件压缩程序压缩
        //      deflate:　用 zlib 的格式压缩
        // 如果缺失，则默认值为 "plain"。
        //"enctype": "gzip",
        "url": "http://xxx.xxx.xxx/firmware"
    }
}
```

网关的应答举例：

```json
{
  "api" : "1.0.0",
  "cmd" : "upgrade",
  "ctag" : "8",
  "data" : "iotgateway will be restart to update firmware !",
  "model" : "MQTT-GATEWAY",
  "sender" : "gatewayTest",
  "seq" : 12340,
  "statuscode" : 200,
  "tstamp" : 1520219225
}
```
命令成功后网关会自动重启，并应用新的程序。

### 15. 读取历史记录 get /dev/history:
        发送的消息命令如下:
```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "get /dev/history",
    "seq": 12340,
    "tstamp":1514191772
    "data": {
        "_devid" : "modebusrtu-01",
        "since":   "2018-10-23 12:30:30",
        "until":   "2018-10-24 23:30:30",
        "cmdtype": "do/setvar"
        } 
}
```
        解释:
        "_devid":  "设备id,若不填写该字段,就读取管理命令的历史记录"
		"since":   "启始时间(2018-10-23 12:30:30)"
		"until":   "结束时间(2018-10-24 24:30:30)"
		"cmdtype": "若不带该字段读取管理命令历史时,默认为查询do/setvar的历史命令记录,
		            读设备历史时,默认为查询do/auto_up_data的历史记录,
		            支持模糊查询"

### 16. 设置扩展参数 set /sys/external:
        发送的消息命令如下:
```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "set /sys/external",
    "seq": 12340,
    "tstamp":1514191772
    "data": {
          "historydays":    "90",
          "postgresql_dns": "postgres://postgres:P@ssw0rd@211.159.217.108:port/database"
        } ,
}
```
        解释:
        扩展参数一般不需要设置,特殊情况下使用.
        "historydays":    "可保存历史记录的时间跨度,默认为100天",
		"postgresql_dns": "数据库服务地址 e.g.,postgres://user:password@hostname:port/database"

### 17. 添加计划控制命令 add /sys/schedule:
        发送的消息命令如下:
```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "add /sys/schedule",
    "seq": 12340,
    "tstamp":1514191772
    "data": {
           "_devid": "lxzfl-dn15-01",
           "2019-04-16 10:20 *":{"关阀":""},
           "2019-04-* 10:20 Tue":{"关阀":""},
           "2019-04-16,17,18 10,11,12:05,10,15,20,25 *":{"关阀":""},
           "2019-09,10,11-* 08,09:05,10 Mon,Fri":{"关阀":""}
        }
}
```
        解释:
        格式为:{"计划控制命令的时间点":{"要执行的命令":"命令的参数"}}}
        举例说明:
		   "2019-04-16 10:20 *":{"关阀":""},
		   表示2019年4月16日10点20分执行对设备id是lxzfl-dn15-01(这里是预付费水表)关阀.
		   注意格式,年月日以-分割,时分是以:分割,年月日和时分之间是空格.
		   "2019-04-* 10:20 Tue":{"关阀":""} ,
		   表示2019年4月的每个星期2的10点20执行一次关阀
		   "2019-04-16,17,18 10,11,12:05,10,15,20,25 *":{"关阀":""},
		   表示2019年4月的16,17,18日的10,11,12点的05,10,15,20,25分都执行一次关阀
		   "2019-09,10,11-* 08,09:05,10 Mon,Fri":{"关阀":""}
		   表示2019年9,10,11月的每个星期一和星期五的8,9点的05,10分都执行一次关阀
		   注:若日期定义了,那星期必为*
		      反之星期定义了,日期必为*
		
### 18. 删除设备的计划控制命令 del /sys/schedule:
        发送的消息命令如下:
```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "del /sys/schedule",
    "seq": 12340,
    "tstamp":1514191772
    "data": {
       "devid": "lxzfl-dn15-01"
    }
}
```
        解释:
         删除设备id是lxzfl-dn15-01的计划控制命令

### 19. 查看计划控制命令 list /sys/schedule:
        发送的消息命令如下:
```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "list /sys/schedule",
    "seq": 12340,
    "tstamp":1514191772
}
```
        解释:
         查看已经添加的所有设备的计划控制命令

### 20. 主动上报网关在线离线状态命令 "push /sys/state"

当网关一上线，就会推送下面的消息:

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "push /sys/state",
    "seq": 12340,
    "tstamp": 1511316809,
    "ctag": "127",
    "data": "1"
}
```


### 21. 定时上报网关设备的当前值命令 "push /dev/vars"

如果开启了定时上报，网关会定时推送下面的消息:

```json
{
    "api": "IOTD/0.9",
    "sender": "SRV001",
    "model": "IOTD 001",
    "cmd": "push /dev/vars",
    "seq": 12340,
    "tstamp": 1511316809,
    "ctag": "127",
     "data": {
        "devid": "test-dev",
        "设置数": 0,
        "递增数": 74
    }
}
```


## 网关命令列表

1. "help": 获取帮助信息，无 data 数据项时返回网关帮助信息，data 字段值为设备名，则返回设备的帮助信息。
1. "get /sys/info": 系统信息获取。
1. "set /sys/info": 系统初始化设置，需要 data 数据项。
1. "set /sys/interval": 设置自动读取设备的间隔时间，单位为秒，值为 0 时表示不自动循环读取，需要 data 数据项。
1. "list /dev/supported": 获取网关所支持的设备。
1. "list /sys/commif": 获取网关的通信接口信息。
1. "set /sys/commif": 设置网关的通信接口，需要 data 数据项。注: 出厂前已设定，一般无需设置，供内部调试使用。
1. "set /sys/time": 网关校时，需要 data 数据项。
1. "add /dev/item": 添加设备，需要 data 数据项。
1. "del /dev/item": 删除设备，需要 data 数据项。
1. "list /dev/items": 获取当前设备列表。
1. "get /dev/var": 读取设备实时数据，是否需要 data 数据项，详见设备的帮助信息。
1. "set /dev/var": 设置设备数据值，需要 data 数据项，详见设备的帮助信息。
1. "upgrade": 网关程序在线更新命令。
1. "get /dev/history": 读取历史记录 
1. "set /sys/external": 设置扩展参数
1. "add /sys/schedule": 添加计划控制命令 
1. "del /sys/schedule": 删除设备的计划控制命令
1. "list /sys/schedule": 查看计划控制命令 
---
1. "push /dev/vars": 定时上报网关设备的当前值。
1. "push /sys/state": 主动上报网关在线离线状态。

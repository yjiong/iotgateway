###初始化命令
```json
{
    "header": {
        "from": {
            "devid": "XXX0001",
            "id": "ABCDEFG",
            "model": "server",
            "version": "v1.0"
        },
        "msgtype": "request"
    },
    "request": {
        "cmd": "do/setvar",
        "requestid": "1234567890-0987",
        "data": {
            "_devid": "DDZY422N-01",
            "初始化":{
            		"报警金额":20//int类型0-9999
            		"报警负荷":10   //float类型0.01-99.99
            		"允许透支金额":10 //int类型0-9999
            		"允许囤积金额":1000 //float类型 0.01-999999.99
            		"尖单价":1.55//float类型0.01-99.99
            		"峰单价":1.25//float类型0.01-99.99
            		"平单价":1.05 //float类型0.01-99.99
            		"谷单价":0.95 //float类型0.01-99.99
            		}
        },
        "return": [
            "requestid",
            "cmd"
        ],
        "timestamp": 1470901793
    }
}
```
###充值命令
```json
{
    "header": {
        "from": {
            "devid": "XXX0001",
            "id": "ABCDEFG",
            "model": "server",
            "version": "v1.0"
        },
        "msgtype": "request"
    },
    "request": {
        "cmd": "do/setvar",
        "requestid": "1234567890-0987",
        "data": {
            "_devid": "DDZY422N-01",
            "充值":88.5 //float类型0.01-9999.99

        },
        "return": [
            "requestid",
            "cmd"
        ],
        "timestamp": 1470901793
    }
}
```
###退款命令
```json
{
    "header": {
        "from": {
            "devid": "XXX0001",
            "id": "ABCDEFG",
            "model": "server",
            "version": "v1.0"
        },
        "msgtype": "request"
    },
    "request": {
        "cmd": "do/setvar",
        "requestid": "1234567890-0987",
        "data": {
            "_devid": "DDZY422N-01",
            "退款":55.8 //float类型0.01-9999.99

        },
        "return": [
            "requestid",
            "cmd"
        ],
        "timestamp": 1470901793
    }
}
```
###强制合闸,断闸和撤销强制命令
```json
{
    "header": {
        "from": {
            "devid": "XXX0001",
            "id": "ABCDEFG",
            "model": "server",
            "version": "v1.0"
        },
        "msgtype": "request"
    },
    "request": {
        "cmd": "do/setvar",
        "requestid": "1234567890-0987",
        "data": {
            "_devid": "DDZY422N-01",
              "强制合闸":"" //无需参数,三种命令:"强制合闸","强制断闸","撤销强制"

        },
        "return": [
            "requestid",
            "cmd"
        ],
        "timestamp": 1470901793
    }
}
```
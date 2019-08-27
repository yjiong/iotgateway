# Iot-GATEWAY

### 先简单写个readme 以后会详细介绍

## Installation

### ubuntu and debian:

arm64安装包[arm64.deb]

armhf安装包[armhf.deb]

amd64安装包[amd64.deb]

uname -m 查看你的cpu构架,选择对应的deb包

```sh
sudo apt-get update
sudo apt-get --no-install-recommends -y install net-tools postgresql
dpkg -i gateway-xxx-v1.3.deb
```


## Usage example

_服务和iot网关的通信报文举例详见wiki
_For more examples and usage, please refer to the [Wiki][wiki]._


## Release History

* 1.3
    * CHANGE: add rest api


## Meta
jiong yao – yjiong@msn.com

Distributed under the XYZ license. See ``LICENSE`` for more information.

[https://github.com/yjiong/iotgateway](https://github.com/yjiong/)

## Contributing

1. Fork it (<https://github.com/yjiong/iotgateway/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

<!-- Markdown link & img dfn's -->
[wiki]: https://github.com/yjiong/iotgateway/wiki
[armhf.deb]:https://github.com/yjiong/iotgateway/releases/download/v1.3/gateway-armhf-v1.3.deb
[arm64.deb]:https://github.com/yjiong/iotgateway/releases/download/v1.3/gateway-arm64-v1.3.deb
[amd64.deb]:https://github.com/yjiong/iotgateway/releases/download/v1.3/gateway-amd64-v1.3.deb

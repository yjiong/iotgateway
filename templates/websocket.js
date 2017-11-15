
window.onload = function() {
  webSocket();
};

var socket;
var url = "ws" + document.location.href.substring(4) + "message";
socket = new WebSocket(url);

function webSocket() {
  socket.onmessage = function (eve) {
    var data = JSON.parse( eve.data );
    if ( data.request.requestid === "webSend" ) {
      WebSocketTest( data );
    }
  };

  socket.onopen = function (event) {
    // alert('连接成功');
    socket.send(JSON.stringify({"request": {"return": ["requestid"], "requestid": "webSend", "cmd": "init/get.do", "data": {}}}))
  };

  socket.onclose = function (event) {
    alert('已断开连接，请在重启完成后使用新的IP地址或者刷新页面');
  };

  socket.onerror = function (event) {
    alert('连接失败，请刷新页面重试！');
  };
}

function onSubmit() {
  var start = document.getElementById("dhcp_start").checked;
  var stop = document.getElementById("dhcp_stop").checked;
  var params = start ? {
    "_server_ip": document.getElementById("hostname").value,
    "_server_port": document.getElementById("port").value,
    "_username": document.getElementById("username").value,
    "_password": document.getElementById("password").value,
    "_server_name": document.getElementById("topic").value,
    // "_interval": document.getElementById("publish_interval").value,
    "_interface_inet": "dhcp"
  } : {
    "_server_ip": document.getElementById("hostname").value,
    "_server_port": document.getElementById("port").value,
    "_username": document.getElementById("username").value,
    "_password": document.getElementById("password").value,
    "_server_name": document.getElementById("topic").value,
    // "_interval": document.getElementById("publish_interval").value,
    "_client_ip": document.getElementById("Local_IP").value,
    "_client_netmask": document.getElementById("Mask").value,
    "_client_gateway": document.getElementById("Gateway").value,
    "_interface_inet": "static"
  }
  var text = JSON.stringify({
    "msgtype": "request",
    "request": {
      "cmd": "init/set.do",
      "requestid": "webSet",
      "data": params,
      "return": ["requestid"]
    }
  });
  if ( document.getElementById("Mask").value === '' && stop === true ) {
    alert( "DHCP停用后，必须填写子网掩码" );
  } else {
    socket.send(text);
    // alert( '已发送，请等待！' )
  }
}

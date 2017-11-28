
function onload() {
  window.location.reload();
  if ( document.getElementById( 'dhcp_start' ) &&
    document.getElementById( 'dhcp_start' ).getAttribute( 'checked' ) === 'checked' ) {
    document.getElementById( 'Local_IP' ).setAttribute( 'disabled', true );
    document.getElementById( 'Mask' ).setAttribute( 'disabled', true );
    document.getElementById( 'Gateway' ).setAttribute( 'disabled', true );
  }
}

function onSwitch( type ) {
  if ( type !== 'start' ) {
    document.getElementById( 'Local_IP' ).removeAttribute( 'disabled' );
    document.getElementById( 'Mask' ).removeAttribute( 'disabled' );
    document.getElementById( 'Gateway' ).removeAttribute( 'disabled' );
  }
  else {
    document.getElementById( 'Local_IP' ).setAttribute( 'disabled', true );
    document.getElementById( 'Mask' ).setAttribute( 'disabled', true );
    document.getElementById( 'Gateway' ).setAttribute( 'disabled', true );
  }
}

function WebSocketTest( data ) {
  // 处理数据,将对应的数据加载到html中
  document.getElementById("deviceCode").innerHTML = data.header.from["_devid"];
  document.getElementById("deviceModel").innerHTML = data.header.from["_model"];
  document.getElementById("softVersion").innerHTML = data.header.from["_version"];
  document.getElementById("runState").innerHTML = data.header.from["_runstate"];
  document.getElementById("Local_IP").value = data.response.data["_client_ip"];
  // document.getElementById("Mask").value = data.response.data["_client_netmask"];
  // document.getElementById("Gateway").value = data.response.data["_client_gateway"];
  document.getElementById("hostname").value = data.response.data["_server_ip"];
  document.getElementById("port").value = data.response.data["_server_port"];
  document.getElementById("username").value = data.response.data["_username"];
  document.getElementById("password").value = data.response.data["_password"];
  document.getElementById("topic").value = data.response.data["_server_name"];
  document.getElementById("publish_interval").value = data.response.data["_interval"];
}



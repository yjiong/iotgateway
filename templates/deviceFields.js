// 每个设备类型都要加devid字段
const deviceFields = {
  'ElectricMeter' : ['commif','devid','dname','devaddr','devtype'],
  'DLT645_2007'   : ['commif','devid','dname','devaddr','baudRate','dataBits','parity','stopBits'],
  'ModbusRtu'     : ['commif','devid','dname','devaddr','baudRate','dataBits','functionCode','parity',
                     'quantity','startingAddress','stopBits'],
  'ModbusTcp'     : ['commif','devid','dname','devaddr','functionCode','quantity','startingAddress'],
  'TC100R8'       : ['commif','devid','dname','devaddr'],
  'WaterMeter'    : ['commif','devid','dname','devaddr','devtype','fcode']
};

const FieldsMap = {
  'ElectricMeter' : {'commif':'string','dname':'string','devaddr':'string','devtype':'string'},
  'DLT645_2007'   : {'commif':'string','dname':'string','baudRate':'string','dataBits':'string','devaddr':'string','parity':'string','stopBits':'string'},
  'ModbusRtu'     : {'commif':'string','dname':'string','baudRate':'string','dataBits':'string','devaddr':'string','functionCode':'string','parity':'string',
                     'quantity':'string','startingAddress':'string','stopBits':'string'},
  'ModbusTcp'     : {'commif':'string','dname':'string','devaddr':'string','functionCode':'int','quantity':'int','startingAddress':'int'},
  'TC100R8'       : {'commif':'string','dname':'string','devaddr':'string'},
  'WaterMeter'    : {'commif':'string','dname':'string','devaddr':'string','devtype':'string','fcode':'string'}
};

const FieldsValidata = {
  'ElectricMeter' : {
    'devid': {name:'设备编号',type:'string',max:32},
    'commif': {name:'连接方式',type:'string'},
    'dname': {name:'设备名称',type:'string',max:64},
    'devaddr': {name:'设备地址',type:'string'},
    'devtype': {name:'设备类型',type:'string'},
  },
  'DLT645_2007'   : {
    'devid': {name:'设备编号',type:'string',max:32},
    'commif': {name:'连接方式',type:'string'},
    'dname': {name:'设备名称',type:'string',max:64},
    'baudRate': {name:'波特率',type:'int', min:0, max:4294967295},
    'dataBits': {name:'dataBits',type:'int', min:0, max:4294967295},
    'devaddr': {name:'设备地址',type:'string'},
    'parity': {name:'parity',type:'string'},
    'stopBits': {name:'stopBits',type:'int', min:0, max:4294967295},
  },
  'ModbusRtu'     : {
    'devid': {name:'设备编号',type:'string',max:32},
    'commif': {name:'连接方式',type:'string'},
    'dname': {name:'设备名称',type:'string',max:64},
    'baudRate': {name:'波特率',type:'int', min:0, max:4294967295},
    'dataBits': {name:'dataBits',type:'int', min:0, max:4294967295},
    'devaddr': {name:'设备地址',type:'string'},
    'functionCode': {name:'functionCode',type:'int', min:0, max:22},
    'parity': {name:'parity',type:'string'},
    'quantity': {name:'quantity',type:'int', min:0, max:4294967295},
    'startingAddress': {name:'startingAddress',type:'int', min:0, max:4294967295},
    'stopBits': {name:'stopBits',type:'int', min:0, max:4294967295},
  },
  'ModbusTcp'     : {
    'devid': {name:'设备编号',type:'string',max:32},
    'commif': {name:'连接方式',type:'string'},
    'dname': {name:'设备名称',type:'string',max:64},
    'devaddr': {name:'设备地址',type:'string'},
    'functionCode': {name:'functionCode',type:'int', min:0, max:22},
    'quantity': {name:'quantity',type:'int', min:0, max:4294967295},
    'startingAddress': {name:'startingAddress',type:'int', min:0, max:4294967295},
  },
  'TC100R8'       : {
    'devid': {name:'设备编号',type:'string',max:32},
    'commif': {name:'连接方式',type:'string'},
    'dname': {name:'设备名称',type:'string',max:64},
    'devaddr': {name:'设备地址',type:'string'},
  },
  'WaterMeter'    : {
    'devid': {name:'设备编号',type:'string',max:32},
    'commif': {name:'连接方式',type:'string'},
    'dname': {name:'设备名称',type:'string',max:64},
    'devaddr': {name:'设备地址',type:'string'},
    'devtype': {name:'设备类型',type:'string'},
    'fcode': {name:'fcode',type:'string'}
  }
};

// 设备修改的接口path
const devicePath = {
  'ElectricMeter' :'meter_add_or_uppdate',
  'DLT645_2007'   :'meter_dtl645_2007_add_or_uppdate',
  'ModbusRtu'     :'modbusrtu_add_or_uppdate',
  'ModbusTcp'     :'modbustcp_add_or_uppdate',
  'TC100R8'       :'tc100r8_add_or_uppdate',
  'WaterMeter'    :'watermeter_add_or_uppdate',
};

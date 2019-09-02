package devapi

import (
	"encoding/json"
	"golang.org/x/net/context"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
	//log "github.com/sirupsen/logrus"

	sjson "github.com/bitly/go-simplejson"
	log "github.com/sirupsen/logrus"
	pb "github.com/yjiong/iotgateway/api"
	"github.com/yjiong/iotgateway/internal/device"
	"github.com/yjiong/iotgateway/internal/gateway"
)

//ModbusTcpapi ..
type ModbusTcpapi struct {
	gw *gateway.Gateway
}

// NewModbusTcpapi creates a new ApplicationAPI.
func NewModbusTcpapi(gateway *gateway.Gateway) *ModbusTcpapi {
	return &ModbusTcpapi{
		gw: gateway,
	}
}

//ModbusTCPUpdate ....
func (p *ModbusTcpapi) ModbusTCPUpdate(ctx context.Context, req *pb.ModbusTcpUpdateRequest) (*pb.ModbusTcpUpdateResponse, error) {
	gateway.GrpcMsg = "req"
	defer func() {
		gateway.GrpcMsg = nil
	}()
	conn := map[string]interface{}{
		device.DevAddr:    req.Devaddr,
		"commif":          req.Commif,
		"FunctionCode":    req.FunctionCode,
		"StartingAddress": req.StartingAddress,
		"Quantity":        req.Quantity,
		device.DevName:    req.Dname,
	}
	jsreq := map[string]interface{}{
		"data": map[string]interface{}{
			device.DevType: "ModbusTcp",
			device.DevID:   req.Devid,
			device.DevConn: conn,
		},
	}
	breq, _ := json.Marshal(jsreq)
	jsonreq, _ := sjson.NewJson(breq)
	go p.gw.DB.InsertDevJdoc("cmdhistory", "api/DevUpdate", jsonreq)
	p.gw.DevUpdate(jsonreq, nil)
	log.Infoln(jsonreq)
	var err error
	if result, ok := gateway.GrpcMsg.(string); !ok {
		err, _ = gateway.GrpcMsg.(error)
	} else {
		pbres := pb.ModbusTcpUpdateResponse{
			Result: result,
		}
		return &pbres, nil
	}
	return nil, err
}

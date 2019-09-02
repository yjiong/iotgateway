package devapi

import (
	"encoding/json"

	"golang.org/x/net/context"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"

	sjson "github.com/bitly/go-simplejson"
	log "github.com/sirupsen/logrus"
	pb "github.com/yjiong/iotgateway/api"
	"github.com/yjiong/iotgateway/internal/device"
	"github.com/yjiong/iotgateway/internal/gateway"
)

//ElectricMeterapi ..
type ElectricMeterapi struct {
	gw *gateway.Gateway
}

// NewElectricMeterapi creates a new ApplicationAPI.
func NewElectricMeterapi(gateway *gateway.Gateway) *ElectricMeterapi {
	return &ElectricMeterapi{
		gw: gateway,
	}
}

//ElectricMeterUpdate ....
func (p *ElectricMeterapi) ElectricMeterUpdate(ctx context.Context, req *pb.ElectricMeterUpdateRequest) (*pb.ElectricMeterUpdateResponse, error) {
	gateway.GrpcMsg = "req"
	defer func() {
		gateway.GrpcMsg = nil
	}()
	conn := map[string]interface{}{
		device.DevAddr: req.Devaddr,
		"commif":       req.Commif,
	}
	jsreq := map[string]interface{}{
		"data": map[string]interface{}{
			device.DevType: req.Devtype,
			device.DevID:   req.Devid,
			device.DevConn: conn,
			device.DevName: req.Dname,
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
		pbres := pb.ElectricMeterUpdateResponse{
			Result: result,
		}
		return &pbres, nil
	}
	return nil, err
}

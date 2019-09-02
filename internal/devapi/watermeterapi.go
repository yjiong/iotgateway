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

//WaterMeterapi ..
type WaterMeterapi struct {
	gw *gateway.Gateway
}

// NewWaterMeterapi creates a new ApplicationAPI.
func NewWaterMeterapi(gateway *gateway.Gateway) *WaterMeterapi {
	return &WaterMeterapi{
		gw: gateway,
	}
}

//WaterMeterUpdate ....
func (p *WaterMeterapi) WaterMeterUpdate(ctx context.Context, req *pb.WaterMeterUpdateRequest) (*pb.WaterMeterUpdateResponse, error) {
	gateway.GrpcMsg = "req"
	defer func() {
		gateway.GrpcMsg = nil
	}()
	conn := map[string]interface{}{
		device.DevAddr: req.Devaddr,
		"commif":       req.Commif,
		"fcode":        req.Fcode,
		device.DevName: req.Dname,
	}
	jsreq := map[string]interface{}{
		"data": map[string]interface{}{
			device.DevType: req.Devtype,
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
		pbres := pb.WaterMeterUpdateResponse{
			Result: result,
		}
		return &pbres, nil
	}
	return nil, err
}

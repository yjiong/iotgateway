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

//Dtl645_2007api ..
type Dtl645_2007api struct {
	gw *gateway.Gateway
}

// NewDtl645_2007api creates a new ApplicationAPI.
func NewDtl645_2007api(gateway *gateway.Gateway) *Dtl645_2007api {
	return &Dtl645_2007api{
		gw: gateway,
	}
}

//Dtl645_2007Update ....
func (p *Dtl645_2007api) Dtl645_2007Update(ctx context.Context, req *pb.Dtl645_2007UpdateRequest) (*pb.Dtl645_2007UpdateResponse, error) {
	gateway.GrpcMsg = "req"
	defer func() {
		gateway.GrpcMsg = nil
	}()
	conn := map[string]interface{}{
		device.DevAddr: req.Devaddr,
		"commif":       req.Commif,
		"BaudRate":     req.BaudRate,
		"DataBits":     req.DataBits,
		"Parity":       req.Parity,
		"StopBits":     req.StopBits,
		device.DevName: req.Dname,
	}
	jsreq := map[string]interface{}{
		"data": map[string]interface{}{
			//device.DevType: req.Devtype,
			device.DevType: "DLT645-2007",
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
		pbres := pb.Dtl645_2007UpdateResponse{
			Result: result,
		}
		return &pbres, nil
	}
	return nil, err
}

package endpoint

import (
	"github.com/Jundong-chan/seckill/config"
	"CommoditySpike/server/seckillcore/pkg/redis"
	"CommoditySpike/server/seckillcore/service"
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type SeckillServiceEndpoint struct {
	SeckillServiceEp endpoint.Endpoint
	HealthCheckEp    endpoint.Endpoint
}

type SeckillRequest struct {
	ProductId string `json:"productid"`
	UserId    string `json:"userid"`
	BuyNum    int    `json:"buynum"`
}

type SeckillResponse struct {
	ProductId string `json:"productid"`
	UserId    string `json:"userid"`
	Error     string `json:"error"`
}

type HealthCheckResponse struct {
	Status bool
}

type HealthCheckRequest struct {
}

func MakeSeckillServiceEp(svc service.SeckillServiceimpl) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(SeckillRequest)
		if !ok {
			return nil, errors.New("解析错误")
		}
		var sreq = config.SecRequest{
			ProductId:  req.ProductId,
			AccessTime: time.Now().Unix(),
			UserId:     req.UserId,
			BuyNum:     req.BuyNum,
		}
		//鉴权用户
		// umodel := model.NewUserModel()
		// err = umodel.CheckUserId(sreq.UserId)
		// if err != nil {
		// 	return nil, err
		// }
		//为了方便压测，将检测用户id的步骤关闭

		result, err := svc.SecKill(&sreq)
		if err != nil {
			return nil, err
		}
		res := result.(config.SecResult)
		switch res.Code {
		case config.Success:
			err = nil
			go redisclient.UpdateStorageToBase(res.ProductId) //更新缓存到数据库
			return SeckillResponse{
				ProductId: res.ProductId,
				UserId:    res.UserId,
			}, nil
		case config.NotSelling:
			err = errors.New("this product is not selling now")

		case config.ServiceBusyErr:
			err = errors.New("service busy")

		case config.Soldout:
			err = errors.New("product has been sold out")

		case config.Buylimit:
			err = errors.New("have getting buy limit")

		case config.ProccessTimeout:
			err = errors.New("ProccessTimeout")

		}
		if res.Code == config.NotSelling || res.Code == config.Soldout || res.Code == config.Buylimit {
			return SeckillResponse{
				ProductId: res.ProductId,
				UserId:    res.UserId,
				Error:     err.Error(),
			}, nil
		}
		return SeckillResponse{
			ProductId: res.ProductId,
			UserId:    res.UserId,
			Error:     err.Error(),
		}, err

	}
}

func MakeHealthCheckEp(svc service.CheckServiceimpl) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthCheckResponse{
			Status: status,
		}, nil
	}

}

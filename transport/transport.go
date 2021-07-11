package transport

import (
	"../endpoint"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	DecodeFailed = errors.New("Decode failed")
	EncodeFailed = errors.New("encode failed")
)

func decodeSeckillServiceRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.SeckillRequest
	var userid string
	//为了方便压测，并没有对token进行解析，只解析了userid
	cookie := r.Header.Get("Cookie")
	if cookie != "" {
		fmt.Println("解析到cookie：", cookie)
		cookie = strings.Replace(cookie, " ", "", -1)
		cookies := strings.Split(cookie, ";")
		if strings.Contains(cookies[0], "userid=") {
			userid = strings.TrimPrefix(cookies[0], " userid=")
		} else {
			userid = strings.TrimPrefix(cookies[1], " userid=")
		}
		fmt.Println("解析到userid：", userid)
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, DecodeFailed
	}
	if userid != "" {
		req.UserId = userid
	}
	return req, nil
}

func encodeSeckillServiceResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res, ok := response.(endpoint.SeckillResponse)
	if !ok {
		return EncodeFailed
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	return json.NewEncoder(w).Encode(res)
}

func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.HealthCheckRequest{}, nil
}

func encodeHealthCheckResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res, ok := response.(endpoint.HealthCheckResponse)
	if !ok {
		return errors.New("failed to encodecreateResponse")
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	return json.NewEncoder(w).Encode(res)
}

func MakeHttpHandler(ctx context.Context, enp endpoint.SeckillServiceEndpoint) http.Handler {
	r := mux.NewRouter()
	r.Methods("POST").Path("/buy").Handler(gokithttp.NewServer(
		enp.SeckillServiceEp,
		decodeSeckillServiceRequest,
		encodeSeckillServiceResponse,
	))
	r.Methods("POST", "GET").Path("/buy/health").Handler(gokithttp.NewServer(
		enp.HealthCheckEp,
		decodeHealthCheckRequest,
		encodeHealthCheckResponse,
	))

	return r
}

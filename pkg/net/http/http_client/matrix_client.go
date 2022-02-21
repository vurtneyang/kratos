package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"kratos/pkg/log"
	"kratos/pkg/net/metadata"
	xtime "kratos/pkg/time"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MatrixConfig struct {
	Host       string
	XProjectID string
	XAppID     string
	MasterKey  string
	ClientKey  string
	Timeout    xtime.Duration
}

type MatrixClient struct {
	config *MatrixConfig
	cli    *http.Client
}

const (
	clientType = "client"
	masterType = "master"
)

func NewMatrixClient(conf *MatrixConfig) *MatrixClient {
	return &MatrixClient{
		config: conf,
		cli: &http.Client{
			Transport: &http.Transport{
				IdleConnTimeout: time.Hour * 6,
			},
			Timeout: time.Duration(conf.Timeout),
		},
	}
}

type Args struct {
	Query string `json:"query" binding:"required"` // 查询语句
}

type Response struct {
	Code    string          `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Error   string          `json:"error"`
	Now     int64           `json:"now"`
}

// Post 适用于graphql方式请求调用
func (c *MatrixClient) Post(ctx context.Context, path, params string, reply interface{}) error {
	args := &Args{
		Query: params,
	}
	qlParams, err := json.Marshal(args)
	if err != nil {
		return errors.Wrap(err, "graphQl json Marshal args err")
	}
	return c.Request(ctx, "POST", clientType, path, string(qlParams), reply)
}

// ClientPost 适用于使用clientKey签名方式请求调用
func (c *MatrixClient) ClientPost(ctx context.Context, path string, params interface{}, reply interface{}) error {
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "ClientPost json Marshal params err")
	}
	return c.Request(ctx, "POST", clientType, path, string(paramsStr), reply)
}

// MasterPost 适用于使用masterKey签名方式请求调用
func (c *MatrixClient) MasterPost(ctx context.Context, path string, params interface{}, reply interface{}) error {
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "MasterPost json Marshal params err")
	}
	return c.Request(ctx, "POST", masterType, path, string(paramsStr), reply)
}

func (c *MatrixClient) Request(ctx context.Context, method, cType, path, params string, reply interface{}) error {
	bt := time.Now()

	body, response, err := c.do(ctx, method, cType, path, params)

	var msg string
	if response != nil && response.Code != "" {
		err = json.Unmarshal(response.Data, reply)
		msg = fmt.Sprintf("code:%s message:%s (%s)", response.Code, response.Message, response.Error)
	}

	lf := log.Infov
	var code = 0
	var errMsg string
	if err != nil {
		lf = log.Errorv
		code = -1
		errMsg = fmt.Sprintf("error:%s %s", err.Error(), msg)
	}
	duration := time.Since(bt)

	_metricClientReqDur.Observe(int64(duration/time.Millisecond), path)
	_metricClientReqCodeTotal.Inc(path, strconv.Itoa(code))

	lf(ctx,
		log.KVString("method", method),
		log.KVString("path", path),
		log.KVFloat64("ts", duration.Seconds()),
		log.KVString("source", "matrix-access-log"),
		log.KVString("error", errMsg),
		log.KVString("args", params),
		log.KVString("reply", string(body)),
	)

	return err
}

func (c *MatrixClient) do(ctx context.Context, method, cType, path, params string) (body []byte, response *Response, err error) {
	response = &Response{}

	request, err := c.buildRequest(ctx, method, cType, path, params)
	if err != nil {
		return nil, nil, errors.Wrap(err, "buildRequest err")
	}
	resp, err := c.cli.Do(request)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cli do err")
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "resp io ReadAll err")
	}

	err = json.Unmarshal(body, response)
	if err != nil {
		return body, nil, errors.Wrap(err, "json Unmarshal body err")
	}

	if response.Code != "OK" {
		err = errors.New(response.Code)
	}

	return
}

func (c *MatrixClient) buildRequest(ctx context.Context, method, cType, path, params string) (request *http.Request, err error) {
	url := fmt.Sprintf("%s/%s", c.config.Host, path)
	request, err = http.NewRequest(method, url, strings.NewReader(params))
	if err != nil {
		return nil, errors.Wrap(err, "buildRequest err")
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Project-ID", c.config.XProjectID)

	if cType == masterType {
		request.Header.Add("X-Sign", GetSign(c.config.MasterKey, time.Now().Unix()))

		userId := metadata.Int64(ctx, metadata.Mid)
		request.Header.Add("X-GID", strconv.FormatInt(userId, 10))
		request.Header.Add("X-APPID", c.config.XAppID)
	} else {
		request.Header.Add("X-Sign", GetSign(c.config.ClientKey, time.Now().Unix()))
	}

	return
}

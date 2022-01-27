package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"kratos/pkg/log"
	xtime "kratos/pkg/time"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MatrixConfig struct {
	Host           string
	WorkspaceToken string
	XProjectID     string
	Secret         string
	Timeout        xtime.Duration
}

type MatrixClient struct {
	config *MatrixConfig
	cli    *http.Client
}

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

func (c *MatrixClient) Get(ctx context.Context, path, params string, reply interface{}) error {
	return c.Request(ctx, "GET", path, params, reply)
}

func (c *MatrixClient) Post(ctx context.Context, path, params string, reply interface{}) error {
	return c.Request(ctx, "POST", path, params, reply)
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

func (c *MatrixClient) Request(ctx context.Context, method, path, params string, reply interface{}) error {
	bt := time.Now()

	body, response, err := c.do(method, path, params)

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

func (c *MatrixClient) do(method, path, params string) (body []byte, response *Response, err error) {
	response = &Response{}

	args := &Args{
		Query: params,
	}
	val, err := json.Marshal(args)
	if err != nil {
		return nil, nil, errors.Wrap(err, "json Marshal args err")
	}

	request, err := c.buildRequest(method, path, string(val))
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

func (c *MatrixClient) buildRequest(method, path, params string) (request *http.Request, err error) {
	url := fmt.Sprintf("%s/%s", c.config.Host, path)
	request, err = http.NewRequest(method, url, strings.NewReader(params))
	if err != nil {
		return nil, errors.Wrap(err, "buildRequest err")
	}
	request.Header.Add("X-Sign", GetSign(c.config.Secret, time.Now().Unix()))
	request.Header.Add("Workspace-Token", c.config.WorkspaceToken)
	request.Header.Add("X-Project-ID", c.config.XProjectID)
	request.Header.Add("Content-Type", "application/json")

	return
}

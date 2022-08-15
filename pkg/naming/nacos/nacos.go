package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	xtime "kratos/pkg/time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"kratos/pkg/log"
)

const (
	DefaultClusterName        = "DEFAULT"
	DefaultGroupName          = "DEFAULT_GROUP"
	DefaultNameSpaceID        = "public"
	DefaultGrpcPort           = "7883"
	DefaultHttpPort           = "3000"
	modeHeartBeat      string = "hb"
	modeSubscribe      string = "sb"
)

type Option interface {
	apply(opts *options)
}

type options struct {
	groupName   string
	clusters    string
	nameSpaceID string
	mode        string
	hbInterval  time.Duration
}

type ServerConf struct {
	Addr    string `json:"addr"`
	TimeOut string `json:"timeout"`
}

type NacosServerConf struct {
	IpAddr      string `json:"ipAddr"`
	IpAddr2     string `json:"ipAddr2"` //兼容rpcx注册集群不一致的问题
	Port        uint64 `json:"port"`
	NameSpaceId string `json:"nameSpaceId"`
	ProjectId   string `json:"projectId"`
	ClientKey   string `json:"clientKey"`
	MasterKey   string `json:"masterKey"`
	AppCode     string `json:"appCode"`
	ApiServer   string `json:"apiServer"`
}

type NacosClientConf struct {
	TimeOutMs            uint64 `json:"timeOutMs"`
	BeatInterval         int64  `json:"beatInterval"`
	NameSpaceId          string `json:"nameSpaceId"`
	CacheDir             string `json:"cacheDir"`
	LogDir               string `json:"logDir"`
	UpdateThreadNum      int    `json:"updateThreadNum"`
	NotLoadCacheAtStart  bool   `json:"notLoadCacheAtStart"`
	UpdateCacheWhenEmpty bool   `json:"updateCacheWhenEmpty"`
}

type NacosConf struct {
	Server      *ServerConf
	NacosServer *NacosServerConf
	NacosClient *NacosClientConf
	Timeout     xtime.Duration
}

type watcher struct {
	serviceName string
	clusters    []string
	groupName   string
	ctx         context.Context
	cancel      context.CancelFunc
	watchChan   chan bool
	cli         naming_client.INamingClient
}

// Registry is nacos registry.
type Registry struct {
	opts options
	cli  naming_client.INamingClient
}

type nacosKey string

var _nacosKey nacosKey = "kratos/pkg/naming/nacos/time"

// FromContext returns the trace bound to the context, if any.
func FromContext(ctx context.Context) (t time.Time, ok bool) {
	t, ok = ctx.Value(_nacosKey).(time.Time)
	return
}

// NewContext new a trace context.
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, _nacosKey, time.Now())
}

// 将服务注册到nacos中
func RegisterNacos() error {
	cluster, groupName, serverName := "", os.Getenv("APP_GROUP"), os.Getenv("APP_ID")
	if groupName == "" {
		groupName = DefaultGroupName
	}
	if cluster == "" {
		cluster = DefaultClusterName
	}

	client, err := NewNameClient()
	if err != nil {
		return err
	}
	// Register Instance
	ip, _ := externalIP()
	registerInfo := vo.RegisterInstanceParam{
		Ip:          ip.String(),
		Port:        getGrpcPort(),
		ServiceName: serverName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"gRPC": fmt.Sprintf("%d", getGrpcPort()), "HTTP": fmt.Sprintf("%d", getHttpPort())},
		ClusterName: cluster,   // default value is DEFAULT
		GroupName:   groupName, // default value is DEFAULT_GROUP
	}
	_, err = client.RegisterInstance(registerInfo)
	return err
}

// 将服务从nacos中注销
func DeregisterNacos() error {
	cluster, groupName, serverName := "", os.Getenv("APP_GROUP"), os.Getenv("APP_ID")
	if groupName == "" {
		groupName = DefaultGroupName
	}
	if cluster == "" {
		cluster = DefaultClusterName
	}
	client, err := NewNameClient()
	// unRegister Instance
	ip, _ := externalIP()
	unRegisterInfo := vo.DeregisterInstanceParam{
		Ip:          ip.String(),
		Port:        getGrpcPort(),
		ServiceName: serverName,
		Ephemeral:   true,
		Cluster:     cluster,   // default value is DEFAULT
		GroupName:   groupName, // default value is DEFAULT_GROUP
	}
	_, err = client.DeregisterInstance(unRegisterInfo)

	return err
}

// 获取服务的IP地址
func externalIP() (net.IP, bool) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, false
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, false
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, false
		}
	}
	return nil, true
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}

// 获取GRPC端口
func getGrpcPort() (port uint64) {
	str := os.Getenv("APP_RPC_PORT")
	if str == "" {
		str = DefaultGrpcPort
	}
	ports, _ := strconv.Atoi(str)
	return uint64(ports)
}

// 获取HTTP端口
func getHttpPort() (port uint64) {
	str := os.Getenv("APP_RPC_PORT")
	if str == "" {
		str = DefaultHttpPort
	}
	ports, _ := strconv.Atoi(str)
	return uint64(ports)
}

func NewNameClient() (c naming_client.INamingClient, err error) {
	// 为了便于迁移先写死
	NacosServer := os.Getenv("NACOS_SERVERS")
	if NacosServer == "" {
		panic("Get env:NACOS_SERVERS error")
	}
	serverConfig := make([]constant.ServerConfig, 0)
	ss := strings.Split(NacosServer, "")
	for _, v := range ss {
		c := strings.Split(v, ":")
		addr := c[0]
		port, _ := strconv.Atoi(c[1])
		serverConfig = append(serverConfig, constant.ServerConfig{
			IpAddr: addr,
			Port:   uint64(port),
		})
	}
	clientConfig := constant.ClientConfig{
		TimeoutMs:            10000, //# 10s 异常会摘除节点
		BeatInterval:         4000,  // 心跳间隔4s
		NamespaceId:          "public",
		UpdateThreadNum:      20,
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
	}
	c, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	)
	return c, err
}

func (w *watcher) Stop() error {
	w.cancel()
	//close
	return nil
}

func newWatcher(ctx context.Context, cli naming_client.INamingClient, serviceName string, groupName string, clusters []string) (*watcher, error) {
	w := &watcher{
		serviceName: serviceName,
		clusters:    clusters,
		groupName:   groupName,
		cli:         cli,
		watchChan:   make(chan bool, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)

	e := w.cli.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		Clusters:    clusters,
		GroupName:   groupName,
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			w.watchChan <- true
		},
	})
	return w, e
}

func Target(cluster, groupName, serviceName string, ops ...Option) string {
	if groupName == "" {
		groupName = DefaultGroupName
	}
	if cluster == "" {
		cluster = DefaultClusterName
	}

	// 变更注册方式
	NacosServer := os.Getenv("NACOS_SERVERS")
	if NacosServer == "" {
		panic("Get env:NACOS_SERVERS error")
	}
	addStr := "nacos://"
	ns := strings.Split(NacosServer, " ")
	for i, v := range ns {
		if i != 0 {
			addStr += ","
		}
		addStr = addStr + v
	}

	opts := &options{
		groupName:   groupName,
		clusters:    cluster,
		nameSpaceID: DefaultNameSpaceID,
		mode:        modeHeartBeat,
		hbInterval:  10 * time.Second,
	}
	for _, v := range ops {
		v.apply(opts)
	}

	str := fmt.Sprintf("%s?s=%s&n=%s&cs=%s&g=%s&m=%s&d=%d", addStr, serviceName, opts.nameSpaceID, opts.clusters, opts.groupName, opts.mode, opts.hbInterval/time.Millisecond)

	return str
}

type TracePlugin struct{}

func (p *TracePlugin) PreCall(ctx context.Context, servicePath, serviceMethod string, args interface{}) error {
	return nil
}

func (p *TracePlugin) PostCall(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, err error) error {
	var code, errMsg string
	lf := log.Infov
	duration := time.Second
	if err != nil {
		lf = log.Errorv
		errMsg = err.Error()
		code = "-1"
	}

	t, ok := FromContext(ctx)
	if ok {
		duration = time.Since(t)
		_metricClientReqDur.Observe(int64(duration/time.Millisecond), servicePath, serviceMethod)
		_metricClientReqCodeTotal.Inc(servicePath, serviceMethod, code)
	}

	var resp interface{}
	if jsonData, ok := reply.(*json.RawMessage); ok {
		jsonVal, _ := jsonData.MarshalJSON()
		resp = string(jsonVal)
	} else {
		resp = reply
	}

	lf(ctx,
		log.KVString("service", servicePath),
		log.KVString("path", serviceMethod),
		log.KVFloat64("ts", duration.Seconds()),
		log.KVString("source", "rpcx-access-log"),
		log.KVString("error", errMsg),
		log.KVString("args", fmt.Sprintf("%+v", args)),
		log.KVString("reply", fmt.Sprintf("%+v", resp)),
	)

	return nil
}

package nacos

import (
	"context"
	"fmt"
	xtime "kratos/pkg/time"
	"net"
	"strconv"
	"strings"
	"time"

	"kratos/pkg/conf/paladin"
	"kratos/pkg/log"
	bm "kratos/pkg/net/http/blademaster"
	"kratos/pkg/net/rpc/warden"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	nclient "github.com/rpcxio/rpcx-nacos/client"
	"github.com/smallnest/rpcx/client"
)

const (
	DefaultClusterName = "DEFAULT"
	DefaultGroupName   = "DEFAULT_GROUP"
	DefaultNameSpaceID = "public"

	modeHeartBeat string = "hb"
	modeSubscribe string = "sb"
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

// Dao dao
type RpcXDao struct {
	Conf   *RpcxConf      // 配置文件
	Client client.XClient // Rpcx连接池
}

type ServerConf struct {
	Addr    string `json:"addr"`
	TimeOut string `json:"timeout"`
}

type NacosServerConf struct {
	IpAddr      string `json:"ipAddr"`
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

type RpcxConf struct {
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

func NewRpcxDao(cluster, groupName, serverName string) (dao *RpcXDao, err error) {
	if groupName == "" {
		groupName = DefaultGroupName
	}
	if cluster == "" {
		cluster = DefaultClusterName
	}
	dao = &RpcXDao{}
	cfg := RpcxConf{}
	err = paladin.Get("nacos.toml").UnmarshalTOML(&cfg)
	if err != nil {
		log.Error("[Dao.New] UnmarshalToml err:%v", err)
		return dao, err
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = xtime.Duration(time.Millisecond * 500)
	}
	dao.Conf = &cfg
	// New RpcClients
	dao.Client, err = newClient(dao.Conf, cluster, groupName, serverName)
	if err != nil {
		log.Error("New Dao newClient err:%v", err)
		return dao, err
	}
	return dao, err
}

func newClient(conf *RpcxConf, cluster, groupName, serverName string) (c client.XClient, err error) {
	if groupName == "" {
		groupName = DefaultGroupName
	}
	if cluster == "" {
		cluster = DefaultClusterName
	}
	serverConfig := []constant.ServerConfig{{
		IpAddr: conf.NacosServer.IpAddr,
		Port:   conf.NacosServer.Port,
	}}

	clientConfig := constant.ClientConfig{
		TimeoutMs:            conf.NacosClient.TimeOutMs,
		BeatInterval:         conf.NacosClient.BeatInterval,
		NamespaceId:          conf.NacosClient.NameSpaceId,
		CacheDir:             conf.NacosClient.CacheDir,
		LogDir:               conf.NacosClient.LogDir,
		UpdateThreadNum:      conf.NacosClient.UpdateThreadNum,
		NotLoadCacheAtStart:  conf.NacosClient.NotLoadCacheAtStart,
		UpdateCacheWhenEmpty: conf.NacosClient.UpdateCacheWhenEmpty,
	}

	// cluster 按照业务方定义的名称
	//cluster := conf.NacosServer.NameSpaceId + "_" + serverName
	log.Info("serverConfig:%v", serverConfig)
	log.Info("clientConfig:%v", clientConfig)
	log.Info("cluster:%v", cluster)

	discovery, err := nclient.NewNacosDiscovery(serverName, cluster, groupName, clientConfig, serverConfig)
	if err != nil {
		log.Debug("Discovery err:%v", err)
		return nil, err
	}
	//
	c = client.NewXClient(serverName, client.Failover, client.RandomSelect, discovery, client.DefaultOption)
	c.Auth(conf.NacosServer.NameSpaceId)
	pc := c.GetPlugins()
	pc.Add(&TracePlugin{})
	c.SetPlugins(pc)
	return c, err
}

// 将服务注册到nacos中
func RegisterNacos(cluster, groupName, serverName string) error {
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
func DeregisterNacos(cluster, groupName, serverName string) error {
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
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	paladin.Get("grpc.toml").Unmarshal(&ct)
	ct.Get("Server").UnmarshalTOML(&cfg)
	dsnArr := strings.Split(cfg.Addr, ":")
	portS, _ := strconv.Atoi(dsnArr[1])
	port = uint64(portS)
	if port == 0 {
		port = 9000
	}
	return
}

// 获取HTTP端口
func getHttpPort() (port uint64) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	paladin.Get("http.toml").Unmarshal(&ct)
	ct.Get("Server").UnmarshalTOML(&cfg)
	dsnArr := strings.Split(cfg.Addr, ":")
	portS, _ := strconv.Atoi(dsnArr[1])
	port = uint64(portS)
	if port == 0 {
		port = 8000
	}
	return
}

func NewNameClient() (c naming_client.INamingClient, err error) {
	conf := RpcxConf{}
	err = paladin.Get("nacos.toml").UnmarshalTOML(&conf)
	if err != nil {
		log.Error("[Dao.New] UnmarshalToml err:%v", err)
		return c, err
	}

	serverConfig := []constant.ServerConfig{{
		IpAddr: conf.NacosServer.IpAddr,
		Port:   conf.NacosServer.Port,
	}}
	clientConfig := constant.ClientConfig{
		TimeoutMs:            conf.NacosClient.TimeOutMs,
		BeatInterval:         conf.NacosClient.BeatInterval,
		NamespaceId:          conf.NacosClient.NameSpaceId,
		CacheDir:             conf.NacosClient.CacheDir,
		LogDir:               conf.NacosClient.LogDir,
		UpdateThreadNum:      conf.NacosClient.UpdateThreadNum,
		NotLoadCacheAtStart:  conf.NacosClient.NotLoadCacheAtStart,
		UpdateCacheWhenEmpty: conf.NacosClient.UpdateCacheWhenEmpty,
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

func Target(nacosAddr string, cluster, groupName, serviceName string, ops ...Option) string {
	if groupName == "" {
		groupName = DefaultGroupName
	}
	if cluster == "" {
		cluster = DefaultClusterName
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
	tmp := nacosAddr
	if strings.HasPrefix(nacosAddr, "https://") {
		tmp = "nacoss://" + nacosAddr[8:]
	} else if strings.HasPrefix(nacosAddr, "http://") {
		tmp = "nacos://" + nacosAddr[7:]
	}
	str := fmt.Sprintf("%s?s=%s&n=%s&cs=%s&g=%s&m=%s&d=%d", tmp, serviceName, opts.nameSpaceID, opts.clusters, opts.groupName, opts.mode, opts.hbInterval/time.Millisecond)

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

	lf(ctx,
		log.KVString("service", servicePath),
		log.KVString("path", serviceMethod),
		log.KVFloat64("ts", duration.Seconds()),
		log.KVString("source", "rpcx-access-log"),
		log.KVString("error", errMsg),
		log.KVString("args", fmt.Sprintf("%+v", args)),
		log.KVString("reply", fmt.Sprintf("%+v", reply)),
	)

	return nil
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

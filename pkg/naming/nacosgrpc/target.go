package nacosgrpc

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultClusterName = "DEFAULT"
	DefaultGroupName   = "DEFAULT_GROUP"
	DefaultNameSpaceID = "public"
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

type op struct {
	f func(opts *options)
}

func (c *op) apply(opts *options) {
	c.f(opts)
}

func OptionGroupName(g string) Option {
	return &op{
		f: func(opts *options) {
			opts.groupName = g
		},
	}
}

func OptionNameSpaceID(g string) Option {
	return &op{
		f: func(opts *options) {
			opts.nameSpaceID = g
		},
	}
}

func OptionClusters(c []string) Option {
	return &op{
		f: func(opts *options) {
			if len(c) > 0 {
				opts.clusters = strings.Join(c, ",")
			}
		},
	}
}

func OptionModeHeartBeat(d time.Duration) Option {
	return &op{
		f: func(opts *options) {
			opts.mode = modeHeartBeat
			opts.hbInterval = d
		},
	}
}

func OptionModeSubscribe() Option {
	return &op{
		f: func(opts *options) {
			opts.mode = modeSubscribe
		},
	}
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

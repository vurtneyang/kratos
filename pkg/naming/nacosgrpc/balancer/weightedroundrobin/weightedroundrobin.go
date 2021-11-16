package weightedroundrobin

import (
	"errors"
	"math"
	"sync"

	"kratos/pkg/naming/nacosgrpc/rand"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const Name = "nacos_weighted_round_robin"

var ErrTotalWeightExceedLimit = errors.New("total weight exceed the max uint32")

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &wrrPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type attributeKey struct{}

var aKey attributeKey = attributeKey{}

func SetWeight(addr resolver.Address, weight uint32) resolver.Address {
	addr.Attributes = addr.Attributes.WithValues(aKey, weight)
	return addr
}

func GetWeight(addr resolver.Address) uint32 {
	v := addr.Attributes.Value(aKey)
	wt, _ := v.(uint32)
	return wt
}

type wrrPickerBuilder struct{}

type weightedConns struct {
	subConns []balancer.SubConn
	next     int
}

func (*wrrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := make(map[uint32]weightedConns)
	var total uint32
	for sc, v := range info.ReadySCs {
		w := GetWeight(v.Address)
		if ary, ok := scs[w]; ok {
			ary.subConns = append(ary.subConns, sc)
			scs[w] = ary
		} else {
			scs[w] = weightedConns{
				subConns: []balancer.SubConn{sc},
				next:     0,
			}
			if (math.MaxUint32 - int(total)) < int(w) {
				return base.NewErrPicker(ErrTotalWeightExceedLimit)
			}
			total += w
		}
	}
	for k, v := range scs {
		v.next = rand.Intn(len(v.subConns))
		scs[k] = v
	}
	return &wrrPicker{
		weightedConns: scs,
		totalWeight:   total,
	}
}

type wrrPicker struct {
	weightedConns map[uint32]weightedConns
	mu            sync.Mutex
	totalWeight   uint32
}

func (c *wrrPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	c.mu.Lock()
	var rnd uint32
	var l, r uint32
	var p weightedConns
	var k uint32
	if c.totalWeight > 0 {
		rnd = rand.Uint32n(c.totalWeight)
		for w, v := range c.weightedConns {
			r = l + w
			if rnd >= l && (rnd < r || (r == l && rnd == l)) {
				p = v
				k = w
				break
			}
			l += w
		}
	} else {
		p = c.weightedConns[0]
		k = 0
	}
	sc := p.subConns[p.next]
	p.next = (p.next + 1) % len(p.subConns)
	c.weightedConns[k] = p
	c.mu.Unlock()
	return balancer.PickResult{SubConn: sc}, nil
}

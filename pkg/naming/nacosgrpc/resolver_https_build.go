package nacosgrpc

import (
	"google.golang.org/grpc/resolver"
)

// func init() {
// 	resolver.Register(NewNacossBuilder())
// }

func NewNacossBuilder() resolver.Builder {
	return &nacossResolverBuilder{}
}

type nacossResolverBuilder struct{}

// // Build
func (*nacossResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r, err := newNacosResolver(target, cc)
	if err != nil {
		return nil, err
	}
	go r.start()
	return r, nil
}

func (*nacossResolverBuilder) Scheme() string {
	return "nacoss"
}

package nacosgrpc

import (
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(NewBuilder())
}

func NewBuilder() resolver.Builder {
	return &nacosResolverBuilder{}
}

type nacosResolverBuilder struct{}

// // Build
func (*nacosResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r, err := newNacosResolver(target, cc)
	if err != nil {
		return nil, err
	}
	go r.start()
	return r, nil
}

func (*nacosResolverBuilder) Scheme() string {
	return "nacos"
}

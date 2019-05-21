package resolver

import (
	"context"
	"net"
)

type dnsResolver struct {
}

// NewDNSResolver creates a resolver that uses DNS to find hosts
func NewDNSResolver() Resolver {
	return dnsResolver{}
}

func (r dnsResolver) Resolve(ctx context.Context, target string) ([]Result, error) {
	addrs, err := net.LookupHost(target)
	if err != nil {
		return nil, err
	}

	var results = make([]Result, 0, len(addrs))
	for _, addr := range addrs {
		results = append(results, Result{IP: addr, Label: createLabelPair("ip", addr)})
	}

	return results, nil
}

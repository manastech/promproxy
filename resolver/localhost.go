package resolver

import (
	"context"
	"os"
	"promproxy/util"
)

type localhostResolver struct {
}

// NewLocalhostResolver creates a resolver that always resolve to local host
func NewLocalhostResolver() Resolver {
	return localhostResolver{}
}

func (localhostResolver) Resolve(ctx context.Context, target string) ([]Result, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	label := util.CreateLabelPair("hostname", hostname)
	result := Result{IP: "127.0.0.1", Label: label}
	return []Result{result}, nil
}

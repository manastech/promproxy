package resolver

import (
	"context"
	"io/ioutil"
	"promproxy/util"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type dockerResolver struct {
	client *client.Client
}

var cpusetRegex = regexp.MustCompile("/docker/([0-9a-f]+)")

// NewDockerResolver creates a resolver that finds targets within docker containers
// running in the same machine
func NewDockerResolver(ctx context.Context) (Resolver, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	_, err = c.NetworkInspect(ctx, "promproxy")
	if err != nil {
		if client.IsErrNotFound(err) {
			_, err = c.NetworkCreate(ctx, "promproxy", types.NetworkCreate{})
			if err != nil {
				return nil, err
			}
		}
	}

	cpusetBytes, err := ioutil.ReadFile("/proc/self/cpuset")
	if err != nil {
		return nil, err
	}
	cpusetMatches := cpusetRegex.FindStringSubmatch(string(cpusetBytes))
	selfID := cpusetMatches[1]
	c.NetworkConnect(ctx, "promproxy", selfID, nil)

	return dockerResolver{client: c}, nil
}

func (r dockerResolver) Resolve(ctx context.Context, target string) ([]Result, error) {
	opts := types.ContainerListOptions{}
	containers, err := r.client.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	var results []Result

	for _, container := range containers {
		ip, err := r.matchContainer(ctx, target, container)
		if err != nil {
			return nil, err
		}

		if ip != "" {
			label := util.CreateLabelPair("container", container.Names[0])
			results = append(results, Result{IP: ip, Label: label})
		}
	}

	return results, nil
}

func (r dockerResolver) matchContainer(ctx context.Context, target string, container types.Container) (string, error) {
	composeHost := strings.Join([]string{
		container.Labels["com.docker.compose.service"],
		container.Labels["com.docker.compose.project"],
	}, ".")

	if composeHost == target {
		if net, ok := container.NetworkSettings.Networks["promproxy"]; ok {
			return net.IPAddress, nil
		}

		if err := r.client.NetworkConnect(ctx, "promproxy", container.ID, nil); err != nil {
			return "", err
		}

		container, err := r.client.ContainerInspect(ctx, container.ID)
		if err != nil {
			return "", err
		}

		return container.NetworkSettings.Networks["promproxy"].IPAddress, nil
	}

	rancherStackService := container.Labels["io.rancher.stack_service.name"]
	if rancherStackService != "" {
		stackAndService := strings.Split(rancherStackService, "/")
		rancherHost := strings.Join([]string{stackAndService[1], stackAndService[0]}, ".")
		if rancherHost == target {
			ipParts := strings.Split(container.Labels["io.rancher.container.ip"], "/")
			return ipParts[0], nil
		}
	}

	return "", nil
}

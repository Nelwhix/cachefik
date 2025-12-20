package docker

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func DiscoverServices() ([]Service, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	return listServices(context.Background(), cli)
}

func listServices(ctx context.Context, cli client.ContainerAPIClient) ([]Service, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	var services []Service

	for _, c := range containers {
		labels := c.Labels

		if labels["cachefik.enable"] != "true" {
			continue
		}

		rule := labels["cachefik.rule"]
		if rule == "" {
			continue
		}

		portStr := labels["cachefik.port"]
		if portStr == "" {
			continue
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		if c.NetworkSettings == nil || len(c.NetworkSettings.Networks) == 0 {
			continue
		}

		var ip string
		for _, n := range c.NetworkSettings.Networks {
			ip = n.IPAddress
			break
		}

		upstream := fmt.Sprintf("http://%s:%d", ip, port)
		services = append(services, Service{
			Rule:     rule,
			Upstream: upstream,
		})
	}

	sort.Slice(services, func(i, j int) bool {
		return len(services[i].PathPrefix()) > len(services[j].PathPrefix())
	})

	return services, nil
}

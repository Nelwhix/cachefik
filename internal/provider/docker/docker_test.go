package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockContainerAPIClient struct {
	mock.Mock
	client.ContainerAPIClient
}

func (m *mockContainerAPIClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func TestListServices(t *testing.T) {
	testCases := []struct {
		desc             string
		containers       []types.Container
		expectedServices []Service
	}{
		{
			desc: "simple case",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.rule":   "PathPrefix(`/test`)",
						"cachefik.port":   "8080",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "172.17.0.2",
							},
						},
					},
				},
			},
			expectedServices: []Service{
				{
					Rule:     "PathPrefix(`/test`)",
					Upstream: "http://172.17.0.2:8080",
				},
			},
		},
		{
			desc: "disabled container",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "false",
					},
				},
			},
			expectedServices: nil,
		},
		{
			desc: "missing rule",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.port":   "8080",
					},
				},
			},
			expectedServices: nil,
		},
		{
			desc: "missing port",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.rule":   "PathPrefix(`/test`)",
					},
				},
			},
			expectedServices: nil,
		},
		{
			desc: "invalid port",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.rule":   "PathPrefix(`/test`)",
						"cachefik.port":   "abc",
					},
				},
			},
			expectedServices: nil,
		},
		{
			desc: "multiple networks, picks first",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.rule":   "PathPrefix(`/multi`)",
						"cachefik.port":   "9000",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{
							"net1": {
								IPAddress: "192.168.1.1",
							},
							"net2": {
								IPAddress: "192.168.1.2",
							},
						},
					},
				},
			},
			expectedServices: []Service{
				{
					Rule:     "PathPrefix(`/multi`)",
					Upstream: "http://192.168.1.1:9000",
				},
			},
		},
		{
			desc: "no networks",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.rule":   "PathPrefix(`/test`)",
						"cachefik.port":   "8080",
					},
					NetworkSettings: &types.SummaryNetworkSettings{
						Networks: map[string]*network.EndpointSettings{},
					},
				},
			},
			expectedServices: nil,
		},
		{
			desc: "nil network settings",
			containers: []types.Container{
				{
					Labels: map[string]string{
						"cachefik.enable": "true",
						"cachefik.rule":   "PathPrefix(`/test`)",
						"cachefik.port":   "8080",
					},
					NetworkSettings: nil,
				},
			},
			expectedServices: nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			m := new(mockContainerAPIClient)
			m.On("ContainerList", mock.Anything, mock.Anything).Return(test.containers, nil)

			services, err := listServices(context.Background(), m)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedServices, services)
		})
	}
}

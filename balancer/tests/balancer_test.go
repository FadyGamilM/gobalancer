package tests

import (
	"testing"

	"github.com/FadyGamilM/gobalancer/balancer"
	"github.com/FadyGamilM/gobalancer/config"
	"github.com/FadyGamilM/gobalancer/transport"
	"github.com/stretchr/testify/require"
)

func TestBalancerRoundRobin(t *testing.T) {
	// set the variables we depend on
	// port := 5050

	router := transport.CreateRouter()

	balancerConfig := &config.BalancerConfig{
		Strategy: "RoundRobin",
		Engine:   router,
		Services: []config.Service{
			{
				Name: "demo-service",
				Replicas: []string{
					"http://localhost:8081",
					"http://localhost:8082",
				},
			},
		},
	}

	/*
		- in the config we defined 1-service with 2-replicas
		- so we are expecting to have 2 servers, one for each replica
		- we must ensure that the server.url is different
	*/
	gobalancer := balancer.NewBalancer(balancerConfig)

	// the number of created servers for service's replicas are correct
	require.Equalf(t, 2, len(gobalancer.ServerList.Servers), "number of created servers are >> %v \nnumber of replicas >> %v", len(gobalancer.ServerList.Servers), len(gobalancer.Configs.Services[0].Replicas))

	// the urls of the servers are correct as the replicas
	require.Equalf(t, "http://localhost:8081", gobalancer.ServerList.Servers[0].Url.String(), "the server url is >> %v \nservice replica url >> %v", gobalancer.ServerList.Servers[0].Url.String(), gobalancer.Configs.Services[0].Replicas[0])

	require.Equalf(t, "http://localhost:8082", gobalancer.ServerList.Servers[1].Url.String(), "the server url is >> %v \nservice replica url >> %v", gobalancer.ServerList.Servers[0].Url.String(), gobalancer.Configs.Services[0].Replicas[1])

}

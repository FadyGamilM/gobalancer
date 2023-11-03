package main

import (
	"flag"
	"fmt"

	"github.com/FadyGamilM/gobalancer/balancer"
	"github.com/FadyGamilM/gobalancer/config"
	"github.com/FadyGamilM/gobalancer/transport"
	log "github.com/sirupsen/logrus"
)

func main() {
	// define gin router
	router := transport.CreateRouter()

	// consumes the port from the flags at runtime
	port := flag.Int("server_port", 8080, "the port our server listening on")

	// parse the flags after all flags are defined and before they are used
	flag.Parse()

	// TODO => later we will create the config by loading the data from the yaml file, so we have to set the config.Engine manually after retriving it from this configYamlLoader function
	// define the go-balancer configs
	balancerConfig := &config.BalancerConfig{
		Engine: router,
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

	// define the go-balancer type
	gobalancer := balancer.NewBalancer(balancerConfig)

	// setup endpoints of the handler
	gobalancer.HandleRequests()

	// define http server and utilize the handler of the balancer (gin handler at the end)
	balancerServer := transport.CreateServer(fmt.Sprintf(":%d", *port), gobalancer)

	// init the server and handler unexpected errors
	if err := transport.InitServer(balancerServer); err != nil {
		log.Fatal(err)
	}
}

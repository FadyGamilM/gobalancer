package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FadyGamilM/gobalancer/balancer"
	"github.com/FadyGamilM/gobalancer/config"
	"github.com/FadyGamilM/gobalancer/pkg"
	"github.com/FadyGamilM/gobalancer/transport"
	log "github.com/sirupsen/logrus"
)

func main() {
	// define gin router
	router := transport.CreateRouter()

	// consumes the port from the flags at runtime
	port := flag.Int("server_port", 8080, "the port our server listening on")

	// consumes the yaml config file from the flags at runtime
	yamlConfigFile := flag.String("config_path", "./config/configs.yaml", "the yaml config file of the go-balancer")

	// parse the flags after all flags are defined and before they are used
	flag.Parse()

	// TODO => later we will create the config by loading the data from the yaml file, so we have to set the config.Engine manually after retriving it from this configYamlLoader function
	// read the go-balancer configs
	yamlFile, err := os.Open(*yamlConfigFile)
	if err != nil {
		log.Errorf("error opening the yaml config file >> %v", err)
	}
	_, err = pkg.LoadYamlConfigs(yamlFile)
	if err != nil {
		log.Errorf("error reading the yaml configs >> %v", err)
	}

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

	gobalancer.Configs.Engine.GET("/", gobalancer.BalanceLoad)

	// define http server and utilize the handler of the balancer (gin handler at the end)
	balancerServer := transport.CreateServer(fmt.Sprintf(":%d", *port), gobalancer)

	// init the server and handler unexpected errors
	if err := transport.InitServer(balancerServer); err != nil {
		log.Fatal(err)
	}
}

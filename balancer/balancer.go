package balancer

import (
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/FadyGamilM/gobalancer/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Balancer struct {
	Configs    *config.BalancerConfig
	ServerList *config.ServersList
}

func NewBalancer(c *config.BalancerConfig) *Balancer {
	// initialize list of servers with 0 servers
	servers := make([]*config.Server, 0)

	// the config consists of strategy, gin router, and services, we now care about services
	// => for each idx, service
	for _, service := range c.Services {
		// for now i will ingore the service name
		// => for each idx, replica_url
		for _, replica := range service.Replicas {
			// get the url in the right format
			replicaURL, err := url.Parse(replica)
			if err != nil {
				log.Printf("error trying to parse the replica url >> %v\n", err)
				os.Exit(1)
			}
			rProxy := httputil.NewSingleHostReverseProxy(replicaURL)
			// add new server
			servers = append(servers, &config.Server{
				Url:   *replicaURL,
				Proxy: rProxy,
			})
		}
	}
	return &Balancer{
		Configs: c,
		ServerList: &config.ServersList{
			Current: uint32(0),
			Servers: servers,
		},
	}
}

func (b *Balancer) BalanceLoad(c *gin.Context) {
	req := c.Request
	resp := c.Writer

	// ==> logging
	log.Info("receiving a new request to host [load balancer host] >> %s", req.Host)

	// 1. read request path  = host:port/service_a/rest_of_url
	// 2. load balance against the service_a and the url will be = host{i}:port{i}/rest_of_url
	// ==> so we have multiple servers (Hosts) host the same service (aka horizontial scalling)
	nextServer := b.ServerList.NextServer()
	log.Println("forwarding the request to the server number >> ", nextServer)
	// 3. forward the request to the proxy of the server
	b.ServerList.Servers[nextServer].Proxy.ServeHTTP(resp, req)
}

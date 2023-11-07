package balancer

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/FadyGamilM/gobalancer/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ServiceMatcherName string

type Balancer struct {
	Configs *config.BalancerConfig
	// the ServersPerService is a map where the key is the matcher of the service and the value is the serversList (the serversList is an array of servers and a current pointer pointing to the server to receive the request)
	ServersPerService map[ServiceMatcherName]*config.ServersList
}

func NewBalancer(c *config.BalancerConfig) *Balancer {
	// initialize the strategies supported by this laod-balancer
	config.InitBalancingStrategies()

	// initialize list of all servers of all services with 0 servers
	servers_perService := make(map[ServiceMatcherName]*config.ServersList, 0)

	// => for each idx, service
	for _, service := range c.Services {
		// initialize a list of servers for it (initally zero servers)
		servers := make([]*config.Server, 0)

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
		// now we setupped all the servers for all replicas (server per replica) of the current service, lets add the servers to this service via the matcher
		servers_perService[ServiceMatcherName(service.Matcher)] = &config.ServersList{
			Servers:           servers,
			ServiceName:       service.Name,
			BalancingStrategy: config.LoadStrategy(c.Strategy),
		}
	}
	// now setup the entire balancer structure
	return &Balancer{
		Configs:           c,
		ServersPerService: servers_perService,
	}
}

// for all registered services in our balancer
// check if the matcher of this service matches with the request path
// for the matched service -> return its ServersList
func (b *Balancer) findService(reqPath string) (*config.ServersList, error) {
	log.Infof("searching for the matcher service for the request >> :%v", reqPath)
	for matcher, serversList := range b.ServersPerService {
		if strings.HasPrefix(reqPath, string(matcher)) {
			log.Infof("found service %v which has matched the matcher ", serversList.ServiceName)
			return serversList, nil
		}
	}

	// if we reached here, so this service is not registered in our load balancer
	log.Errorf("couldn't find a service matching the request path")
	return nil, errors.New("couldn't found a matcher service for this given url")
}

func (b *Balancer) BalanceLoad(c *gin.Context) {
	req := c.Request
	resp := c.Writer

	// ==> logging
	log.Info("receiving a new request to host [load balancer host] >> %s", req.Host)

	serversList, err := b.findService(req.URL.Path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"response": "couldn't find a service",
		})
		return
	}

	// 1. read request path  = host:port/service_a/rest_of_url
	// 2. load balance against the service_a and the url will be = host{i}:port{i}/rest_of_url
	// ==> so we have multiple servers (Hosts) host the same service (aka horizontial scalling)
	nextServer, err := serversList.BalancingStrategy.NextServer(serversList.Servers)
	log.Println("forwarding the request to the server number >> ", nextServer)
	// 3. forward the request to the proxy of the server
	nextServer.Proxy.ServeHTTP(resp, req)
}

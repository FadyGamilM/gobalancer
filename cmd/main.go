package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// var (
// 	port = flag.Int("server_port", 8080, "the port our server listening on")
// )

type Service struct {
	Name     string
	Replicas []string
}

type BalancerConfig struct {
	*gin.Engine
	Services []Service
	Strategy string
}

type Balancer struct {
	Configs    *BalancerConfig
	ServerList *ServersList
}

func NewBalancer(c *BalancerConfig) *Balancer {
	// initialize list of servers with 0 servers
	servers := make([]*Server, 0)

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
			servers = append(servers, &Server{
				Url:   *replicaURL,
				Proxy: rProxy,
			})
		}
	}
	return &Balancer{
		Configs: c,
		ServerList: &ServersList{
			current: uint32(0),
			Servers: servers,
		},
	}
}

type Server struct {
	Url   url.URL
	Proxy *httputil.ReverseProxy
}

type ServersList struct {
	Servers []*Server
	// TODO => abstract this round robin later to a generic startegy
	// the current server the request will be forwareded to
	// -> i will use round robin for now
	current uint32
}

func (sl *ServersList) NextServer() uint32 {
	next := atomic.AddUint32(&sl.current, uint32(1))
	// lets say we have 3 servers
	// current goes from 0 -> 1 -> 2
	// now current is 3 so we should forward to the server number 0 (following the round robin)
	next = next % uint32(len(sl.Servers)) // handle wraparound using modulo operator
	return next
}

func CreateServer(port string, balancer *Balancer) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: balancer.Configs.Engine,
	}
}

func InitServer(srv *http.Server) error {
	err := srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func createRouter() *gin.Engine {
	return gin.Default()
}

func (b *Balancer) HandleRequests() {
	// the passed method is the same as serverHttp method
	b.Configs.Engine.GET("/", func(c *gin.Context) {
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
	})
}

func main() {
	// define gin router
	router := createRouter()

	// consumes the port from the flags at runtime
	port := flag.Int("server_port", 8080, "the port our server listening on")

	// parse the flags after all flags are defined and before they are used
	flag.Parse()

	// define the go-balancer configs
	balancerConfig := &BalancerConfig{
		Engine: router,
		Services: []Service{
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
	gobalancer := NewBalancer(balancerConfig)

	// setup endpoints of the handler
	gobalancer.HandleRequests()

	// define http server and utilize the handler of the balancer (gin handler at the end)
	balancerServer := CreateServer(fmt.Sprintf(":%d", *port), gobalancer)

	// init the server and handler unexpected errors
	if err := InitServer(balancerServer); err != nil {
		log.Fatal(err)
	}
}

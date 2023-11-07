package config

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// => aliases for strategies strings
var (
	ROUND_ROBIN_STRATEGY          = "roundrobin"
	WEIGHTED_ROUND_ROBIN_STRATEGY = "weighted_roundrobin"
	DEFAULT                       = "default"
)

type RoundRobin struct {
	// Each algorithms should be aware of who is the next server and keeps track of it
	CurrentServerIndex uint32
}

func NewRoundRobinStrategy() BalancingStrategy {
	return &RoundRobin{
		CurrentServerIndex: uint32(0),
	}
}

func (rr *RoundRobin) NextServer(servers []*Server) (*Server, error) {
	next := atomic.AddUint32(&rr.CurrentServerIndex, uint32(1))
	// lets say we have 3 servers
	// current goes from 0 -> 1 -> 2
	// now current is 3 so we should forward to the server number 0 (following the round robin)
	next = (next - 1) % uint32(len(servers)) // handle wraparound using modulo operator
	return servers[next], nil
}

// the map from strategy name to concrete-strategy-factory
// the map is to a factory func not to the BalancingStrategy object itself because i want a new instance for each request so we can start with a fresh currentServerIndex value for each request to our loadbalancer server
var StrategyFactories map[string]func() BalancingStrategy

func InitBalancingStrategies() {
	// initialize the map
	StrategyFactories = make(map[string]func() BalancingStrategy, 0)

	// set the round-robin strategy to its factory
	StrategyFactories[ROUND_ROBIN_STRATEGY] = NewRoundRobinStrategy

	// set the default strategy to round-robin also
	StrategyFactories[DEFAULT] = NewRoundRobinStrategy
}

// LoadStrategy is used when we create a new balancer instance
func LoadStrategy(strategyName string) BalancingStrategy {
	strategy, ok := StrategyFactories[strategyName]
	if !ok {
		log.Errorf("given strategy name >> %v is not implemented in the current version of load-balancer", strategyName)
		return StrategyFactories[DEFAULT]()
	}

	return strategy()
}

type Service struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
	// => each service should have some matcher so when the url path contains this matcher, we forward the request to this service
	Matcher string `yaml:"matcher"`
}

type BalancerConfig struct {
	*gin.Engine
	Services []Service `yaml:"services"`
	Strategy string    `yaml:"strategy"` // one of the above defined strategy names
}

type Server struct {
	Url   url.URL
	Proxy *httputil.ReverseProxy
}

// @ Description:
//   - BalancingStrategy is the abstraction for all load-balancing algorithms
//   - Each load-balancing algo must implements these methods to fullfill the strategy
//   - Next() method receives all servers of the current service and returns the server to receive the request and error if there is any
type BalancingStrategy interface {
	NextServer([]*Server) (*Server, error)
}

type ServersList struct {
	Servers []*Server
	// TODO => abstract this round robin later to a generic startegy
	// the current server the request will be forwareded to
	// -> i will use round robin for now
	// Current uint32 // ====> abstracted [DONE]
	BalancingStrategy BalancingStrategy

	ServiceName string
}

// func (sl *ServersList) NextServer() uint32 {
// 	next := atomic.AddUint32(&sl.Current, uint32(1))
// 	// lets say we have 3 servers
// 	// current goes from 0 -> 1 -> 2
// 	// now current is 3 so we should forward to the server number 0 (following the round robin)
// 	next = (next - 1) % uint32(len(sl.Servers)) // handle wraparound using modulo operator
// 	return next
// }

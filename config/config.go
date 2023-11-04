package config

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type Service struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
}

type BalancerConfig struct {
	*gin.Engine
	Services []Service `yaml:"services"`
	Strategy string    `yaml:"strategy"`
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
	Current uint32
}

func (sl *ServersList) NextServer() uint32 {
	next := atomic.AddUint32(&sl.Current, uint32(1))
	// lets say we have 3 servers
	// current goes from 0 -> 1 -> 2
	// now current is 3 so we should forward to the server number 0 (following the round robin)
	next = (next - 1) % uint32(len(sl.Servers)) // handle wraparound using modulo operator
	return next
}

package transport

import (
	"net/http"

	"github.com/FadyGamilM/gobalancer/balancer"
	"github.com/gin-gonic/gin"
)

func CreateServer(port string, balancer *balancer.Balancer) *http.Server {
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

func CreateRouter() *gin.Engine {
	return gin.Default()
}


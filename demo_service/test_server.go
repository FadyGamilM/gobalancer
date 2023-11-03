package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	// consumes the port from the flags at runtime
	port = flag.Int("server_port", 8081, "the port our server listening on")
)

type DemoRouter struct {
	*gin.Engine
}

func createDemoRouter() *DemoRouter {
	return &DemoRouter{
		Engine: gin.Default(),
	}
}

func createDemoServer(port string, r *DemoRouter) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: r.Engine,
	}
}

func initServer(srv *http.Server) error {
	log.Println("up and running on port >> ", srv.Addr)
	return srv.ListenAndServe()
}

func (r *DemoRouter) setupRouter() {
	r.Engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"response": c.Request.Host,
		})
	})
}

func main() {
	flag.Parse()
	demoRouter := createDemoRouter()

	demoRouter.setupRouter()

	demoServer := createDemoServer(
		fmt.Sprintf(":%d", *port),
		demoRouter,
	)
	if err := initServer(demoServer); err != nil {
		log.Fatal()
	}
}

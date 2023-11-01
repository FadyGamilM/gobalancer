package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateServer(port string, r *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    port,
		Handler: r,
	}
}

func InitServer(srv *http.Server) error {
	err := srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

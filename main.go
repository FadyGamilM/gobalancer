package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
)

var (
	port = flag.Int("server_port", 8080, "the port our server listening on")
)

func main() {
	// => flag.[] will returns the pointer not the value
	msg := flag.String(
		"msg",
		"default msg",
		"message to be printed",
	)

	// must be called after all flags be defined and before all flags being used
	flag.Parse()

	log.Println(*msg)

}

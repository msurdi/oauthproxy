package main

import (
	"flag"
	"github.com/msurdi/oauthproxy/server"
)

// Entry point
func main() {
	var configFile = flag.String("config", "oauthproxy.conf", "OAuthproxy configuration file")
	flag.Parse()
	server.LoadConfig(*configFile)
	server.Run()
}

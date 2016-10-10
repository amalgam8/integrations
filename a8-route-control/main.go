package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

// Icons can be found at http://fontawesome.io/icons/

func main() {
	serviceInstancesByContainerID = make(map[string]serviceInstance)
	go hostDockerQuery()
	go getAmalgam8ContainerIds()

	hostname, _ := os.Hostname()
	var (
		addr   = flag.String("addr", "/var/run/scope/plugins/a8routing.sock", "unix socket to listen for connections on")
		hostID = flag.String("hostname", hostname, "hostname of the host running this plugin")
	)
	flag.Parse()

	log.Printf("Starting on %s...\n", *hostID)

	os.Remove(*addr)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		os.Remove(*addr)
		os.Exit(0)
	}()

	listener, err := net.Listen("unix", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.Remove(*addr)
	}()

	log.Printf("Listening on: unix://%s", *addr)

	plugin := &Plugin{HostID: *hostID}
	http.HandleFunc("/report", plugin.Report)
	http.HandleFunc("/control", plugin.Control)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

// Plugin groups the methods a plugin needs
type Plugin struct {
	HostID string

	lock       sync.Mutex
	routesEnabled bool
}

type request struct {
	NodeID  string
	Control string
}

type response struct {
	ShortcutReport *report `json:"shortcutReport,omitempty"`
}

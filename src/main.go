package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/miekg/dns"
)

var (
	config     *Config
	blackList  *BlackList
	corpdomain *CorpDomain
	cache      Cache

	configFile = flag.String("c", "etc/config.yaml", "configuration file")
)

func startServer() {
	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", HandlerTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", HandlerUDP)

	tcpServer := &dns.Server{
		Addr:         "0.0.0.0:53",
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	udpServer := &dns.Server{
		Addr:         "0.0.0.0:53",
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		if err := tcpServer.ListenAndServe(); err != nil {
			log.Fatal("TCP-server start failed", err.Error())
		}
	}()
	go func() {
		if err := udpServer.ListenAndServe(); err != nil {
			log.Fatal("UDP-server start failed", err.Error())
		}
	}()
}

func listenInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	for range sig {
		log.Println("Terminating...")
	}
}

func main() {
	flag.Parse()

	var err error
	config, err = loadConfig(COLD)
	if err != nil {
		log.Fatal(err)
	}

	corpdomain = NewCorpDomain()
	corpdomain.Set(config.CorpDomain)
	corpdomain.SetExclude(config.ExcludeCorpDomain)

	cache = NewMemoryCache()

	blackList = UpdateList()
	go listUpdater()

	startServer()
	go runPrometheus()
	listenInterrupt()
}

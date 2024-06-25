package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	totalRequestsTCP = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "total",
		Help:      "total requests",

		ConstLabels: map[string]string{
			"type": "tcp",
		},
	}))

	totalRequestsUDP = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "total",
		Help:      "total requests",

		ConstLabels: map[string]string{
			"type": "udp",
		},
	}))

	totalRequestsFailed = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "failed",
		Help:      "failed requests",
	}))

	totalRequestsBlocked = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "blocked",
		Help:      "blocked requests",
	}))

	totalRequestsSuccess = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "success",
		Help:      "success requests",
	}))

	totalRequestsToPublicDNS = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "public",
		Help:      "public requests",
	}))

	totalRequestsToCorpDNS = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "corporate",
		Help:      "corporate requests",
	}))

	totalCacheHits = prometheus.NewCounter(prometheus.CounterOpts(prometheus.Opts{
		Namespace: "dns",
		Subsystem: "requests",
		Name:      "cache",
		Help:      "cached requests",
	}))
)

func runPrometheus() {
	prometheus.MustRegister(totalRequestsTCP)
	prometheus.MustRegister(totalRequestsUDP)
	prometheus.MustRegister(totalRequestsFailed)
	prometheus.MustRegister(totalRequestsBlocked)
	prometheus.MustRegister(totalRequestsSuccess)
	prometheus.MustRegister(totalRequestsToPublicDNS)
	prometheus.MustRegister(totalRequestsToCorpDNS)
	prometheus.MustRegister(totalCacheHits)

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":9970",
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           router,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

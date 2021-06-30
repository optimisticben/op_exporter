package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

//Define the metrics we wish to expose
var (
	gasPrice = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "op_gasPrice",
			Help: "Gas price."},
		[]string{"network", "layer"},
	)
)

func init() {
	//Register metrics with prometheus
	prometheus.MustRegister(gasPrice)
}

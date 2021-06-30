package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc"
	"gitlab.com/zcash/zcashd_exporter/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address on which to expose metrics and web interface.",
	).Default(":9100").String()
	rpcProvider = kingpin.Flag(
		"rpc.provider",
		"Address for RPC provider.",
	).Default("http://127.0.0.1:8545").String()
	networkLabel = kingpin.Flag(
		"label.network",
		"Label to apply to the metrics to identify the network.",
	).Default("mainnet").String()
	versionFlag = kingpin.Flag(
		"version",
		"Display binary version.",
	).Default("False").Bool()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	if *versionFlag {
		fmt.Printf("(version=%s, gitcommit=%s)\n", version.Version, version.GitCommit)
		fmt.Printf("(go=%s, user=%s, date=%s)\n", version.GoVersion, version.BuildUser, version.BuildDate)
		os.Exit(0)
	}
	log.Infoln("exporter config", *listenAddress, *rpcProvider, *networkLabel)

	log.Infoln("Starting op_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>OP Exporter</title></head>
		<body>
		<h1>OP Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		</body>
		</html>`))
	})
	go getRollupGasPrices()
	log.Infoln("Listening on", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}

}

func getRollupGasPrices() {
	rpcClient := jsonrpc.NewClientWithOpts(*rpcProvider, &jsonrpc.RPCClientOpts{})
	var rollupGasPrices *GetRollupGasPrices
	for {
		if err := rpcClient.CallFor(&rollupGasPrices, "rollup_gasPrices"); err != nil {
			log.Warnln("Error calling rollup_gasPrices", err)
		} else {
			l1GasPriceString := rollupGasPrices.L1GasPrice
			l1GasPrice, err := hexutil.DecodeUint64(l1GasPriceString)
			if err != nil {
				log.Warnln("Error converting gasPrice " + l1GasPriceString)
			}
			gasPrice.WithLabelValues(
				"kovan", "layer1").Set(float64(l1GasPrice))
			l2GasPriceString := rollupGasPrices.L2GasPrice
			l2GasPrice, err := hexutil.DecodeUint64(l2GasPriceString)
			if err != nil {
				log.Warnln("Error converting gasPrice " + l2GasPriceString)
			}
			gasPrice.WithLabelValues(
				"kovan", "layer2").Set(float64(l2GasPrice))
			log.Warnln("Got L1 gas prices: ", l1GasPrice)
			log.Warnln("Got L2 gas prices: ", l2GasPrice)

		}
		time.Sleep(time.Duration(30) * time.Second)
	}
}
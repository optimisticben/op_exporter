package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
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
	unhealthyTimePeriod = time.Minute * 10
)

type healthCheck struct {
	mu         *sync.RWMutex
	height     uint64
	healthy    bool
	updateTime time.Time
}

func healthHandler(health *healthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health.mu.RLock()
		defer health.mu.RUnlock()
		w.Write([]byte(fmt.Sprintf(`{ "healthy": "%t" }`, health.healthy)))
	}
}

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

	health := healthCheck{
		mu:         new(sync.RWMutex),
		height:     0,
		healthy:    false,
		updateTime: time.Now(),
	}
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", healthHandler(&health))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>OP Exporter</title></head>
		<body>
		<h1>OP Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		<p><a href="/health">Health</a></p>
		</body>
		</html>`))
	})
	go getRollupGasPrices()
	go getBlockNumber(&health)
	log.Infoln("Listening on", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}

}

func getBlockNumber(health *healthCheck) {
	rpcClient := jsonrpc.NewClientWithOpts(*rpcProvider, &jsonrpc.RPCClientOpts{})
	var blockNumberResponse *string
	for {
		if err := rpcClient.CallFor(&blockNumberResponse, "eth_blockNumber"); err != nil {
			log.Warnln("Error calling eth_blockNumber", err)
		} else {
			log.Infoln("Got block number: ", *blockNumberResponse)
			health.mu.Lock()
			currentHeight, err := hexutil.DecodeUint64(*blockNumberResponse)
			blockNumber.WithLabelValues(
				*networkLabel, "layer2").Set(float64(currentHeight))
			if err != nil {
				log.Warnln("Error decoding block height", err)
				continue
			}
			// TODO: handle error
			lastHeight := health.height
			// If the currentHeight is the same as the lastHeight
			if currentHeight == lastHeight {
				currentTime := time.Now()
				lastTime := health.updateTime
				if currentTime.Add(-unhealthyTimePeriod).Before(lastTime) {
					health.healthy = false
				}
			} else {
				health.height = currentHeight
				health.updateTime = time.Now()
				health.healthy = true
			}
			health.mu.Unlock()
		}
		time.Sleep(time.Duration(30) * time.Second)
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
				*networkLabel, "layer1").Set(float64(l1GasPrice))
			l2GasPriceString := rollupGasPrices.L2GasPrice
			l2GasPrice, err := hexutil.DecodeUint64(l2GasPriceString)
			if err != nil {
				log.Warnln("Error converting gasPrice " + l2GasPriceString)
			}
			gasPrice.WithLabelValues(
				*networkLabel, "layer2").Set(float64(l2GasPrice))
			log.Infoln("Got L1 gas string: ", l1GasPriceString)
			log.Infoln("Got L1 gas prices: ", l1GasPrice)
			log.Infoln("Got L2 gas string: ", l2GasPriceString)
			log.Infoln("Got L2 gas prices: ", l2GasPrice)

		}
		time.Sleep(time.Duration(30) * time.Second)
	}
}

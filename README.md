# op_exporter

A prometheus exporter to collect information from an Optimistic Ethereum node and serve metrics for collection

## Usage

```
go build && ./op_exporter --rpc.provider="https://kovan-sequencer.optimism.io" --label.network="kovan"
```

## Metrics

```
# HELP op_gasPrice Gas price.
# TYPE op_gasPrice gauge
op_gasPrice{layer="layer1",network="kovan"} 6.9e+09
op_gasPrice{layer="layer2",network="kovan"} 1
```

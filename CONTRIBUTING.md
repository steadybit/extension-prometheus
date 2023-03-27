# Contributing

## Running Locally

### Installing Prometheus Through Homebrew

#### Installation

```sh
brew install prometheus
brew services start prometheus
```

#### Configuration

Through `/opt/homebrew/etc`.

##### Scrape Configs for Local Platform & Agent

Add the following to the `prometheus.yml`.

```yaml
scrape_configs:
  - job_name: "prometheus"
    static_configs:
    - targets: ["localhost:9091"]

  - job_name: "agent"
    metrics_path: '/prometheus'
    static_configs:
      - targets: ["localhost:42899"]

  - job_name: "platform"
    metrics_path: '/actuator/prometheus'
    static_configs:
      - targets: ["localhost:9090"]
```

### Starting the Extension

```sh
export STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_0_NAME=local;
export STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_0_ORIGIN=http://127.0.0.1:9091
go run .
```

## References

 - [Collection of sample queries & alert rules](https://awesome-prometheus-alerts.grep.to/)

## Contributor License Agreement (CLA)

In order to accept your pull request, we need you to submit a CLA. You only need to do this once. If you are submitting a pull request for the first time, just submit a Pull Request and our CLA Bot will give you instructions on how to sign the CLA before merging your Pull Request.

All contributors must sign an [Individual Contributor License Agreement](https://github.com/steadybit/.github/blob/main/.github/cla/individual-cla.md).

If contributing on behalf of your company, your company must sign a [Corporate Contributor License Agreement](https://github.com/steadybit/.github/blob/main/.github/cla/corporate-cla.md). If so, please contact us via office@steadybit.com.

If for any reason, your first contribution is in a PR created by other contributor, please just add a comment to the PR
with the following text to agree our CLA: "I have read the CLA Document and I hereby sign the CLA".

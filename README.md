<img src="./logo.png" height="130" align="right" alt="Prometheus logo depicting a fire next to the text 'Prometheus'">

# Steadybit extension-prometheus

A [Steadybit](https://www.steadybit.com/) check implementation to gather Prometheus metrics within chaos engineering experiment executions. These can be used as checks within experiments, e.g., to implement pre- and post-conditions.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.github.steadybit.extension_prometheus).

## Configuration

| Environment Variable                                       | Helm value               | Meaning                                                                                          | Required |
|------------------------------------------------------------|--------------------------|--------------------------------------------------------------------------------------------------|----------|
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_NAME`         | `prometheus.name`        | Name of the Prometheus instance                                                                  | yes      |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_ORIGIN`       | `prometheus.origin`      | Url of the Prometheus                                                                            | yes      |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_HEADER_KEY`   | `prometheus.headerKey`   | Optional header key to send to the Prometheus API. Typically used for authentication purposes.   | no       |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_HEADER_VALUE` | `prometheus.headerValue` | Optional header value to send to the Prometheus API. Typically used for authentication purposes. | no       |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

We recommend that you deploy the extension with our [official Helm chart](https://github.com/steadybit/extension-prometheus/tree/main/charts/steadybit-extension-prometheus).

### Helm

```sh
helm repo add steadybit https://steadybit.github.io/extension-prometheus
helm repo update

helm upgrade steadybit-extension-prometheus \\
  --install \\
  --wait \\
  --timeout 5m0s \\
  --create-namespace \\
  --namespace steadybit-extension \\
  --set prometheus.name="dev" \\
  --set prometheus.origin="http://prometheus-server.default.svc.cluster.local" \\
  steadybit/steadybit-extension-prometheus
```

### Docker

You may alternatively start the Docker container manually.

```sh
docker run \\
  --env STEADYBIT_LOG_LEVEL=info \\
  --expose 8087 \\
  --env STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_0_NAME="{{SYMBOLIC_NAME}}" \\
  --env STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_0_ORIGIN="{{PROMETHEUS_INSTANCE_SERVER_ORIGIN}}" \\
  ghcr.io/steadybit/extension-prometheus:latest
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.

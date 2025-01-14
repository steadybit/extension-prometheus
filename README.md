<img src="./logo.png" height="130" align="right" alt="Prometheus logo depicting a fire next to the text 'Prometheus'">

# Steadybit extension-prometheus

A [Steadybit](https://www.steadybit.com/) check implementation to gather Prometheus metrics within chaos engineering experiment executions. These can be used as checks within experiments, e.g., to implement pre- and post-conditions.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_prometheus).

## Configuration

| Environment Variable                                         | Helm value                               | Meaning                                                                                                                | Required |
|--------------------------------------------------------------|------------------------------------------|------------------------------------------------------------------------------------------------------------------------|----------|
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_NAME`           | `prometheus.name`                        | Name of the Prometheus instance                                                                                        | yes      |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_ORIGIN`         | `prometheus.origin`                      | Url of the Prometheus                                                                                                  | yes      |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_HEADER_KEY`     | `prometheus.headerKey`                   | Optional header key to send to the Prometheus API. Typically used for authentication purposes.                         | no       |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_HEADER_VALUE`   | `prometheus.headerValue`                 | Optional header value to send to the Prometheus API. Typically used for authentication purposes.                       | no       |
| `STEADYBIT_EXTENSION_DISCOVERY_ATTRIBUTES_EXCLUDES_INSTANCE` | `discovery.attributes.excludes.instance` | List of Target Attributes which will be excluded during discovery. Checked by key equality and supporting trailing "*" | no       |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-prometheus.enabled=true \
--set extension-prometheus.prometheus.name="dev" \
--set extension-prometheus.prometheus.origin="http://prometheus-server.default.svc.cluster.local" \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-prometheus/blob/main/charts/steadybit-extension-prometheus/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-prometheus/blob/main/charts/steadybit-extension-prometheus).

```bash
helm repo add steadybit-extension-prometheus https://steadybit.github.io/extension-prometheus
helm repo update
helm upgrade steadybit-extension-prometheus \
  --install \
  --wait \
  --timeout 5m0s \
  --create-namespace \
  --namespace steadybit-agent \
  --set prometheus.name="dev" \
  --set prometheus.origin="http://prometheus-server.default.svc.cluster.local" \
  steadybit-extension-prometheus/steadybit-extension-prometheus
```

### Linux Package

Please use
our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts)
to install the extension on your Linux machine. The script will download the latest version of the extension and install
it using the package manager.

After installing, configure the extension by editing `/etc/steadybit/extension-prometheus` and then restart the service.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-registration) for more
information about extension registration and how to verify.

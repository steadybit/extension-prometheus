<img src="./logo.png" height="130" align="right" alt="Prometheus logo depicting a fire next to the text 'Prometheus'">

# Steadybit extension-prometheus

*Open Beta: This extension generally works, but you may discover some rough edges.*

A [Steadybit](https://www.steadybit.com/) check implementation to gather Prometheus metrics within chaos engineering experiment executions. These can be used as checks within experiments, e.g., to implement pre- and post-conditions.

## Configuration

| Environment Variable                                       |                                                                                                  |
|------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_NAME`         | Name of the Prometheus instance                                                                  |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_ORIGIN`       | Url of the Prometheus                                                                            |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_HEADER_KEY`   | Optional header key to send to the Prometheus API. Typically used for authentication purposes.   |
| `STEADYBIT_EXTENSION_PROMETHEUS_INSTANCE_<n>_HEADER_VALUE` | Optional header value to send to the Prometheus API. Typically used for authentication purposes. |


## Deployment

We recommend that you deploy the extension with our [official Helm chart](https://github.com/steadybit/extension-prometheus/tree/main/charts/steadybit-extension-prometheus).

## Agent Configuration

The Steadybit agent needs to be configured to interact with the Prometheus extension by adding the following environment variables:

```shell
# Make sure to adapt the URLs and indices in the environment variables names as necessary for your setup

STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL=http://steadybit-extension-prometheus.steadybit-extension.svc.cluster.local:8087
STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL=http://steadybit-extension-prometheus.steadybit-extension.svc.cluster.local:8087
```

When leveraging our official Helm charts, you can set the configuration through additional environment variables on the agent:

```
--set agent.env[0].name=STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL \
--set agent.env[0].value="http://steadybit-extension-prometheus.steadybit-extension.svc.cluster.local:8087" \
--set agent.env[1].name=STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL \
--set agent.env[1].value="http://steadybit-extension-prometheus.steadybit-extension.svc.cluster.local:8087"
```

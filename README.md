# Service Mesh Hub Enterprise Helm Charts

## Installation:
```shell script
helm repo add service-mesh-hub-enterprise https://storage.googleapis.com/service-mesh-hub-enterprise/service-mesh-hub-enterprise
helm repo update
helm install smh-e service-mesh-hub-enterprise/service-mesh-hub-enterprise -n service-mesh-hub --set service-mesh-hub-ui.license.key=<your license key>
```

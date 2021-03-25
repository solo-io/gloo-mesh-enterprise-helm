# Gloo Mesh Enterprise Helm Chart (Deprecated)

This repository has been deprecated. See the Gloo Mesh docs for Enterprise Helm documentation: https://docs.solo.io/gloo-mesh/latest/reference/helm/.

## Installation:
```shell script
helm repo add gloo-mesh-enterprise https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise
helm repo update
helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise -n gloo-mesh --set licenseKey=<your license key>
```

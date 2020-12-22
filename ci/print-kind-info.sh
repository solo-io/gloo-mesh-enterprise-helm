#!/bin/bash

set -x

# generate names: $1 allows to make several envs in parallel 
mgmtCluster=mgmt-cluster
remoteCluster=remote-cluster

for namespace in gloo-mesh istio-system; do
kubectl --context kind-$mgmtCluster get pod -n $namespace
kubectl --context kind-$remoteCluster get pod -n $namespace
kubectl --context kind-$mgmtCluster describe pod -n $namespace
kubectl --context kind-$remoteCluster describe pod -n $namespace
kubectl --context kind-$mgmtCluster get mesh -n $namespace
kubectl --context kind-$mgmtCluster get workloads -n $namespace
kubectl --context kind-$mgmtCluster get traffictargets -n $namespace
kubectl --context kind-$mgmtCluster get trafficpolicies -n $namespace -o yaml
kubectl --context kind-$mgmtCluster get accesspolicies -n $namespace -o yaml
kubectl --context kind-$mgmtCluster get virtualmesh -n $namespace -o yaml
done

kubectl --context kind-$mgmtCluster -n gloo-mesh logs deployment/discovery
kubectl --context kind-$mgmtCluster -n gloo-mesh logs deployment/networking
kubectl --context kind-$mgmtCluster -n gloo-mesh logs deployment/enterprise-extender
kubectl --context kind-$mgmtCluster -n gloo-mesh logs deployment/rbac-webhook

kubectl --context kind-$mgmtCluster -n gloo-mesh port-forward deployment/discovery 9091& sleep 2; echo INPUTS:; curl -v localhost:9091/snapshots/input; echo OUTPUTS:; curl -v localhost:9091/snapshots/input; killall kubectl

kubectl --context kind-$mgmtCluster -n gloo-mesh port-forward deployment/networking 9091& sleep 2; echo INPUTS:; curl -v localhost:9091/snapshots/input; echo OUTPUTS:; curl -v localhost:9091/snapshots/input; killall kubectl

# and process and disk info to debug out of disk space issues in CI
# this is too verbose: ps -auxf
df -h

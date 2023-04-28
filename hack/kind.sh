#!/usr/bin/env bash
set -o errexit

# desired cluster name; default is "kind"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"
IP_FAMILY="${IP_FAMILY:-ipv4}"
NODE_IMAGE="${NODE_IMAGE:-kindest/node:v1.24.0}"

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --image "${NODE_IMAGE}" --name "${KIND_CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: "${IP_FAMILY}"
nodes:
- role: control-plane
- role: worker
- role: worker
EOF


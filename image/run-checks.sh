#!/bin/sh
set -ex

shipyard
trap "shipyard -action stop" EXIT

export KUBECONFIG=$HOME/.shipyard/config

helm init

retry=30
expected="kube-system *tiller.* *1/1 *Running *"
while ! kubectl get pod --all-namespaces | grep -Pz "$expected"; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for tiller up. Aborting!"
        exit 1
    fi
    echo "Waiting for tiller up..."
    sleep 5
done

helm registry install quay.io/giantswarm/cert-operator-lab-chart -- \
       -n cert-operator-lab \
       --set imageTag=latest \
       --wait

retry=30
expected="certificate\.giantswarm\.io"
while ! kubectl get thirdpartyresources | grep -Pz "$expected"; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for TPR. Aborting!"
        exit 1
    fi
    echo "Waiting for TPR..."
    sleep 1
done

sleep 5

helm registry install quay.io/giantswarm/cert-resource-lab-chart -- \
       -n cert-resource-lab

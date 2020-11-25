#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

run_deploy_operator() {
  set -o nounset
  set -o errexit

  hypercloud::log::status "Deploy hypercloud-operator-go"
  local image=hypercloud-operator-go:v0.0.0

  make docker-build IMG=${image}
  kind load docker-image ${image}
  make deploy IMG=${image}


  kubectl get pod -n hypercloud-system



  set +o nounset
  set +o errexit


}


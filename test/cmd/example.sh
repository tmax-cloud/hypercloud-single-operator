#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


run_simple_tests() {
  set -o nounset
  set -o errexit

  hypercloud::log::status "Testing kubernetes cluster(v1:pods/v1:namespaces)"

  hypercloud::log::status "Create namespace test"
  kubectl create ns test
  hypercloud::test::get_object_assert namespace/test "{{$id_field}}" 'test'

  hypercloud::log::status "Create nginx pod"
  kubectl create -f - << __EOF__
{
  "kind": "Pod",
  "apiVersion": "v1",
  "metadata": {
    "name": "test",
    "namespace": "test"
  },
  "spec": {
    "containers": [
      {
        "name": "nginx",
        "image": "nginx"
      }
    ]
  }
}
__EOF__

  hypercloud::test::get_object_assert pod/test "{{$id_field}}" 'test' "-n test"

  hypercloud::log::status "Delete nginx pod"
  kubectl delete pod test -n test
  hypercloud::test::get_object_assert pods "{{range.items}}{{$id_field}}:{{end}}" '' "-n test"

  hypercloud::log::status "Delete namespace test"
  kubectl delete ns test

  set +o nounset
  set +o errexit

}


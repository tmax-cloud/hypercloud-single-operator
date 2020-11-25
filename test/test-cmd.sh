#!/usr/bin/env bash


set -o errexit
set -o nounset
set -o pipefail

HYPERCLOUD_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/..
source "${HYPERCLOUD_ROOT}/test/lib/logging.sh"
source "${HYPERCLOUD_ROOT}/test/lib/test.sh"
source "${HYPERCLOUD_ROOT}/test/lib/sh2ju.sh"

source "${HYPERCLOUD_ROOT}/test/cmd/example.sh"
source "${HYPERCLOUD_ROOT}/test/cmd/deploy.sh"


function record_command() {
    set +o nounset
    set +o errexit

    local name="$1"
    local output="${KUBE_JUNIT_REPORT_DIR:-/tmp/junit-results}"
    echo "Recording: ${name}"
    echo "Running command: $*"
    juLog -output="${output}" -class="test-cmd" -name="${name}" "$@"
    local exitCode=$?
    if [[ ${exitCode} -ne 0 ]]; then
      # Record failures for any non-canary commands
      if [ "${name}" != "record_command_canary" ]; then
        echo "Error when running ${name}"
        foundError="${foundError}""${name}"", "
      fi
    elif [ "${name}" == "record_command_canary" ]; then
      # If the canary command passed, fail
      echo "record_command_canary succeeded unexpectedly"
      foundError="${foundError}""${name}"", "
    fi

    set -o nounset
    set -o errexit
}


runTests() {

  foundError=""

  hypercloud::log::status "Checking kubectl version"
  kubectl version

  export id_field=".metadata.name"
  export labels_field=".metadata.labels"
  export annotations_field=".metadata.annotations"
  export service_selector_field=".spec.selector"
  export rc_replicas_field=".spec.replicas"
  export rc_status_replicas_field=".status.replicas"
  export rc_container_image_field=".spec.template.spec.containers"
  export rs_replicas_field=".spec.replicas"
  export port_field="(index .spec.ports 0).port"
  export port_name="(index .spec.ports 0).name"
  export second_port_field="(index .spec.ports 1).port"
  export second_port_name="(index .spec.ports 1).name"
  export image_field="(index .spec.containers 0).image"
  export pod_container_name_field="(index .spec.containers 0).name"
  export container_name_field="(index .spec.template.spec.containers 0).name"
  export hpa_min_field=".spec.minReplicas"
  export hpa_max_field=".spec.maxReplicas"
  export hpa_cpu_field=".spec.targetCPUUtilizationPercentage"
  export template_labels=".spec.template.metadata.labels.name"
  export statefulset_replicas_field=".spec.replicas"
  export statefulset_observed_generation=".status.observedGeneration"
  export job_parallelism_field=".spec.parallelism"
  export deployment_replicas=".spec.replicas"
  export secret_data=".data"
  export secret_type=".type"
  export change_cause_annotation='.*kubernetes.io/change-cause.*'
  export pdb_min_available=".spec.minAvailable"
  export pdb_max_unavailable=".spec.maxUnavailable"
  export generation_field=".metadata.generation"
  export container_len="(len .spec.template.spec.containers)"
  export image_field0="(index .spec.template.spec.containers 0).image"
  export image_field1="(index .spec.template.spec.containers 1).image"

  record_command run_simple_tests
  record_command run_deploy_operator

  if [[ -n "${foundError}" ]]; then
    echo "FAILED TESTS: ""${foundError}"
    exit 1
  fi

  hypercloud::log::status "Test done"

}

runTests


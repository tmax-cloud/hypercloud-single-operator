#!/usr/bin/env bash

# A set of helpers for tests
readonly reset=$(tput sgr0)
readonly  bold=$(tput bold)
readonly black=$(tput setaf 0)
readonly   red=$(tput setaf 1)
readonly green=$(tput setaf 2)

# Prints the calling file and line number $1 levels deep
# Defaults to 2 levels so you can call this to find your own caller
hypercloud::test::get_caller() {
  local levels=${1:-2}
  local caller_file="${BASH_SOURCE[${levels}]}"
  local caller_line="${BASH_LINENO[${levels}-1]}"
  echo "$(basename "${caller_file}"):${caller_line}"
}

# Force exact match of a returned result for a object query.  Wrap this with || to support multiple
# valid return types.
# This runs `kubectl get` once and asserts that the result is as expected.
# $1: Object on which get should be run
# $2: The go-template to run on the result
# $3: The expected output
# $4: Additional args to be passed to kubectl
hypercloud::test::get_object_assert() {
  hypercloud::test::object_assert 1 "$@"
}

# Asserts that the output of a given get query is as expected.
# Can run the query multiple times before failing it.
# $1: Number of times the query should be run before failing it.
# $2: Object on which get should be run
# $3: The go-template to run on the result
# $4: The expected output
# $5: Additional args to be passed to kubectl
hypercloud::test::object_assert() {
  local tries=$1
  local object=$2
  local request=$3
  local expected=$4
  local args=${5:-}

  for j in $(seq 1 "${tries}"); do
    # shellcheck disable=SC2086
    # Disabling because "args" needs to allow for expansion here
    res=$(eval kubectl get "${kube_flags[@]}" ${args} "${object}" -o go-template=\""${request}"\")
    if [[ "${res}" =~ ^$expected$ ]]; then
        echo -n "${green}"
        echo "$(hypercloud::test::get_caller 3): Successful get ${object} ${request}: ${res}"
        echo -n "${reset}"
        return 0
    fi
    echo "Waiting for Get ${object} ${request} ${args}: expected: ${expected}, got: ${res}"
    sleep $((j-1))
  done

  echo "${bold}${red}"
  echo "$(hypercloud::test::get_caller 3): FAIL!"
  echo "Get ${object} ${request}"
  echo "  Expected: ${expected}"
  echo "  Got:      ${res}"
  echo "${reset}${red}"
  caller
  echo "${reset}"
  return 1
}

hypercloud::test::get_object_jsonpath_assert() {
  local object=$1
  local request=$2
  local expected=$3
  local args=${4:-}

  res=$(eval kubectl get "${kube_flags[@]}" ${args} "${object}" -o jsonpath=\""${request}"\")

  if [[ "${res}" =~ ^$expected$ ]]; then
      echo -n "${green}"
      echo "$(hypercloud::test::get_caller): Successful get ${object} ${request}: ${res}"
      echo -n "${reset}"
      return 0
  else
      echo "${bold}${red}"
      echo "$(hypercloud::test::get_caller): FAIL!"
      echo "Get ${object} ${request}"
      echo "  Expected: ${expected}"
      echo "  Got:      ${res}"
      echo "${reset}${red}"
      caller
      echo "${reset}"
      return 1
  fi
}

#hypercloud::test::object_assert 1 pods "{{range.items}}{{.metadata.name}}:{{end}}" 'test'
#hypercloud::test::object_assert 1 pod/dind "{{.metadata.name}}" 'dind' "-n kind"
#hypercloud::test::get_object_jsonpath_assert pod/dind "{.metadata.name}" 'dind' "-n kind"


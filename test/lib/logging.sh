#!/usr/bin/env bash


VERBOSE="${VERBOSE:-2}"

hypercloud::log::errexit() {
  echo "errexit"
  return
}

hypercloud::log::info() {
  local V="${V:-0}"
  if [[ ${VERBOSE} < ${V} ]]; then
    return
  fi

  for message; do
    echo "${message}"
  done
}

hypercloud::log::status() {
  local V="${V:-0}"
  if [[ ${VERBOSE} < ${V} ]]; then
    return
  fi

  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "+++ ${timestamp} ${1}"
  shift
  for message; do
    echo "    ${message}"
  done
}


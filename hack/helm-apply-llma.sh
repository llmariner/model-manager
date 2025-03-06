#!/usr/bin/env bash

set -eo pipefail

base=$(readlink -f $(dirname $0))

LLMA_PATH=${1:?LLMariner Path}
EXTRA_VALS=$2
KUBE_CONTEXT=${3:-$(kubectl config current-context)}
HELM_ENV=$4

extra_flags=()
if [ -n "$EXTRA_VALS" ]; then
  IFS=',' read -r -a APPS <<< "$EXTRA_VALS"
  for val in "${APPS[@]}"; do
    extra_flags+=("--values $base/$val")
  done
fi
if [ ! -z "$HELM_ENV" ]; then
  extra_flags+=("--environment $HELM_ENV")
fi

echo "base: $base"

cd ${LLMA_PATH}/provision/dev
helmfile apply \
         --skip-refresh \
         --skip-diff-on-install \
         --kube-context ${KUBE_CONTEXT} \
         --selector app=llmariner \
         --values $base/values.yaml ${extra_flags[@]}

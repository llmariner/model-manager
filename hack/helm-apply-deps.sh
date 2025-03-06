#!/usr/bin/env bash

set -eo pipefail

LLMA_PATH=${1:?LLMariner Path}
DEP_APPS=$2
KUBE_CONTEXT=${3:-$(kubectl config current-context)}
HELM_ENV=${4:-""}

extra_flags=()
if [ -n "$DEP_APPS" ]; then
  IFS=',' read -r -a APPS <<< "$DEP_APPS"
  for app in "${APPS[@]}"; do
    extra_flags+=("-l app=$app")
  done
fi

cd ${LLMA_PATH}/provision/dev
helmfile apply \
         --skip-diff-on-install \
         --kube-context ${KUBE_CONTEXT} \
         --environment ${HELM_ENV} \
         -f helmfile.yaml ${extra_flags[@]}

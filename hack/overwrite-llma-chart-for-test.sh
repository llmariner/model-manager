#!/usr/bin/env bash

set -eo pipefail

LLMA_PATH=${1:?LLMariner Path}

pip install pyyaml
python $(dirname $0)/overwrite-llma-chart-for-test.py \
       ${LLMA_PATH}/deployments/llmariner/Chart.yaml \
       $(realpath --relative-to=${LLMA_PATH}/deployments/llmariner deployments)

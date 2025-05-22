#!/bin/bash
set -euo pipefail

# Define a function to log messages with timestamp
log() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S')] $1"
}

export CLUSTER_CTX=kind-kagent

CLUSTER_CTX=kind-kagent
# Loop through each challenge defined in the .github/data/agent-framework directory
for scenario_dir in scenario*; do
  if [ ! -d "$scenario_dir" ]; then
    continue
  fi
  pushd $scenario_dir
  pnpm i || npm i
  echo "pwd=$(pwd)"
  for challenge_file in *.yaml; do
      # reset environment
      bash "./run.sh"
      ../run-challenge.sh "$scenario_dir" "$challenge_file"
  done
  popd
done
#!/bin/bash

scenario_dir="$1"
challenge_file="$2"

# Extract the challenge name and description from the YAML metadata file
NAME=$(yq eval '.metadata.name' "$challenge_file")
DESCRIPTION=$(yq eval '.spec.description' "$challenge_file")
USER_PROMPT=$(yq eval '.spec.prompt' "$challenge_file")

log() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S')] $1"
}

# Run the challenge scenario using a Bash script generated from markdown in the README file
log "*********************************************************************"
log "Running challenge: $NAME - $DESCRIPTION"
log "*********************************************************************"
log "User Prompt: $USER_PROMPT"


echo "Waiting for pods to be stable..."
while kubectl --context ${CLUSTER_CTX} get pods -A | grep ContainerCreating; do sleep 5; done
while kubectl --context ${CLUSTER_CTX} get pods -A | grep Terminating; do sleep 5; done

# Test baseline
timeout --signal=INT 3m mocha ./test.js --timeout 10000 --retries 5

# Break the environment by executing commands defined in each step of the challenge
log "Breaking the environment..."
STEPS_COUNT=$(yq '.spec.steps | length' "$challenge_file")
for ((i=0; i<$STEPS_COUNT; i++)); do
    yq ".spec.steps[$i].run" "$challenge_file" | while IFS= read -r cmd; do
    echo "$cmd" >> "$challenge_file".$i.sh
    done
    echo "Waiting for pods to be stable..."
while kubectl --context ${CLUSTER_CTX} get pods -A | grep ContainerCreating; do sleep 5; done
while kubectl --context ${CLUSTER_CTX} get pods -A | grep Terminating; do sleep 5; done

####TODO    sh "$challenge_file".$i.sh
    sh "$challenge_file".$i.sh
done
rm -f "$challenge_file".*.sh
echo "Waiting for pods to be stable..."
# while kubectl --context ${CLUSTER_CTX} get pods -A | grep ContainerCreating; do sleep 5; done
while kubectl --context ${CLUSTER_CTX} get pods -A | grep Terminating; do sleep 5; done
kubectl --context ${CLUSTER_CTX} get pods -A

log "Testing cluster after breaking..."
timeout --signal=INT 1m mocha ./test.js --timeout 10000 || true

# Try to fix the broken environment using the Agent Framework (apps/agent-framework) and OpenAI API

log "Trying to fix thekagent broken environment using the Agent Framework..."

# Pipe the output of kagent invoke to the thought log file
touch $NAME.thought.log
####TODO echo "$USER_PROMPT" | kagent invoke --agent "k8s-agent" --task - > $NAME.thought.log 2>&1
timeout --signal=INT 3m bash -c 'echo "$1" | kagent invoke --agent "k8s-agent" --task -' -- "$USER_PROMPT" > $NAME.thought.log 2>&1

log "Testing cluster after fixing..."
kubectl --context ${CLUSTER_CTX} get pods -A
if mocha ./test.js --timeout 10000; then
  log "---------------> challenge SUCCESSFUL <------------------"
  rm -f $NAME.failure
  cat $NAME.thought.log > results/$NAME.success
else
  log "---------------> challenge FAILED <----------------------"
  rm -f $NAME.success
  cat $NAME.thought.log > results/$NAME.failure
fi


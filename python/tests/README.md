## Artifacts

The test data is split into two groups - test cases and agent configurations. 

### Test cases

The test cases file consists of one or more test cases, each having an input prompt (this is the query that's sent directly to the agent) and the expected output, which is the response that the agent should return.

```yaml
version: "1.0"
metadata:
  description: "Authorization Policy Test Cases"

test_cases:
  - name: deny_post_8080
    input: "Deny results with POST method on port 8080 on all workloads in the foo namespace"
    category: AuthorizationPolicy
    expected_output:
      apiVersion: security.istio.io/v1
      kind: AuthorizationPolicy
      metadata:
        name: httpbin
        namespace: foo
      spec:
        action: DENY
        rules:
          - to:
              - operation:
                  methods:
                    - POST
                  ports:
                    - "8080"
```

### Agent configuration

The agent configuration specifies the system prompt for the agent. This is the context that the agent uses to generate the response. 

```yaml
version: "1.0"
name: "istio_authpolicy_crd_agent"
metadata:
  description: "Agent for generating Istio Authorization Policy CRDs"
  version: "0.0.1"

system_messages:
  - |
    You're an Istio CRD agent. You modify or create a new JSON based on the UQ. The JSON must conform to the PROTO SPEC. The response must only include one or more AuthorizationPolicy resource type.

```
## Running tests

To run the tests you pass in the test case, the agent file and specify the model you want to use:

```bash
uv run main.py run test_cases/authpolicy_test_cases.yaml agents/istio_crd_agent_0.yaml --model gpt-4o-mini
```

Once you've created the baseline results (or first test results), you can modify the prompt or model and run the tests again to see if the results change.

To check for changes in the results, you can use the `compare` command:

```bash
uv run main.py compare test_results/results1.json test_results/results2.json
```

The `compare` command will output the differences between the two test results files and provide a summary of the changes between the two runs:

```console
Comparing results_20250122_153628.json with results_20250122_153642.json
Model changed: True
Prompt changed: False

Analyzing 2 test cases:

=== Test 1 ===
Input: Deny results with POST method on port 8080 on all workloads in the foo namespace
Duration delta: 6569.08ms
Similarity: 95.46% → 95.46% (Δ: +0.00%)
No differences in output

=== Test 2 ===
Input: Allow GET requests on port 3000 for service-a in the bar namespace
Duration delta: 3134.77ms
Similarity: 87.52% → 83.91% (Δ: -3.62%)
Output differences:
--- results_20250122_153628.json (Test 2)
+++ results_20250122_153642.json (Test 2)
@@ -2,16 +2,16 @@
   "apiVersion": "security.istio.io/v1",
   "kind": "AuthorizationPolicy",
   "metadata": {
-    "name": "allow-get-service-a",
+    "name": "allow-get-port-3000",
     "namespace": "bar"
   },
   "spec": {
+    "action": "ALLOW",
     "selector": {
-      "matchLabels": {
+      "match_labels": {
         "app": "service-a"
       }
     },
-    "action": "ALLOW",
     "rules": [
       {
         "to": [

=== Summary Statistics ===
Tests with differences: 1 of 2
Average similarity delta: -1.81%
Average duration delta: +4851.93ms
```

## Writing agent prompts

I found the following structure to work best: 

```console
# Role
<explain the role of the agent>

# Context
<give the context, from my testing the CRD YAML spec works best (I tried Istio docs and protos as well)>

# Examples
<list of input/output pairs; without examples, the rate dropped for almost 40%>
```

I also got better results with gpt-4o-mini than gpt-4o.

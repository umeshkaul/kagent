import asyncio

from autogen_agentchat.agents import AssistantAgent, UserProxyAgent
from autogen_agentchat.conditions import TextMentionTermination
from autogen_agentchat.teams import RoundRobinGroupChat
from autogen_agentchat.ui import Console
from autogen_ext.models.openai import OpenAIChatCompletionClient

from kagent.tools.argo import (
    # Resource Generation
    ArgoCRDTool,
    ArgoCRDToolConfig,
    # Installation & Verification
    VerifyArgoRolloutsControllerInstall,
)
from kagent.tools.k8s import (
    ApplyManifest,
    CreateResource,
    DeleteResource,
    DescribeResource,
    GetResources,
    GetResourceYAML,
    PatchResource,
)

# Model configuration
model_client = OpenAIChatCompletionClient(
    model="gpt-4o",
)

# Define tool groups for better organization
k8s_tools = [CreateResource(), ApplyManifest(), PatchResource(), DeleteResource()]

argo_management_tools = [
    GetResources(),
    GetResourceYAML(),
    DescribeResource(),
]

argo_setup_tools = [
    VerifyArgoRolloutsControllerInstall(),
    ArgoCRDTool(ArgoCRDToolConfig(model="gpt-4o-mini", openai_api_key=None)),
]

ARGO_SYSTEM_MESSAGE = """You are an Argo Rollouts specialist focused on progressive delivery and deployment automation. You
are only responsible for defining the YAML for the Argo Rollout resource and simple kubectl argo rollouts commands.

Your key responsibility is assisting users with migrating their Kubernetes deployments to Argo Rollouts:
   - Convert Kubernetes deployments to Argo Rollout resources
   - Define the Argo Rollout resource YAML

There are ways to migrate to Rollout:
- Convert an existing Deployment resource to a Rollout resource.
- Reference an existing Deployment from a Rollout using workloadRef field.

When converting a Deployment to a Rollout, it involves changing three fields:
1. Replacing the apiVersion from apps/v1 to argoproj.io/v1alpha1
2. Replacing the kind from Deployment to Rollout
3. Replacing the deployment strategy with a blue-green or canary strategy

For example, the following Rollout has been converted from a Deployment:
apiVersion: argoproj.io/v1alpha1  # Changed from apps/v1
kind: Rollout                     # Changed from Deployment
metadata:
  name: rollouts-demo
spec:
  selector:
    matchLabels:
      app: rollouts-demo
  template:
    metadata:
      labels:
        app: rollouts-demo
    spec:
      containers:
      - name: rollouts-demo
        image: argoproj/rollouts-demo:blue
        ports:
        - containerPort: 8080
  strategy:
    canary:                        # Changed from rollingUpdate or recreate
      steps:
      - setWeight: 20
      - pause: {}

Reference Deployment From Rollout:
- Instead of removing Deployment you can scale it down to zero and reference it from the Rollout resource:
    1. Create a Rollout resource.
    2. Reference an existing Deployment using workloadRef field.
    3. In the workloadRef field set the scaleDown attribute, which specifies how the Deployment should be scaled down. There are three options available:
        - never: the Deployment is not scaled down
        - onsuccess: the Deployment is scaled down after the Rollout becomes healthy
        - progressively: as the Rollout is scaled up the Deployment is scaled down.

For example, a Rollout resource referencing a Deployment:
apiVersion: argoproj.io/v1alpha1               # Create a rollout resource
kind: Rollout
metadata:
  name: rollout-ref-deployment
spec:
  replicas: 5
  selector:
    matchLabels:
      app: rollout-ref-deployment
  workloadRef:                                 # Reference an existing Deployment using workloadRef field
    apiVersion: apps/v1
    kind: Deployment
    name: rollout-ref-deployment
    scaleDown: onsuccess
  strategy:
    canary:
      steps:
        - setWeight: 20
        - pause: {duration: 10s}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: rollout-canary
  name: rollout-ref-deployment
spec:
  replicas: 0                                  # Scale down existing deployment
  selector:
    matchLabels:
      app: rollout-ref-deployment
  template:
    metadata:
      labels:
        app: rollout-ref-deployment
    spec:
      containers:
        - name: rollouts-demo
          image: argoproj/rollouts-demo:blue
          imagePullPolicy: Always
          ports:
            - containerPort: 8080

Best Practices:
- When migrating a Deployment which is already serving live production traffic, a Rollout 
should run next to the Deployment before deleting the Deployment or scaling down the Deployment. 
Not following this approach might result in downtime. It also allows for the Rollout to be tested 
before deleting the original Deployment. Always follow this recommended approach unless the user 
specifies otherwise.
- Do not delete the original Deployment until the user explicitly confirms that the Rollout is ready to take over production traffic.
- Ensure proper validation at each stage
- Document any configuration changes
- Add a succint summary of the actions you took and the results.
"""

argo_agent = AssistantAgent(
    "argo_agent",
    description="Argo Rollouts specialist for progressive delivery management.",
    tools=argo_management_tools + argo_setup_tools + k8s_tools,
    model_client=model_client,
    system_message=ARGO_SYSTEM_MESSAGE,
    reflect_on_tool_use=True,
)

user_proxy = UserProxyAgent("user_proxy", input_func=input)  # Use input() to get user input from console.

team = RoundRobinGroupChat(
    participants=[argo_agent, user_proxy],
    termination_condition=TextMentionTermination(text="TERMINATE"),
)

# Usage example:
task = "Help me translate my current reviews-v1 Kubernetes deployment in the default namespace to an Argo Rollout using the canary strategy and apply it to the cluster."

asyncio.run(Console(team.run_stream(task=task)))

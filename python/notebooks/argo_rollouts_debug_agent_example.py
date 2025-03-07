import asyncio

from autogen_agentchat.agents import AssistantAgent, UserProxyAgent
from autogen_agentchat.conditions import TextMentionTermination
from autogen_agentchat.teams import RoundRobinGroupChat
from autogen_agentchat.ui import Console
from autogen_ext.models.openai import OpenAIChatCompletionClient

# Import all required tools
from kagent.tools.argo import ArgoCRDTool, ArgoCRDToolConfig, PauseRollout, PromoteRollout, SetRolloutImage
from kagent.tools.k8s import (
    DescribeResource,
    GetPodLogs,
    GetResources,
    GetResourceYAML,
)
from kagent.tools.prometheus import (
    Config as PrometheusConfig,
)
from kagent.tools.prometheus import (
    LabelNamesTool,
    QueryRangeTool,
    QueryTool,
    SeriesQueryTool,
)

# Prometheus configuration with analysis thresholds
PROMETHEUS_CONFIG = PrometheusConfig(
    name="prom_config",
    base_url="http://localhost:9090/api/v1",
)

ARGO_DEBUG_SYSTEM_MESSAGE = """
You are an Argo debugging and deployment specialist focused on managing, troubleshooting, and resolving issues with Argo Rollouts deployments.
Assume that the Argo Rollouts controller is installed and configured correctly.

Core Capabilities:
1. Rollout Management:
   - Check rollout status and phase
   - Monitor progression
   - Identify stalled states
   - Track promotion status

2. Rollout Diagnostics:
   - Analyze failure conditions in the Argo Rollout resources statuses or Argo Rollouts controller logs
   - Debug promotion/abortion issues
   - Validate step execution
   - Verify traffic routing resources (Istio, Gateway API, etc)

3. Configuration Verification includes:
   - Validate rollout spec and status
   - Check analysis runs if an analysis is attached to the rollout.
   - Verify metric templates. The prometheus_agent will help you with this.
   - Confirm traffic routing rules. The k8s_agent will help you with this based on the traffic controller configuration (Istio, Gateway API, etc).

Standard Procedures:
1. Status Assessment:
   - Check rollout phase
   - Review analysis results
   - Validate configuration

2. Issue Resolution:
   - Identify root cause
   - Suggest remediation steps
   - Verify fixes
   - Document findings

Best Practices:
1. Always check rollout status first with the status rollout tool.
2. If an analysis is running, check the status of the analysis with the kubectl describe tool.
2. Identify the traffic controller configuration (Istio, Gateway API, etc) and validate the traffic routing rules.
4. If an analysis is running, Validate analysis metrics with the prometheus agent.
5. Document troubleshooting steps

Example commands:
# kubectl get rollouts -A -oyaml # list all rollouts in the cluster
# kubectl get rollout <name> -n <namespace> -oyaml # get a specific rollout resource
# kubectl describe analysisrun <name> -n <namespace> # describe a specific analysisrun
"""

# Create model client
model_client = OpenAIChatCompletionClient(
    model="gpt-4o",
)

# Create Argo debug agent with all necessary tools
argo_debug_agent = AssistantAgent(
    "argo_agent",
    description="Argo Rollouts specialist for deployment and debugging",
    tools=[
        PauseRollout(),
        PromoteRollout(),
        SetRolloutImage(),
        ArgoCRDTool(ArgoCRDToolConfig(model="gpt-4o-mini", openai_api_key=None)),
        GetResources(),
        GetResourceYAML(),
        DescribeResource(),
        GetPodLogs(),
        QueryTool(config=PROMETHEUS_CONFIG),
        QueryRangeTool(config=PROMETHEUS_CONFIG),
        SeriesQueryTool(config=PROMETHEUS_CONFIG),
        LabelNamesTool(config=PROMETHEUS_CONFIG),
    ],
    model_client=model_client,
    system_message=ARGO_DEBUG_SYSTEM_MESSAGE,
)

# Examples:
# task = "Create an Argo Rollout to deploy a new version of the demo application with the color purple using the Kubernetes Gateway API in my cluster."
# task = "Check if there are any argo rollout in the cluster in the process of promotion?"
# task = "Use the Kubernetes Gateway API and Argo Rollouts to create rollout resources for the canary and stable services for the demo application in my cluster."
# task = "Check if the Argo Rollouts controller is running and in a healthy state in the cluster?"
# task = "Create an Argo Rollout to deploy a new version of reviews-v1 using this image docker.io/istio/examples-bookinfo-reviews-v1:1.20.1?"
task = "Why is my reviews-v2 Argo Rollout not available?"

user_proxy = UserProxyAgent("user_proxy", input_func=input)  # Use input() to get user input from console.

team = RoundRobinGroupChat(
    participants=[argo_debug_agent, user_proxy],
    termination_condition=TextMentionTermination(text="TERMINATE"),
)

asyncio.run(Console(team.run_stream(task=task)))

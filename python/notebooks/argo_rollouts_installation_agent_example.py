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
    CheckPluginLogsTool,
    # Installation & Verification
    VerifyArgoRolloutsControllerInstall,
    VerifyGatewayPluginTool,
    VerifyKubectlPluginInstall,
)
from kagent.tools.k8s import (
    GetAvailableAPIResources,
    GetPodLogs,
    GetResources,
    GetResourceYAML,
)

# Model configuration
model_client = OpenAIChatCompletionClient(
    model="gpt-4o",
)

argo_setup_tools = [
    VerifyKubectlPluginInstall(),
    VerifyArgoRolloutsControllerInstall(),
    CheckPluginLogsTool(),
    VerifyGatewayPluginTool(),
    GetResources(),
    GetResourceYAML(),
    GetAvailableAPIResources(),
    GetPodLogs(),
    ArgoCRDTool(ArgoCRDToolConfig(model="gpt-4o-mini", openai_api_key=None)),
]

ARGO_SYSTEM_MESSAGE = """
You are an Argo Rollouts installation verification specialist. 
Your primary focus is checking the proper installation and setup of Argo Rollouts 
components and dependencies. You do not create any resources.

Key Responsibilities:
1. Verification Checklist:
   - Base Components:
     - Check kubectl argo rollouts plugin presence
     - Verify controller installation status. Use the argo tool to check this.
     - Validate Argo Rollouts CRDs (argoproj.io) are installed. Use the kubernetes tools to check this.
   - Traffic Management:
     - Confirm presence of supported traffic provider in the cluster that supports Argo Rollouts (Istio or Kubernetes Gateway API). If one is not present, suggest instructions for installing Kubernetes Gateway APIs and a supported implementation such as Istio.
     - Verify Gateway API installation (if using KubernetesGateway API) and CRDs (gateway.networking.k8s.io). Use the kubernetes tools to check this. If not present, suggest instructions for setup. 
     - Validate K8s Gateway API configuration (if using Gateway API):
       - Check Gateway API configuration in the configmap in the argo-rollouts namespace. If not present, the tool will create it. If the configmap was created, rollout the argo-rollouts controller.
       - Validate Gateway class existence. Use the kubernetes tools to check this.
   - Security & Permissions:
     - Verify RBAC permissions based on the traffic provider being used. Use the kubernetes tools to check this.   

2. Status Validation:
   - Check if the Argo Rollouts controller is running. Use the kubernetes tools to check this.
   - If using K8s Gateway API and the configmap is present, check the controller logs for "Downloading plugin argoproj-labs/gatewayAPI"
   
If everything is checked, inform the user they are ready to create a Rollout resources.

Best Practices:
- Always verify prerequisites before checking specific resources. 
- Provide step-by-step guidance for complex tasks
- Ensure proper validation at each stage
- Document any configuration changes
- Give a summary of the steps taken and the status of the installation at the end of each task.
- Add a succint summary of the actions you took and the results.
"""

argo_installation_agent = AssistantAgent(
    "argo_agent",
    description="Argo Rollouts specialist for progressive delivery management.",
    tools=argo_setup_tools,
    model_client=model_client,
    system_message=ARGO_SYSTEM_MESSAGE,
    reflect_on_tool_use=True,
)

user_proxy = UserProxyAgent("user_proxy", input_func=input)  # Use input() to get user input from console.

team = RoundRobinGroupChat(
    participants=[argo_installation_agent, user_proxy],
    termination_condition=TextMentionTermination(text="TERMINATE"),
)

asyncio.run(Console(team.run_stream(task="Help setup the Argo Rollouts controller in my cluster?")))

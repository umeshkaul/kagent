import importlib
from typing import AsyncGenerator, Dict, List

import yaml
from autogen_agentchat.agents import BaseChatAgent
from autogen_agentchat.base import ChatAgent, TaskResult
from autogen_agentchat.conditions import MaxMessageTermination, TextMentionTermination
from autogen_agentchat.messages import AgentEvent, ChatMessage
from autogen_agentchat.teams import SelectorGroupChat
from autogen_core.models import ChatCompletionClient
from autogen_core.tools import FunctionTool
from autogen_agentchat.teams import MagenticOneGroupChat

from .model import Agent, Team, Tool


class AutogenOrchestrator:
    def __init__(self, config_path: str):
        self.config_path = config_path
        self.teams: Dict[str, Team] = {}
        self.agents: Dict[str, Agent] = {}
        self.tools: Dict[str, Tool] = {}
        self.load_config()

    def load_config(self):
        """Load and parse YAML configuration files"""
        with open(self.config_path, "r") as f:
            configs = list(yaml.safe_load_all(f))

        for config in configs:
            kind = config["kind"]
            metadata = config["metadata"]
            spec = config["spec"]

            if kind == "AutogenTeam":
                self.teams[metadata["name"]] = Team(**spec)
            elif kind == "AutogenAgent":
                self.agents[metadata["name"]] = Agent(**spec)
            elif kind == "AutogenTool":
                self.tools[metadata["name"]] = Tool(**spec)

    def create_autogen_agent(self, agent: Agent, model_client: ChatCompletionClient) -> BaseChatAgent:
        """Create an AutoGen agent instance from configuration"""
        return agent.build(model_client)

    def get_team_agents(self, team_name: str) -> List[Agent]:
        """Get all agents belonging to a team based on label selector"""
        team = self.teams[team_name]

        team_agents = []
        for agent_name, agent in self.agents.items():
            agent_labels = self.get_agent_labels(agent_name)
            labels = team.selector.match_labels if team.selector else {}
            if all(agent_labels.get(k) == v for k, v in labels.items()):
                team_agents.append(agent)
        return team_agents

    def get_agent_labels(self, agent_name: str) -> Dict[str, str]:
        """Helper to get agent labels from configuration"""
        for config in yaml.safe_load_all(open(self.config_path)):
            if config["kind"] == "AutogenAgent" and config["metadata"]["name"] == agent_name:
                return config["metadata"].get("labels", {})
        return {}

    async def execute_prompt(
        self, team_name: str, prompt: str, model_client: ChatCompletionClient
    ) -> AsyncGenerator[AgentEvent | ChatMessage | TaskResult, None]:
        """Execute a prompt with the specified team"""
        team_agents = self.get_team_agents(team_name)

        # Create AutoGen agents
        autogen_agents: List[ChatAgent] = [self.create_autogen_agent(agent, model_client) for agent in team_agents]

        # # Add human proxy agent
        # human_proxy = UserProxyAgent(
        #     name="human_proxy",
        #     system_message="A proxy for human interaction. Validates final decisions and provides oversight.",
        #     human_input_mode="NEVER",
        #     max_consecutive_auto_reply=10
        # )
        # autogen_agents.append(human_proxy)

        text_mention_termination = TextMentionTermination("TERMINATE")
        max_messages_termination = MaxMessageTermination(max_messages=25)
        termination = text_mention_termination | max_messages_termination

        # Create group chat
        group_chat = SelectorGroupChat(
            autogen_agents,
            model_client=model_client,
            termination_condition=termination,
        )

        # Execute prompt
        async for item in group_chat.run_stream(task=prompt):
            yield item

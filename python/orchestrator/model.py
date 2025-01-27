import importlib
from abc import ABC, abstractmethod
from enum import Enum
from typing import Any, Dict, List, Optional

from autogen_agentchat.agents import AssistantAgent, BaseChatAgent
from autogen_core.models import ChatCompletionClient
from autogen_core.tools import BaseTool, FunctionTool
from pydantic import BaseModel, Field, GetCoreSchemaHandler, TypeAdapter
from pydantic_core import CoreSchema, core_schema
from mcp import ClientSession, StdioServerParameters


class ModelError(Exception):
    """Base class for all orchestrator errors"""


class ToolError(ModelError):
    """Error loading tool"""

    def __init__(self, tool_name: str, message: str):
        self.tool_name = tool_name
        self.message = message
        super().__init__(f"Tool {tool_name} error: {message}")


class ToolBuilder(ABC):
    """
    Builds a tool from a tool name
    """

    @abstractmethod
    def build(self, tool_name: str) -> list[BaseTool]:
        pass


class ToolType(Enum):
    Builtin = "BuiltIn"


class BuiltInTool(ToolBuilder, str):
    """
    https://docs.pydantic.dev/latest/concepts/types/#as-a-method-on-a-custom-type
    """

    @classmethod
    def __get_pydantic_core_schema__(cls, source_type: Any, handler: GetCoreSchemaHandler) -> CoreSchema:
        return core_schema.no_info_after_validator_function(cls, handler(str))

    def build(self, tool_name: str) -> list[BaseTool]:
        """Load builtin tool from tools package"""
        # Dynamically import and get the tool function from the tools package
        module_path = tool_name.split(".")
        module_name = "tools." + ".".join(module_path[:-1])
        try:
            module = importlib.import_module(module_name)
            tool = getattr(module, module_path[-1])
        except ImportError as e:
            raise ToolError(tool_name, "Tool not found") from e
        except AttributeError as e:
            raise ToolError(tool_name, "Tool not found") from e
        assert isinstance(tool, FunctionTool)
        return [tool]


# class McpTool(ToolBuilder, BaseModel):
#     name: str

#     def build(self, tool_name: str) -> list[BaseTool]:
#         return self.tools.build(tool_name)


class Tool(BaseModel):
    type: ToolType
    tool: BuiltInTool

    def build(self) -> list[BaseTool]:
        return self.tool.build(self.tool)


class AgentBuilder(ABC):
    @abstractmethod
    def build(self, name: str, model_client: ChatCompletionClient) -> BaseChatAgent:
        pass


class AgentType(Enum):
    AssistantAgent = "AssistantAgent"


class AssistantAgentBuilder(BaseModel):
    description: str
    system_message: str = Field(alias="systemMessage")
    tools: Optional[List[Tool]] = None

    def build(self, name: str, model_client: ChatCompletionClient) -> BaseChatAgent:
        return AssistantAgent(
            name=name,
            description=self.description,
            system_message=self.system_message,
            tools=[tool for f in (self.tools or []) for tool in f.build()],
            model_client=model_client,
        )


class Agent(BaseModel):
    name: str  # Needs to be unique
    type: AgentType
    agent: AssistantAgentBuilder  # will be a Union in the future

    def build(self, model_client: ChatCompletionClient) -> BaseChatAgent:
        return self.agent.build(self.name, model_client)


class Selector(BaseModel):
    match_labels: Dict[str, str] = Field(alias="matchLabels")


class Team(BaseModel):
    team_name: str = Field(alias="teamName")
    max_chat_rounds: Optional[int] = Field(alias="maxChatRounds")
    selector: Optional[Selector] = None

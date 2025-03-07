from typing import Sequence

from autogen_agentchat.base import TerminatedException, TerminationCondition
from autogen_agentchat.messages import (
    AgentEvent,
    ChatMessage,
    StopMessage,
    TextMessage,
)
from autogen_core import Component
from pydantic import BaseModel
from typing_extensions import Self


class TextMessageTerminationConfig(BaseModel):
    """Configuration for the TextMessageTermination termination condition."""

    source: str | None = None
    """The source of the text message to terminate the conversation."""


class TextMessageTermination(TerminationCondition, Component[TextMessageTerminationConfig]):
    """Terminate the conversation if a TextMessage is received."""

    component_config_schema = TextMessageTerminationConfig
    component_provider_override = "kagent.terminations.TextMessageTermination"

    def __init__(self, source: str | None = None) -> None:
        self._terminated = False
        self._source = source

    @property
    def terminated(self) -> bool:
        return self._terminated

    async def __call__(self, messages: Sequence[AgentEvent | ChatMessage]) -> StopMessage | None:
        if self._terminated:
            raise TerminatedException("Termination condition has already been reached")
        for message in messages:
            if isinstance(message, TextMessage) and (self._source is None or message.source == self._source):
                self._terminated = True
                return StopMessage(content="Stop message received", source="TextMessageTermination")
        return None

    async def reset(self) -> None:
        self._terminated = False

    def _to_config(self) -> TextMessageTerminationConfig:
        return TextMessageTerminationConfig(source=self._source)

    @classmethod
    def _from_config(cls, config: TextMessageTerminationConfig) -> Self:
        return cls(source=config.source)

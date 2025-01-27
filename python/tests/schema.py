from dataclasses import dataclass
from typing import Any, Optional


@dataclass
class AgentMetadata:
    description: str
    version: str


@dataclass
class AgentDefinition:
    name: str
    system_messages: list[str]
    metadata: AgentMetadata
    tools: Optional[list[str]] = None


@dataclass
class TestMetadata:
    description: str


@dataclass
class TestCase:
    name: str
    category: str
    input: str
    expected_output: dict[str, Any]


@dataclass
class TestSuite:
    version: str
    metadata: TestMetadata
    test_cases: list[TestCase]


@dataclass
class TestResult:
    input: str
    expected_output: Optional[str]
    actual_output: str
    category: str
    duration_ms: float
    similarity: float


@dataclass
class TestRunResult:
    timestamp: str
    config: dict[str, Any]
    results: list[TestResult]

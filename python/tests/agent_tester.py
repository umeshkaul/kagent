import difflib
import hashlib
import json
import logging
from dataclasses import asdict
from datetime import datetime
from pathlib import Path

from autogen_agentchat.agents import AssistantAgent
from schema import TestCase, TestResult, TestRunResult


class AgentTester:
    def __init__(self, agent: "AssistantAgent", test_cases: list[TestCase], results_dir: str = "test_results"):
        self.agent = agent
        self.test_cases = test_cases
        self.results_dir = Path(results_dir)
        self.results_dir.mkdir(exist_ok=True)

    async def run_tests(self) -> TestRunResult:
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

        model = self.agent._model_client._to_config().model_dump().get("model")

        # Get all system messages once for the entire test run
        system_msg = "\n".join([msg.content for msg in self.agent._system_messages])

        # Create config once for the entire test run
        config = {
            "agent_name": self.agent.name,
            "tools": "\n".join([tool.name for tool in self.agent._tools]),
            "model": model,
            "prompt": system_msg,
            "prompt_hash": hashlib.sha256(system_msg.encode("utf-8")).hexdigest(),
        }

        results = []
        for test_case in self.test_cases:
            start = datetime.now()
            # Run the agent with the test input
            response = await self.agent.run(task=test_case.input)
            end = datetime.now()

            duration_ms = (end - start).total_seconds() * 1000

            # Only get the last TextMessage and get the content of it
            output = response.messages[-1].content

            similarity = self._calculate_similarity(test_case.expected_output, json.loads(output))

            # Create test result without config
            result = TestResult(
                category=test_case.category,
                input=test_case.input,
                expected_output=test_case.expected_output,
                actual_output=json.loads(output),
                duration_ms=duration_ms,
                similarity=similarity,
            )
            results.append(result)

        # Create the test run result that combines config and individual results
        test_run_result = TestRunResult(timestamp=timestamp, config=config, results=results)

        # Save results
        self._save_results(test_run_result)

        return test_run_result

    def _calculate_similarity(self, dict1: dict, dict2: dict):
        # Convert both dictionaries to strings with consistent formatting
        str1 = json.dumps(dict1, sort_keys=True)
        str2 = json.dumps(dict2, sort_keys=True)

        return difflib.SequenceMatcher(None, str1, str2).ratio() * 100

    def _save_results(self, test_run_result: TestRunResult):
        # Convert results to JSON-serializable format
        results_dict = {
            "timestamp": test_run_result.timestamp,
            "config": test_run_result.config,
            "results": [asdict(result) for result in test_run_result.results],
        }

        # Save to JSON file
        results_file = (
            self.results_dir / f"results_{test_run_result.timestamp}_{test_run_result.config.get('model')}.json"
        )
        with open(results_file, "w") as f:
            json.dump(results_dict, f, indent=2)

        logging.info(f"Results saved to: {results_file}")

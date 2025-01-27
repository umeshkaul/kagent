import argparse
import asyncio
import os
from dotenv import load_dotenv
from pathlib import Path

from agent_tester import AgentTester
from autogen_ext.models.openai import OpenAIChatCompletionClient
from autogen_ext.models.openai.config import ResponseFormat
from comparator import analyze_results_command, compare_results_command
from dotenv import load_dotenv
from loader import create_agent, load_agent_definition, load_test_cases
import logging

load_dotenv()


async def run_test_command(test_cases_file: Path, agent_def_file: Path, model: str, results_dir: str = "test_results"):
    test_suite = load_test_cases(test_cases_file)
    agent_def = load_agent_definition(agent_def_file)

    model_client = OpenAIChatCompletionClient(
        model=model,
        response_format=ResponseFormat(type="json_object"),
    )

    agent = create_agent(agent_def, model_client)

    tester = AgentTester(agent, test_cases=test_suite.test_cases, results_dir=results_dir)

    logging.info(f"Running {len(test_suite.test_cases)} test cases for agent: {agent_def.name}...")
    results = await tester.run_tests()
    return results


def main():
    parser = argparse.ArgumentParser(description="Test runner for agents")
    subparsers = parser.add_subparsers(dest="command", help="Command to execute")

    # Run tests command
    run_parser = subparsers.add_parser("run", help="Run tests from a YAML file")
    run_parser.add_argument("test_file", type=Path, help="Path to the YAML file containing test cases")
    run_parser.add_argument("agent_file", type=Path, help="Path to the agent definition YAML file")
    run_parser.add_argument(
        "--results-dir",
        type=str,
        default="test_results",
        help="Directory to store test results (default: test_results)",
    )
    run_parser.add_argument("--model", type=str, help="OpenAI model to use for testing", required=True)

    # Compare results command
    compare_parser = subparsers.add_parser("compare", help="Compare two test result files")
    compare_parser.add_argument("file1", type=Path, help="First test results file")
    compare_parser.add_argument("file2", type=Path, help="Second test results file")

    # Analyze results command
    analyze_parser = subparsers.add_parser("analyze", help="Analyze a single test results file")
    analyze_parser.add_argument("file", type=Path, help="Test results file to analyze")

    args = parser.parse_args()

    if args.command == "run":
        asyncio.run(run_test_command(args.test_file, args.agent_file, args.model, args.results_dir))
    elif args.command == "compare":
        compare_results_command(args.file1, args.file2)
    elif args.command == "analyze":
        analyze_results_command(args.file)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()

import difflib
import json
import logging
from pathlib import Path
from typing import Any


def analyze_results_command(results_file: Path) -> dict[str, Any]:
    """Analyze a single test run results file for similarity to expected outputs."""
    if not results_file.exists():
        raise FileNotFoundError(f"Results file not found: {results_file}")

    with open(results_file) as f:
        run = json.load(f)

    test_analyses = []
    total_similarity = 0
    total_duration = 0
    similarity_ranges = {
        "excellent": 0,  # 90-100%
        "good": 0,  # 75-90%
        "fair": 0,  # 50-75%
        "poor": 0,  # <50%
    }

    for test in run["results"]:
        # Calculate similarity between expected and actual output
        expected_str = json.dumps(test["expected_output"], sort_keys=True)
        actual_str = json.dumps(test["actual_output"], sort_keys=True)
        similarity = difflib.SequenceMatcher(None, expected_str, actual_str).ratio() * 100

        # Track similarity distribution
        if similarity >= 90:
            similarity_ranges["excellent"] += 1
        elif similarity >= 75:
            similarity_ranges["good"] += 1
        elif similarity >= 50:
            similarity_ranges["fair"] += 1
        else:
            similarity_ranges["poor"] += 1

        total_similarity += similarity
        total_duration += test["duration_ms"]

        # Generate diff if outputs don't match
        differences = []
        if test["actual_output"] != test["expected_output"]:
            diff = difflib.unified_diff(
                json.dumps(test["expected_output"], indent=2).splitlines(),
                json.dumps(test["actual_output"], indent=2).splitlines(),
                fromfile="expected",
                tofile="actual",
                lineterm="",
            )
            differences = list(diff)

        analysis = {
            "category": test["category"],
            "input": test["input"],
            "similarity": similarity,
            "duration_ms": test["duration_ms"],
            "differences": differences,
        }
        test_analyses.append(analysis)

    num_tests = len(test_analyses)
    analysis_results = {
        "file": results_file.name,
        "model": run["config"]["model"],
        "total_tests": num_tests,
        "test_analyses": test_analyses,
        "summary": {
            "avg_similarity": total_similarity / num_tests if num_tests > 0 else 0,
            "avg_duration": total_duration / num_tests if num_tests > 0 else 0,
            "similarity_distribution": similarity_ranges,
            "similarity_by_category": {},
        },
    }

    # Calculate average similarity by category
    category_totals = {}
    category_counts = {}
    for analysis in test_analyses:
        cat = analysis["category"]
        if cat not in category_totals:
            category_totals[cat] = 0
            category_counts[cat] = 0
        category_totals[cat] += analysis["similarity"]
        category_counts[cat] += 1

    for cat in category_totals:
        analysis_results["summary"]["similarity_by_category"][cat] = {
            "avg_similarity": category_totals[cat] / category_counts[cat],
            "test_count": category_counts[cat],
        }

    # Print analysis results
    logging.info(f"\nAnalyzing {results_file.name}")
    logging.info(f"Model: {run['config']['model']}")
    logging.info(f"\nAnalyzing {num_tests} test cases:")

    # Print summary statistics
    logging.info("\n=== Summary Statistics ===")
    logging.info(f"Average similarity to expected output: {analysis_results['summary']['avg_similarity']:.2f}%")
    logging.info(f"Average duration: {analysis_results['summary']['avg_duration']:.2f}ms")

    logging.info("\nSimilarity Distribution:")
    logging.info(f"Excellent (90-100%): {similarity_ranges['excellent']} tests")
    logging.info(f"Good (75-90%): {similarity_ranges['good']} tests")
    logging.info(f"Fair (50-75%): {similarity_ranges['fair']} tests")
    logging.info(f"Poor (<50%): {similarity_ranges['poor']} tests")

    logging.info("\nPerformance by Category:")
    for cat, stats in analysis_results["summary"]["similarity_by_category"].items():
        logging.info(f"{cat}:")
        logging.info(f"  - Average similarity: {stats['avg_similarity']:.2f}%")
        logging.info(f"  - Number of tests: {stats['test_count']}")

    logging.info("\nDetailed Test Analysis:")
    for analysis in test_analyses:
        logging.info(f"\n=== Test Category: {analysis['category']} ===")
        logging.info(f"Input: {analysis['input']}")
        logging.info(f"Similarity to expected: {analysis['similarity']:.2f}%")
        logging.info(f"Duration: {analysis['duration_ms']:.2f}ms")

        if analysis["differences"]:
            logging.info("Differences from expected output:")
            for line in analysis["differences"]:
                logging.info(line)

    return analysis_results


def compare_results_command(results_file1: Path, results_file2: Path) -> dict[str, Any]:
    """Compare two specific test result files including detailed test case analysis."""
    if not results_file1.exists():
        raise FileNotFoundError(f"Results file not found: {results_file1}")
    if not results_file2.exists():
        raise FileNotFoundError(f"Results file not found: {results_file2}")

    with open(results_file1) as f:
        run1 = json.load(f)
    with open(results_file2) as f:
        run2 = json.load(f)

    # Compare results
    test_comparisons = []
    total_similarity_run1 = 0
    total_similarity_run2 = 0
    better_similarity_count_run1 = 0
    better_similarity_count_run2 = 0
    equal_similarity_count = 0

    for idx, (test1, test2) in enumerate(zip(run1["results"], run2["results"], strict=False), 1):
        similarity1 = test1.get("similarity", 0)
        similarity2 = test2.get("similarity", 0)
        total_similarity_run1 += similarity1
        total_similarity_run2 += similarity2

        # Track which run performed better for this test
        if similarity1 > similarity2:
            better_similarity_count_run1 += 1
        elif similarity2 > similarity1:
            better_similarity_count_run2 += 1
        else:
            equal_similarity_count += 1

        comparison = {
            "test_number": idx,
            "input": test1["input"],
            "differences": [],
            "duration_delta": test2.get("duration_ms", 0) - test1.get("duration_ms", 0),
            "similarity_old": similarity1,
            "similarity_new": similarity2,
            "similarity_delta": similarity2 - similarity1,
        }

        # Compare outputs
        if test1["actual_output"] != test2["actual_output"]:
            diff = difflib.unified_diff(
                json.dumps(test1["actual_output"], indent=2).splitlines(),
                json.dumps(test2["actual_output"], indent=2).splitlines(),
                fromfile=f"{results_file1.name} (Test {idx})",
                tofile=f"{results_file2.name} (Test {idx})",
                lineterm="",
            )
            comparison["differences"] = list(diff)

        test_comparisons.append(comparison)

    num_tests = len(test_comparisons)
    avg_similarity_run1 = total_similarity_run1 / num_tests if num_tests > 0 else 0
    avg_similarity_run2 = total_similarity_run2 / num_tests if num_tests > 0 else 0

    comparison_results = {
        "file1": results_file1.name,
        "file2": results_file2.name,
        "test_comparisons": test_comparisons,
        "total_tests": num_tests,
        "tests_with_differences": sum(1 for t in test_comparisons if t["differences"]),
        "prompt_changed": run1["config"]["prompt"] != run2["config"]["prompt"],
        "model_changed": run1["config"]["model"] != run2["config"]["model"],
        "config_changed": run1["config"] != run2["config"],
        "run1_stats": {
            "model": run1["config"]["model"],
            "avg_similarity": avg_similarity_run1,
            "better_tests_count": better_similarity_count_run1,
        },
        "run2_stats": {
            "model": run2["config"]["model"],
            "avg_similarity": avg_similarity_run2,
            "better_tests_count": better_similarity_count_run2,
        },
        "equal_tests_count": equal_similarity_count,
    }

    # Print detailed comparison results
    logging.info(f"\nComparing {results_file1.name} with {results_file2.name}")
    logging.info(f"Model changed: {comparison_results['model_changed']}")
    logging.info(f"Prompt changed: {comparison_results['prompt_changed']}")
    logging.info(f"\nAnalyzing {comparison_results['total_tests']} test cases:")

    for test in test_comparisons:
        logging.info(f"\n=== Test {test['test_number']} ===")
        logging.info(f"Input: {test['input']}")
        logging.info(f"Duration delta: {test['duration_delta']:.2f}ms")
        logging.info(
            f"Similarity: {test['similarity_old']:.2f}% → {test['similarity_new']:.2f}% "
            f"(Δ: {test['similarity_delta']:+.2f}%)"
        )

        if test["differences"]:
            logging.info("Output differences:")
            for line in test["differences"]:
                logging.info(line)
        else:
            logging.info("No differences in output")

    # Print enhanced summary statistics
    if test_comparisons:
        avg_similarity_delta = sum(t["similarity_delta"] for t in test_comparisons) / num_tests
        avg_duration_delta = sum(t["duration_delta"] for t in test_comparisons) / num_tests

        logging.info("\n=== Summary Statistics ===")
        logging.info(
            f"Tests with differences: {comparison_results['tests_with_differences']} of {comparison_results['total_tests']}"
        )

        # Print overall performance comparison
        logging.info("\nOverall Performance Comparison:")
        logging.info(f"Run 1 ({run1['config']['model']}):")
        logging.info(f"  - Average similarity: {avg_similarity_run1:.2f}%")
        logging.info(f"  - Better performance in: {better_similarity_count_run1} tests")

        logging.info(f"\nRun 2 ({run2['config']['model']}):")
        logging.info(f"  - Average similarity: {avg_similarity_run2:.2f}%")
        logging.info(f"  - Better performance in: {better_similarity_count_run2} tests")

        logging.info(f"\nEqual performance in: {equal_similarity_count} tests")
        logging.info(f"Overall similarity delta: {avg_similarity_delta:+.2f}%")
        logging.info(f"Overall duration delta: {avg_duration_delta:+.2f}ms")

        # Determine overall winner
        if avg_similarity_run1 > avg_similarity_run2:
            winner = f"Run 1 ({run1['config']['model']})"
            margin = avg_similarity_run1 - avg_similarity_run2
        elif avg_similarity_run2 > avg_similarity_run1:
            winner = f"Run 2 ({run2['config']['model']})"
            margin = avg_similarity_run2 - avg_similarity_run1
        else:
            winner = "Tie"
            margin = 0

        if winner != "Tie":
            logging.info(f"\nOverall Winner: {winner}")
            logging.info(f"Winning margin: {margin:.2f}% higher average similarity")

    return comparison_results

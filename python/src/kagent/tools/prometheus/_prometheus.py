import urllib.parse
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Union

from autogen_core.tools import FunctionTool
from typing_extensions import Annotated

from ..common.client import HttpClient


class QueryResponseFormat(str, Enum):
    JSON = "json"
    PROMETHEUS = "prometheus"


class TargetState(str, Enum):
    ACTIVE = "active"
    DROPPED = "dropped"
    ANY = "any"


async def _query(
    query: Annotated[str, "Prometheus expression query string"],
) -> Dict[str, Any]:
    """Executes an instant query at a single point in time."""
    base_url = "http://localhost:9090/api/v1"
    c = HttpClient(base_url=base_url)

    async with c as client:
        params = {
            "query": query,
        }
        return await client._make_request("GET", "query", params=params)


async def _query_range(
    query: Annotated[str, "Prometheus expression query string"],
    start: Annotated[Union[datetime, float], "Start timestamp"],
    end: Annotated[Union[datetime, float], "End timestamp"],
    step: Annotated[Union[str, float], "Query resolution step width"],
    timeout: Annotated[Optional[str], "Evaluation timeout"],
) -> Dict[str, Any]:
    """Evaluates an expression query over a range of time."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "query": query,
            "start": client._format_time(start),
            "end": client._format_time(end),
            "step": str(step),
            "timeout": timeout,
        }
        return await client._make_request("GET", "query_range", params=params)


async def _get_series(
    match: Annotated[List[str], "Series selector arguments"],
    start: Annotated[Optional[Union[datetime, float]], "Start timestamp"],
    end: Annotated[Optional[Union[datetime, float]], "End timestamp"],
    limit: Annotated[Optional[int], "Maximum number of returned series"],
) -> List[Dict[str, str]]:
    """Returns the list of time series that match a certain label set."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "match[]": match,
            "start": client._format_time(start) if start else None,
            "end": client._format_time(end) if end else None,
            "limit": limit,
        }
        result = await client._make_request("GET", "series", params=params)
        return result.get("data", [])


async def _get_label_names(
    start: Annotated[Optional[Union[datetime, float]], "Start timestamp"],
    end: Annotated[Optional[Union[datetime, float]], "End timestamp"],
    match: Annotated[Optional[List[str]], "Series selector"],
    limit: Annotated[Optional[int], "Maximum number of returned items"],
) -> List[str]:
    """Returns a list of label names."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "start": client._format_time(start) if start else None,
            "end": client._format_time(end) if end else None,
            "match[]": match,
            "limit": limit,
        }
        result = await client._make_request("GET", "labels", params=params)
        return result.get("data", [])


async def _get_label_values(
    label_name: Annotated[str, "Label name"],
    start: Annotated[Optional[Union[datetime, float]], "Start timestamp"],
    end: Annotated[Optional[Union[datetime, float]], "End timestamp"],
    match: Annotated[Optional[List[str]], "Series selector"],
    limit: Annotated[Optional[int], "Maximum number of returned items"],
) -> List[str]:
    """Returns a list of label values for a provided label name."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        # URL encode the label name for safety
        encoded_label = urllib.parse.quote(label_name)
        params = {
            "start": client._format_time(start) if start else None,
            "end": client._format_time(end) if end else None,
            "match[]": match,
            "limit": limit,
        }
        result = await client._make_request("GET", f"label/{encoded_label}/values", params=params)
        return result.get("data", [])


async def _get_targets(
    state: Annotated[Optional[TargetState], "Target state filter"],
    scrape_pool: Annotated[Optional[str], "Scrape pool name"],
) -> Dict[str, Any]:
    """Returns an overview of the current state of the Prometheus target discovery."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "state": state.value if state else None,
            "scrape_pool": scrape_pool,
        }
        return await client._make_request("GET", "targets", params=params)


async def _get_rules(
    type: Annotated[Optional[str], "Rule type filter"],
    rule_name: Annotated[Optional[List[str]], "Rule names filter"],
    rule_group: Annotated[Optional[List[str]], "Rule group names filter"],
    file: Annotated[Optional[List[str]], "File paths filter"],
    exclude_alerts: Annotated[Optional[bool], "Exclude alerts flag"],
    match: Annotated[Optional[List[str]], "Label selectors"],
    group_limit: Annotated[Optional[int], "Group limit"],
    group_next_token: Annotated[Optional[str], "Pagination token"],
) -> Dict[str, Any]:
    """Returns a list of alerting and recording rules."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "type": type,
            "rule_name[]": rule_name,
            "rule_group[]": rule_group,
            "file[]": file,
            "exclude_alerts": "true" if exclude_alerts else None,
            "match[]": match,
            "group_limit": group_limit,
            "group_next_token": group_next_token,
        }
        return await client._make_request("GET", "rules", params=params)


async def _get_alerts() -> Dict[str, Any]:
    """Returns a list of all active alerts."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "alerts")


async def _get_target_metadata(
    match_target: Annotated[Optional[str], "Target label selectors"],
    metric: Annotated[Optional[str], "Metric name"],
    limit: Annotated[Optional[int], "Maximum number of targets"],
) -> List[Dict[str, Any]]:
    """Returns metadata about metrics currently scraped from targets."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "match_target": match_target,
            "metric": metric,
            "limit": limit,
        }
        result = await client._make_request("GET", "targets/metadata", params=params)
        return result.get("data", [])


async def _get_alertmanagers() -> Dict[str, Any]:
    """Returns an overview of the current state of the Prometheus alertmanager discovery."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "alertmanagers")


async def _get_metadata(
    metric: Annotated[Optional[str], "Metric name"],
    limit: Annotated[Optional[int], "Maximum number of metrics"],
    limit_per_metric: Annotated[Optional[int], "Maximum entries per metric"],
) -> Dict[str, List[Dict[str, Any]]]:
    """Returns metadata about metrics currently scraped from targets."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "metric": metric,
            "limit": limit,
            "limit_per_metric": limit_per_metric,
        }
        result = await client._make_request("GET", "metadata", params=params)
        return result.get("data", {})


async def _get_status_config() -> Dict[str, str]:
    """Returns currently loaded configuration file."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "status/config")


async def _get_status_flags() -> Dict[str, str]:
    """Returns flag values that Prometheus was configured with."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "status/flags")


async def _get_status_runtime_info() -> Dict[str, Any]:
    """Returns various runtime information properties about the Prometheus server."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "status/runtimeinfo")


async def _get_status_build_info() -> Dict[str, str]:
    """Returns various build information properties about the Prometheus server."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "status/buildinfo")


async def _get_status_tsdb(
    limit: Annotated[Optional[int], "Number of items limit"],
) -> Dict[str, Any]:
    """Returns various cardinality statistics about the Prometheus TSDB."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {"limit": limit} if limit is not None else None
        return await client._make_request("GET", "status/tsdb", params=params)


async def _create_snapshot(
    skip_head: Annotated[Optional[bool], "Skip head block flag"],
) -> Dict[str, str]:
    """Creates a snapshot of all current data."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {"skip_head": "true" if skip_head else None}
        return await client._make_request("POST", "admin/tsdb/snapshot", params=params)


async def _delete_series(
    match: Annotated[List[str], "Series selectors"],
    start: Annotated[Optional[Union[datetime, float]], "Start timestamp"],
    end: Annotated[Optional[Union[datetime, float]], "End timestamp"],
) -> None:
    """Deletes data for a selection of series in a time range."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        params = {
            "match[]": match,
            "start": client._format_time(start) if start else None,
            "end": client._format_time(end) if end else None,
        }
        await client._make_request("POST", "admin/tsdb/delete_series", params=params)


async def _clean_tombstones() -> None:
    """Removes the deleted data from disk and cleans up the existing tombstones."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        await client._make_request("POST", "admin/tsdb/clean_tombstones")


async def _get_status_wal_replay() -> Dict[str, Any]:
    """Returns information about the WAL replay."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "status/walreplay")


async def _get_notifications() -> Dict[str, Any]:
    """Returns a list of all currently active notifications."""
    async with HttpClient(base_url="http://localhost:9090/api/v1/") as client:
        return await client._make_request("GET", "notifications")


query = FunctionTool(
    _query,
    description="Execute an instant Prometheus query at a single point in time. This function allows you to evaluate PromQL expressions and retrieve current metric values. The query can include any valid PromQL expressions including aggregations, mathematical operations, and label matchers.",
    name="query",
)

query_range = FunctionTool(
    _query_range,
    description="Execute a Prometheus query over a time range. This allows you to retrieve metric values over a specified time period with a given resolution step. Useful for generating time series data for graphing or analysis of metric behavior over time.",
    name="query_range",
)

get_series = FunctionTool(
    _get_series,
    description="Retrieve time series data matching specified label selectors. This function allows you to discover which time series exist in your Prometheus database that match certain criteria, including their labels and time range.",
    name="get_series",
)

get_label_names = FunctionTool(
    _get_label_names,
    description="Retrieve all label names that exist in the Prometheus database. This helps in discovering what labels are available for querying and filtering metrics, optionally filtered by time range and series selectors.",
    name="get_label_names",
)

get_label_values = FunctionTool(
    _get_label_values,
    description="Retrieve all possible values for a specific label name in the Prometheus database. This helps in understanding what values exist for a particular label, useful for building queries and understanding metric dimensions.",
    name="get_label_values",
)

get_targets = FunctionTool(
    _get_targets,
    description="Retrieve information about all Prometheus monitoring targets and their current state. This provides visibility into which targets are being scraped successfully, which have been dropped, and their associated metadata.",
    name="get_targets",
)

get_rules = FunctionTool(
    _get_rules,
    description="Retrieve all configured Prometheus rules, including both recording rules and alerting rules. This provides insight into how metrics are being preprocessed and what conditions trigger alerts.",
    name="get_rules",
)

get_alerts = FunctionTool(
    _get_alerts,
    description="Retrieve all currently active alerts in Prometheus. This shows all alert conditions that are currently firing, including their labels, annotations, and when they became active.",
    name="get_alerts",
)

get_target_metadata = FunctionTool(
    _get_target_metadata,
    description="Retrieve metadata about metrics from specific targets. This includes information such as metric help strings, type information, and unit information for metrics from matched targets.",
    name="get_target_metadata",
)

get_metadata = FunctionTool(
    _get_metadata,
    description="Retrieve metadata about metrics across all targets. This provides consolidated metadata information about metrics, including their types, help strings, and unit information.",
    name="get_metadata",
)

get_alertmanagers = FunctionTool(
    _get_alertmanagers,
    description="Retrieve information about all Alertmanager instances known to Prometheus. This shows both active and dropped Alertmanagers and their current status.",
    name="get_alertmanagers",
)

get_status_config = FunctionTool(
    _get_status_config,
    description="Retrieve the current Prometheus configuration. This returns the active configuration file content in YAML format, showing how Prometheus is currently configured.",
    name="get_status_config",
)

get_status_flags = FunctionTool(
    _get_status_flags,
    description="Retrieve all command-line flags that Prometheus was started with. This shows the current operational parameters of the Prometheus server.",
    name="get_status_flags",
)

get_status_runtime_info = FunctionTool(
    _get_status_runtime_info,
    description="Retrieve runtime information about the Prometheus server, including garbage collection statistics, goroutine count, and other operational metrics.",
    name="get_status_runtime_info",
)

get_status_build_info = FunctionTool(
    _get_status_build_info,
    description="Retrieve build information about the Prometheus server, including version, build time, commit hash, and Go version used for building.",
    name="get_status_build_info",
)

get_status_tsdb = FunctionTool(
    _get_status_tsdb,
    description="Retrieve statistics about the Prometheus time series database (TSDB), including cardinality information, label statistics, and storage details.",
    name="get_status_tsdb",
)

create_snapshot = FunctionTool(
    _create_snapshot,
    description="Create a snapshot of the current Prometheus data. This creates a point-in-time backup of the TSDB data, useful for backup purposes or offline analysis. Requires admin API to be enabled.",
    name="create_snapshot",
)

delete_series = FunctionTool(
    _delete_series,
    description="Delete specific time series data from Prometheus. This allows removal of metrics matching certain label selectors within a time range. Requires admin API to be enabled.",
    name="delete_series",
)

clean_tombstones = FunctionTool(
    _clean_tombstones,
    description="Clean up deleted time series data from disk. This removes tombstones created by the delete_series operation and frees up disk space. Requires admin API to be enabled.",
    name="clean_tombstones",
)

get_status_wal_replay = FunctionTool(
    _get_status_wal_replay,
    description="Retrieve information about the Write-Ahead Log (WAL) replay status. This shows the progress of WAL replay during Prometheus startup, which is crucial for understanding database recovery status.",
    name="get_status_wal_replay",
)

get_notifications = FunctionTool(
    _get_notifications,
    description="Retrieve all active notifications from Prometheus. This shows current system notifications about the Prometheus server's state and operations.",
    name="get_notifications",
)

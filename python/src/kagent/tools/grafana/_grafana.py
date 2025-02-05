from typing import Any, Optional

from autogen_core.models import SystemMessage, UserMessage
from autogen_core.tools import FunctionTool
from autogen_ext.models.openai import OpenAIChatCompletionClient
from typing_extensions import Annotated

from ..common.client import HttpClient

SYSTEM_PROMPT = """You are a Grafana dashboard configuration expert. Your task is to generate valid JSON configurations for Grafana dashboards based on user requirements. Follow these guidelines precisely:

Dashboard Structure:
1. Every dashboard must have a root object containing:
   - id: null for new dashboards
   - uid: unique identifier string
   - title: dashboard name
   - tags: array of relevant tags
   - timezone: default "browser"
   - schemaVersion: current Grafana schema version (typically 36)
   - version: dashboard version number
   - time: object with "from" and "to" time range
   - refresh: refresh interval (e.g., "5s")

Panel Configuration:
1. Each panel requires:
   - id: unique numeric identifier within dashboard
   - gridPos: object with {x, y, w, h} for panel positioning
   - title: panel title
   - type: panel type (e.g., "timeseries", "stat", "gauge")
   - datasource: object specifying data source type and uid
   - targets: array of query objects
   - options: panel-specific display options
   - fieldConfig: field formatting and threshold settings

2. For timeseries panels include:
   - options.tooltip.mode: "single" or "multi"
   - options.legend.displayMode: "list" or "table"
   - fieldConfig.defaults.custom.lineWidth
   - fieldConfig.defaults.custom.fillOpacity
   - fieldConfig.defaults.custom.spanNulls

3. For stat panels include:
   - options.textMode: "value" or "value_and_name"
   - options.colorMode: "value" or "background"
   - options.graphMode: "area" or "none"
   - fieldConfig.defaults.thresholds

Query Construction:
1. Each target object must contain:
   - refId: unique query identifier (A, B, C, etc.)
   - datasource: same as panel datasource
   - expr: for Prometheus/Loki queries
   - query: for SQL-based sources
   - format: "time_series" or "table"

2. For PromQL queries:
   - Use proper function names and parameters
   - Include rate() for counter metrics
   - Use proper label matchers with =, !=, =~, !~
   - Group metrics with by() or without()

Template Variables:
1. Include templating.list array containing:
   - name: variable name
   - type: "query", "interval", "custom", etc.
   - datasource: for query variables
   - query: variable query string
   - regex: for value extraction
   - refresh: when to update options
   - includeAll: boolean to allow selecting all options
   - multi: boolean to allow multiple selections

Best Practices:
1. Use description field to document dashboard purpose
2. Include proper panel descriptions
3. Set appropriate min/max steps for Y-axes
4. Configure reasonable thresholds for metrics
5. Use consistent naming conventions
6. Include links to relevant documentation
7. Set appropriate decimals for numeric displays
8. Configure proper null value handling
9. Use variables for reusable components

Error Prevention:
1. Verify all required fields are present
2. Ensure panel IDs are unique
3. Validate gridPos to prevent overlap
4. Check data source references exist
5. Verify query syntax is valid
6. Ensure all arrays are properly terminated
7. Validate all object properties are quoted
8. Check for proper JSON escaping

When generating the configuration:
1. Ask for specific requirements including:
   - Metrics to display
   - Panel types needed
   - Time range requirements
   - Refresh interval needs
   - Template variable requirements
   - Threshold settings
   - Legend preferences
   - Color scheme preferences

2. Generate the configuration in valid JSON format with proper indentation

3. Include comments explaining key configuration choices

Respond with:
1. Complete, valid JSON configuration
2. Explanation of key configuration choices
3. List of required data sources
4. Any assumptions made
5. Suggestions for improvements

Example panel configuration structure:
```json
{
  "id": 1,
  "gridPos": {
    "x": 0,
    "y": 0,
    "w": 12,
    "h": 8
  },
  "type": "timeseries",
  "title": "Request Rate",
  "datasource": {
    "type": "prometheus",
    "uid": "prometheus"
  },
  "targets": [
    {
      "refId": "A",
      "expr": "rate(http_requests_total[5m])",
      "legendFormat": "{{method}} {{path}}"
    }
  ],
  "options": {
    "tooltip": {
      "mode": "multi"
    },
    "legend": {
      "displayMode": "table"
    }
  },
  "fieldConfig": {
    "defaults": {
      "custom": {
        "lineWidth": 2,
        "fillOpacity": 20
      },
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "value": null,
            "color": "green"
          },
          {
            "value": 100,
            "color": "red"
          }
        ]
      }
    }
  }
}
```

Remember to validate the final JSON configuration and ensure all required fields are present and properly formatted.
"""

def get_model_client():
    # TODO: We should have a way to configure externally somehow.
    return OpenAIChatCompletionClient(
        model="gpt-4o-mini",
    )

async def _create_dashboard(
    user_query: Annotated[str, "User query that explains what type of a dashboard to generate"],
) -> Any:

    try:
        model_client = get_model_client()
        result = await model_client.create(
            messages=[SystemMessage(content=SYSTEM_PROMPT), UserMessage(content=user_query, source="user")],
            json_output=True,
        )
        return result.content
    except Exception as e:
        return f"Error generating grafana dashboard json: {str(e)}"

async def _apply_dashboard(
    dashboard: Annotated[str, "The dashboard JSON to create in Grafana"]
) -> Any:
    async with HttpClient(base_url="http://localhost:3000/api", api_key="TODO") as client:
        data = {"dashboard": dashboard, "overwrite": True, "folderUid": "TODO"}
        return await client._make_request("POST", "dashboards/db",  data=data)

generate_dashboard_json = FunctionTool(
    _create_dashboard,
    description="Create a Grafana dashboard JSON from user query",
    name="generate_dashboard_json",
)

create_dashboard = FunctionTool(
    _apply_dashboard,
    description="Create a new Grafana dashboard from JSON",
    name="create_dashboard",
)

from autogen_core.tools import FunctionTool

def _is_prometheus_installed():
    return True

is_prometheus_installed = FunctionTool(
    _is_prometheus_installed,
    description="Check if Prometheus is installed",
    name="is_prometheus_installed",
    )

from typing import Annotated, Optional

from autogen_core.tools import FunctionTool

from .._utils import create_typed_fn_tool
from ..common.shell import run_command


async def _verify_install() -> str:
    return _run_istioctl_command("verify-install")


verify_install = FunctionTool(
    _verify_install,
    description="Verify Istio installation status",
    name="verify_install",
)

VerifyInstall, VerifyInstallToolConfig = create_typed_fn_tool(verify_install, "kagent.tools.istio.VerifyInstallTool", "VerifyInstallTool")


async def _proxy_config(
    pod_name: Annotated[str, "The name of the pod to get proxy configuration for"],
    ns: Annotated[Optional[str], "The namespace of the pod to get proxy configuration for"]
) -> str:
    return _run_istioctl_command(
        f"proxy-config all {'-n ' + ns if ns else ''} {pod_name}"
    )


proxy_config = FunctionTool(
    _proxy_config,
    description="Get proxy configuration for 1 pod",
    name="proxy_config",
)

ProxyConfig, ProxyConfigToolConfig = create_typed_fn_tool(proxy_config, "kagent.tools.istio.ProxyConfigTool", "ProxyConfigTool")


# Function that runs the istioctl command in the shell
def _run_istioctl_command(command: str) -> str:
    return run_command("istioctl", command.split(" "))

from typing import Annotated, Optional

from autogen_core.tools import FunctionTool

from ..common.shell import run_command


async def _verify_install() -> str:
    return _run_istioctl_command("verify-install")

async def _install(
    profile: Annotated[Optional[str], "The Istio profile to install (e.g. default, ambient)"] = "ambient",
) -> str:
    return _run_istioctl_command(f"install --set profile={profile} -y")


async def _uninstall(
    purge: Annotated[Optional[bool], "Whether to purge Istio resources"] = True,
) -> str:
    return _run_istioctl_command(f"uninstall -y {'--purge' if purge else ''}")


async def _proxy_config(
    pod_name: Annotated[str, "The name of the pod to get proxy configuration for"],
    ns: Annotated[Optional[str], "The namespace of the pod to get proxy configuration for"],
) -> str:
    return _run_istioctl_command(f"proxy-config all {'-n ' + ns if ns else ''} {pod_name}")


verify_install = FunctionTool(
    _verify_install,
    description="Verify Istio installation status",
    name="verify_install",
)

install = FunctionTool(
    _install,
    description="Install Istio",
    name="install",
)

uninstall = FunctionTool(
    _uninstall,
    description="Uninstall Istio",
    name="uninstall",
)

proxy_config = FunctionTool(
    _proxy_config,
    description="Get proxy configuration for 1 pod",
    name="proxy_config",
)


# Function that runs the istioctl command in the shell
def _run_istioctl_command(command: str) -> str:
    return run_command("istioctl", command.split(" "))

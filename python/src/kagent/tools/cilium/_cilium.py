from typing import Literal

from autogen_core.tools import FunctionTool
from typing_extensions import Annotated

from .._utils import create_typed_fn_tool
from ..common._shell import run_command


def _cilium_status() -> str:
    return _run_cilium_cli("status")

cilium_status = FunctionTool(
    _cilium_status,
    "Get the status of Cilium installation.",
    name="cilium_status",
)

CiliumStatus, CiliumStatusConfig = create_typed_fn_tool(
    cilium_status, "kagent.tools.cilium.CiliumStatus", "CiliumStatus"
)

def _upgrade_cilium(
        cluster_name: Annotated[str, "The name of the cluster to upgrade Cilium on"] = None,
        datapath_mode: Annotated[Literal["tunnel", "native", "aws-eni", "gke", "azure", "aks-byocni"], "The datapath mode to use for Cilium"] = None,
        ) -> str:
    return _run_cilium_cli(f"upgrade f{'' if cluster_name else '--cluster-name {cluster_name}'} f{'' if datapath_mode else '--datapath-mode {datapath_mode}'}")

upgrade_cilium = FunctionTool(
    _upgrade_cilium,
    "Upgrade Cilium on the cluster.",
    name="upgrade_cilium",
)

UpgradeCilium, UpgradeCiliumConfig = create_typed_fn_tool(
    upgrade_cilium, "kagent.tools.cilium.UpgradeCilium", "UpgradeCilium"

)
def _install_cilium(
        cluster_name: Annotated[str, "The name of the cluster to install Cilium on"] = None,
        datapath_mode: Annotated[Literal["tunnel", "native", "aws-eni", "gke", "azure", "aks-byocni"], "The datapath mode to use for Cilium"] = None,
        ) -> str:
    return _run_cilium_cli(f"install f{'' if cluster_name else '--cluster-name {cluster_name}'} f{'' if datapath_mode else '--datapath-mode {datapath_mode}'}")

install_cilium = FunctionTool(
    _install_cilium,
    "Install Cilium on the cluster.",
    name="install_cilium",
)

InstallCilium, InstallCiliumConfig = create_typed_fn_tool(
    install_cilium, "kagent.tools.cilium.InstallCilium", "InstallCilium"
)

def _uninstall_cilium() -> str:
    return _run_cilium_cli("uninstall")

uninstall_cilium = FunctionTool(
    _uninstall_cilium,
    "Uninstall Cilium from the cluster.",
    name="uninstall_cilium",
)

UninstallCilium, UninstallCiliumConfig = create_typed_fn_tool(
    uninstall_cilium, "kagent.tools.cilium.UninstallCilium", "UninstallCilium"
)

def _run_cilium_cli(command: str) -> str:
    cmd_parts = command.split(" ")
    cmd_parts = [part for part in cmd_parts if part]  # Remove empty strings from the list
    return run_command("cilium", cmd_parts)



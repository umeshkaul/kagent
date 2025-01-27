import tempfile
from typing import Annotated, Optional

from autogen_core.tools import FunctionTool

from ..common.shell import run_command


def _get_pod(
    pod_name: Annotated[str, "The name of the pod to get information about"],
    ns: Annotated[Optional[str], "The namespace of the pod to get information about"],
    output: Annotated[Optional[str], "The output format of the pod information"],
) -> str:
    return _run_kubectl_command(
        f"get pod {pod_name + ' ' if pod_name else ''}{'-n' + ns + ' ' if ns else ''}{'-o' + output if output else ''}"
    )


def _get_pods(
    ns: Annotated[Optional[str], "The namespace of the pod to get information about"],
    all_namespaces: Annotated[Optional[bool], "Whether to get pods from all namespaces"],
    output: Annotated[Optional[str], "The output format of the pod information"],
) -> str:
    if ns and all_namespaces:
        raise ValueError("Cannot specify both ns and all_namespaces=True")
    return _run_kubectl_command(
        f"get pods {'-n' + ns + ' ' if ns else ''}{'-o' + output if output else ''} {'-A' if all_namespaces else ''}"
    )


def _get_services(
    service_name: Annotated[Optional[str], "The name of the service to get information about"],
    all_namespaces: Annotated[Optional[bool], "Whether to get services from all namespaces"],
    ns: Annotated[Optional[str], "The namespace of the service to get information about"],
    output: Annotated[Optional[str], "The output format of the service information"],
) -> str:
    if service_name and all_namespaces:
        raise ValueError("Cannot specify both service_name and all_namespaces=True")

    return _run_kubectl_command(
        f"get services {service_name + ' ' if service_name else ''}{'-n' + ns + ' ' if ns else ''}{'-o' + output if output else ''} {'-A' if all_namespaces else ''}"
    )


def _get_resources(
    name: Annotated[str, "The name of the resource to get information about"],
    resource_type: Annotated[str, "The type of resource to get information about"],
    all_namespaces: Annotated[Optional[bool], "Whether to get resources from all namespaces"],
    ns: Annotated[Optional[str], "The namespace of the resource to get information about"],
    output: Annotated[Optional[str], "The output format of the resource information"],
) -> str:
    if name and all_namespaces:
        raise ValueError("Cannot specify both name and all_namespaces=True")

    return _run_kubectl_command(
        f"get {resource_type} {name if name else ''} {'-n' + ns + ' ' if ns else ''}{'-o' + output if output else ''} {'-A' if all_namespaces else ''}"
    )


def _apply_manifest(
    manifest: Annotated[str, "The path to the manifest file to apply"],
) -> str:
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=True) as tmp_file:
        tmp_file.write(manifest)
        tmp_file.flush()  # Ensure the content is written to disk
        return _run_kubectl_command(f"apply -f {tmp_file.name}")


def _get_pod_logs(
    pod_name: Annotated[str, "The name of the pod to get logs from"],
    ns: Annotated[str, "The namespace of the pod to get logs from"],
):
    return _run_kubectl_command(f"logs {pod_name + ' ' if pod_name else ''}{'-n' + ns if ns else ''}")


get_pods = FunctionTool(
    _get_pods,
    description="Gets pods in Kubernetes from a specific namespace. Always prefer output type `wide` unless otherwise specified.",
    name="get_pods",
)

get_pod = FunctionTool(
    _get_pod,
    description="Gets a single pod in Kubernetes. Always prefer output type `wide` unless otherwise specified.",
    name="get_pod",
)

get_services = FunctionTool(
    _get_services,
    description="Get information about services in Kubernetes. Always prefer output type `wide` unless otherwise specified.",
    name="get_services",
)

apply_manifest = FunctionTool(
    _apply_manifest,
    description="Apply a manifest file to the Kubernetes cluster.",
    name="_apply_manifest",
)

get_resources = FunctionTool(
    _get_resources,
    description="Get information about resources in Kubernetes. Always prefer output type `wide` unless otherwise specified.",
    name="get_resources",
)

get_pod_logs = FunctionTool(
    _get_pod_logs,
    description="Get logs from a specific pod in Kubernetes.",
    name="get_pod_logs",
)


def _run_kubectl_command(command: str) -> str:
    # Split the command and remove empty strings
    cmd_parts = command.split(" ")
    cmd_parts = [part for part in cmd_parts if part]  # Remove empty strings from the list
    return run_command("kubectl", cmd_parts)

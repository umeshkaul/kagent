import random
import tempfile
from enum import Enum
from typing import Annotated, Optional

from autogen_core.tools import FunctionTool

from .._utils import create_typed_fn_tool
from ..common.shell import run_command


def _check_service_connectivity(
    service_name: Annotated[
        Optional[str],
        "Fully qualified service name with port number (e.g. my-service.my-namespace.svc.cluster.local:80)",
    ],
    context: dict = {},
) -> str:
    pod_name = f"curlpod-{random.randint(0, 1000)}"
    _run_kubectl_command(f"run {pod_name} --image=curlimages/curl --restart=Never --command -- sleep 3600", context=context)
    _run_kubectl_command(f"wait --for=condition=ready pod/{pod_name} --timeout=60s", context=context)
    curl_result = _run_kubectl_command(f"exec {pod_name} -- curl {service_name}", context=context)
    _run_kubectl_command(f"delete pod {pod_name}", context=context)
    return curl_result


def _patch_resource(
    resource_type: Annotated[str, "The type of resource to patch (deployment, configmap, pod, service, ...)"],
    resource_name: Annotated[str, "The name of the resource to patch"],
    patch: Annotated[str, "The patch to apply to the resource"],
    ns: Annotated[Optional[str], "The namespace of the resource to patch"],
):
    cmd_parts = ["patch", resource_type, resource_name]
    if ns:
        cmd_parts.extend(["-n", ns])
    cmd_parts.extend(["--type=merge", f"--patch={patch}"])

    return run_command("kubectl", cmd_parts)


def _scale(
    resource_type: Annotated[str, "The type of resource to scale (deployment, statefulset, ...)"],
    name: Annotated[str, "The name of the resource to scale"],
    replicas: Annotated[int, "The number of replicas to scale to"],
    ns: Annotated[Optional[str], "The namespace of the resource to scale"],
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"scale {resource_type}/{name} --replicas={replicas} {f'-n {ns} ' if ns else ''}", context=context)


def _remove_annotation(
    resource_type: Annotated[
        str, "The type of resource to remove the annotation from (deployment, service pod, node, ...)"
    ],
    name: Annotated[str, "The name of the resource to remove the annotation from"],
    annotation_key: Annotated[str, "The key of the annotation to remove"],
    ns: Annotated[Optional[str], "The namespace of the resource to remove the annotation from"],
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"annotate {resource_type} {name} {f'-n {ns} ' if ns else ''} {annotation_key}-", context=context)


def _annotate_resource(
    resource_type: Annotated[str, "The type of resource to annotate (deployment, service, pod, node, ...)"],
    name: Annotated[str, "The name of the resource to annotate"],
    annotations: Annotated[dict[str, str], "The annotations to apply to the resource"],
    ns: Annotated[Optional[str], "The namespace of the resource to annotate"],
    context: dict = {},
) -> str:
    annotation_string = " ".join([f"{key}={value}" for key, value in annotations.items()])
    return _run_kubectl_command(f"annotate {resource_type} {name} {f'-n {ns} ' if ns else ''} {annotation_string}", context=context)


def _remove_label(
    resource_type: Annotated[
        str, "The type of resource to remove the label from (deployment, service, pod, node, ...)"
    ],
    name: Annotated[str, "The name of the resource to remove the label from"],
    label_key: Annotated[str, "The key of the label to remove"],
    ns: Annotated[Optional[str], "The namespace of the resource to remove the label from"],
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"label {resource_type} {name} {f'-n {ns} ' if ns else ''} {label_key}-", context=context)


def _label_resource(
    resource_type: Annotated[str, "The type of resource to label (deployment, service, pod, node, ...)"],
    name: Annotated[str, "The name of the resource to label"],
    labels: Annotated[dict[str, str], "The labels to apply to the resource"],
    ns: Annotated[Optional[str], "The namespace of the resource to label"],
    context: dict = {},
) -> str:
    label_string = " ".join([f"{key}={value}" for key, value in labels.items()])
    return _run_kubectl_command(f"label {resource_type} {name} {f'-n {ns} ' if ns else ''} {label_string}", context=context)


def _create_resource(
    resource_yaml: Annotated[str, "The YAML definition of the resource to create"],
    context: dict = {},
) -> str:
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=True) as tmp_file:
        tmp_file.write(resource_yaml)
        tmp_file.flush()  # Ensure the content is written to disk
        return _run_kubectl_command(f"create -f {tmp_file.name}", context=context)


def _get_events(context: dict = {}) -> str:
    return _run_kubectl_command("get events", context=context)


def _rollout(
    action: Annotated[
        str, "The rollout action to perform on the resource (history, pause, restart, resume, status, undo)"
    ],
    resource_type: Annotated[str, "The type of resource to rollout (deployment, daemonset, ...)"],
    name: Annotated[str, "The name of the resource to rollout"],
    ns: Annotated[Optional[str], "The namespace of the resource to rollout"],
    context: dict = {},
) -> str:
        return _run_kubectl_command(f"rollout {action} {resource_type}/{name} {f'-n {ns} ' if ns else ''}", context=context)


def _get_available_api_resources(context: dict = {}) -> str:
    return _run_kubectl_command("api-resources", context=context)


def _get_cluster_configuration(context: dict = {}) -> str:
    return _run_kubectl_command("config view", context=context)


def _describe_resource(
    resource_type: Annotated[str, "The type of resource to describe (deployment, service, pod, node, ...)"],
    name: Annotated[str, "The name of the resource to describe"],
    ns: Annotated[Optional[str], "The namespace of the resource to describe"],
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"describe {resource_type} {name} {f'-n {ns}' if ns else ''}", context=context)


def _delete_resource(
    resource_type: Annotated[str, "The type of resource to delete (deployment, service, pod, node, ...)"],
    name: Annotated[str, "The name of the resource to delete"],
    ns: Annotated[str, "The namespace of the resource to delete"],
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"delete {resource_type} {name} {f'-n {ns}' if ns else ''}", context=context)


def _get_resource_yaml(
    resource_type: Annotated[
        str, "The type of resource to get the YAML definition for (deployment, service, pod, node, ...)"
    ],
    name: Annotated[
        Optional[str],
        "The name of the resource to get the YAML definition for. If not provided, all resources of the given type will be returned.",
    ],
    ns: Annotated[Optional[str], "The namespace of the resource to get the definition for"],    
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"get {resource_type} {name if name else ''} {f'-n {ns} ' if ns else ''} -o yaml", context=context)


def _execute_command(
    pod_name: Annotated[str, "The name of the pod to execute the command in"],
    ns: Annotated[str, "The namespace of the pod to execute the command in"],
    command: Annotated[str, "The command to execute inside the pod"],
    context: dict = {},
) -> str:
    return _run_kubectl_command(f"exec {pod_name} {f'-n {ns} ' if ns else ''} -- {command}", context=context)


def _get_resources(
    name: Annotated[
        Optional[str],
        "The name of the resource to get information about. If not provided, all resources of the given type will be returned.",
    ],
    resource_type: Annotated[
        str,
        "The type of resource to get information about (deployment, service, pod, node, ...). 'all' is NOT an option, you must specify a resource type.",
    ],
    all_namespaces: Annotated[Optional[bool], "Whether to get resources from all namespaces"],
    ns: Annotated[
        Optional[str],
        "The namespace of the resource to get information about, if unset will default to the current namespace",
    ],
    output: Annotated[Optional[str], "The output format of the resource information"],
    context: dict = {},
) -> str:
    if name and all_namespaces:
        # only use the name if provided, and ignore all_namespaces
        all_namespaces = False

    return _run_kubectl_command(
        f"get {resource_type} {name if name else ''} {'-n' + ns + ' ' if ns else ''}{'-o' + output if output else ''} {'-A' if all_namespaces else ''}",
        context=context)


def _apply_manifest(
    manifest: Annotated[str, "The path to the manifest file to apply"],
    context: dict = {},
) -> str:
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=True) as tmp_file:
        tmp_file.write(manifest)
        tmp_file.flush()  # Ensure the content is written to disk
        return _run_kubectl_command(f"apply -f {tmp_file.name}",
        context=context)


def _get_pod_logs(
    pod_name: Annotated[str, "The name of the pod to get logs from"],
    ns: Annotated[Optional[str], "The namespace of the pod to get logs from"] = "default",
    num_lines: Annotated[Optional[int], "The number of lines to get from the logs"] = 50,
    context: dict = {},
):
    return _run_kubectl_command(f"logs {pod_name} {f'-n {ns}' if ns else ''} --tail {num_lines}", context=context)


check_service_connectivity = FunctionTool(
    _check_service_connectivity,
    description="Check connectivity to a service in Kubernetes.",
    name="check_service_connectivity",
)

CheckServiceConnectivity, CheckServiceConnectivityConfig = create_typed_fn_tool(
    check_service_connectivity, "kagent.tools.k8s.CheckServiceConnectivity", "CheckServiceConnectivity"
)

patch_resource = FunctionTool(
    _patch_resource,
    description="Patch a resource in Kubernetes.",
    name="patch_resource",
)

PatchResource, PatchResourceConfig = create_typed_fn_tool(
    patch_resource, "kagent.tools.k8s.PatchResource", "PatchResource"
)

scale = FunctionTool(
    _scale,
    description="Scale a resource in Kubernetes.",
    name="scale",
)

Scale, ScaleConfig = create_typed_fn_tool(scale, "kagent.tools.k8s.Scale", "Scale")

remove_annotation = FunctionTool(
    _remove_annotation,
    description="Remove an annotation from a resource in Kubernetes.",
    name="remove_annotation",
)

RemoveAnnotation, RemoveAnnotationConfig = create_typed_fn_tool(
    remove_annotation, "kagent.tools.k8s.RemoveAnnotation", "RemoveAnnotation"
)

annotate_resource = FunctionTool(
    _annotate_resource,
    description="Annotate a resource in Kubernetes.",
    name="annotate_resource",
)

AnnotateResource, AnnotateResourceConfig = create_typed_fn_tool(
    annotate_resource, "kagent.tools.k8s.AnnotateResource", "AnnotateResource"
)

remove_label = FunctionTool(
    _remove_label,
    description="Remove a label from a resource in Kubernetes.",
    name="remove_label",
)

RemoveLabel, RemoveLabelConfig = create_typed_fn_tool(remove_label, "kagent.tools.k8s.RemoveLabel", "RemoveLabel")

label_resource = FunctionTool(
    _label_resource,
    description="Label a resource in Kubernetes.",
    name="label_resource",
)

LabelResource, LabelResourceConfig = create_typed_fn_tool(
    label_resource, "kagent.tools.k8s.LabelResource", "LabelResource"
)

create_resource = FunctionTool(
    _create_resource,
    description="Create a resource in Kubernetes.",
    name="create_resource",
)

CreateResource, CreateResourceConfig = create_typed_fn_tool(
    create_resource, "kagent.tools.k8s.CreateResource", "CreateResource"
)

get_events = FunctionTool(
    _get_events,
    description="Get the events in the Kubernetes cluster.",
    name="get_events",
)

GetEvents, GetEventsConfig = create_typed_fn_tool(get_events, "kagent.tools.k8s.GetEvents", "GetEvents")

rollout = FunctionTool(
    _rollout,
    description="Perform a rollout on a resource in Kubernetes.",
    name="rollout",
)

Rollout, RolloutConfig = create_typed_fn_tool(rollout, "kagent.tools.k8s.Rollout", "Rollout")

get_available_api_resources = FunctionTool(
    _get_available_api_resources,
    description="Gets the supported API resources in Kubernetes.",
    name="get_available_api_resources",
)

GetAvailableAPIResources, GetAvailableAPIResourcesConfig = create_typed_fn_tool(
    get_available_api_resources, "kagent.tools.k8s.GetAvailableAPIResources", "GetAvailableAPIResources"
)

get_cluster_configuration = FunctionTool(
    _get_cluster_configuration,
    description="Get the configuration of the Kubernetes cluster.",
    name="get_cluster_configuration",
)

GetClusterConfiguration, GetClusterConfigurationConfig = create_typed_fn_tool(
    get_cluster_configuration, "kagent.tools.k8s.GetClusterConfiguration", "GetClusterConfiguration"
)

describe_resource = FunctionTool(
    _describe_resource,
    description="Describe a resource in Kubernetes.",
    name="describe_resource",
)

DescribeResource, DescribeResourceConfig = create_typed_fn_tool(
    describe_resource, "kagent.tools.k8s.DescribeResource", "DescribeResource"
)

delete_resource = FunctionTool(
    _delete_resource,
    description="Delete a resource in Kubernetes.",
    name="delete_resource",
)

DeleteResource, DeleteResourceConfig = create_typed_fn_tool(
    delete_resource, "kagent.tools.k8s.DeleteResource", "DeleteResource"
)

get_resource_yaml = FunctionTool(
    _get_resource_yaml,
    description="Get the YAML representation of a resource in Kubernetes.",
    name="get_resource_yaml",
)

GetResourceYAML, GetResourceYAMLConfig = create_typed_fn_tool(
    get_resource_yaml, "kagent.tools.k8s.GetResourceYAML", "GetResourceYAML"
)


execute_command = FunctionTool(
    _execute_command,
    description="Executes a command inside a pod in Kubernetes. For example, to run `ls` in a pod named `my-pod` in the namespace `my-namespace`, use `execute_command('my-pod', 'my-namespace', 'ls')`.",
    name="execute_command",
)

ExecuteCommand, ExecuteCommandConfig = create_typed_fn_tool(
    execute_command, "kagent.tools.k8s.ExecuteCommand", "ExecuteCommand"
)

apply_manifest = FunctionTool(
    _apply_manifest,
    description="Apply a YAML resource file to the Kubernetes cluster.",
    name="_apply_manifest",
)

ApplyManifest, ApplyManifestConfig = create_typed_fn_tool(
    apply_manifest, "kagent.tools.k8s.ApplyManifest", "ApplyManifest"
)


get_resources = FunctionTool(
    _get_resources,
    description="Get information about resources in Kubernetes. Always prefer output type `wide` unless otherwise specified. 'all' is NOT an option, you must specify a resource type.",
    name="get_resources",
)

GetResources, GetResourcesConfig = create_typed_fn_tool(get_resources, "kagent.tools.k8s.GetResources", "GetResources")


get_pod_logs = FunctionTool(
    _get_pod_logs,
    description="Get logs from a specific pod in Kubernetes.",
    name="get_pod_logs",
)

GetPodLogs, GetPodLogsConfig = create_typed_fn_tool(get_pod_logs, "kagent.tools.k8s.GetPodLogs", "GetPodLogs")


def _run_kubectl_command(command: str, context: dict = {}) -> str:
    # Split the command and remove empty strings
    cmd_parts = command.split(" ")
    cmd_parts = [part for part in cmd_parts if part]  # Remove empty strings from the list
    # if context has a user_id key
    if "user_id" in context:
        cmd_parts = ["--as", context['user_id'], "--as-group", "system:authenticated"] + cmd_parts
    return run_command("kubectl", cmd_parts)

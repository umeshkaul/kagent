from typing import Literal, Optional

from autogen_core.tools import FunctionTool
from typing_extensions import Annotated

from .._utils import create_typed_fn_tool
from ..common._shell import run_command

# Tools for running cilium-dbg command inside the Cilium pod

def _get_cilium_pod_name() -> str:
    # Get the name of the Cilium pod in the cluster where we can run the cilium-dbg command
    cilium_pod_name = run_command("kubectl", ["get", "pod", "-l", "k8s-app=cilium", "-o", "name", "-n", "kube-system"])
    if not cilium_pod_name:
        raise ValueError("No Cilium pod found in the cluster - make sure Cilium is installed and running")

    return cilium_pod_name.strip()

def _run_cilium_dbg_command(command: str) -> str:
    cilium_pod_name = _get_cilium_pod_name()
    cmd_parts = command.split(" ")
    return run_command("kubectl", ["exec", "-it", cilium_pod_name, "-n", "kube-system", "--", "cilium-dbg", *cmd_parts])

def _get_endpoint_details(
        endpoint_id: Annotated[str, "The ID of the endpoint to get details for"],
        labels: Annotated[Optional[str], "The labels of the endpoint to get details for"] = None,
        output_format: Annotated[Literal["json", "yaml", "jsonpath"], "The output format of the endpoint details"] = "json") -> str:
    if labels:
        return _run_cilium_dbg_command(f"endpoint get -l {labels} -o {output_format}")
    else:
        return _run_cilium_dbg_command(f"endpoint get {endpoint_id} -o {output_format}")

get_endpoint_details = FunctionTool(
    _get_endpoint_details, "List the details of an endpoint in the cluster", name="get_endpoint_details"
)
GetEndpointDetails, GetEndpointDetailsConfig = create_typed_fn_tool(
    get_endpoint_details, "kagent.tools.cilium.GetEndpointDetails", "GetEndpointDetails"
)

def _get_endpoint_logs(endpoint_id: Annotated[str, "The ID of the endpoint to get logs for"]) -> str:
    return _run_cilium_dbg_command(f"endpoint get {endpoint_id}")

get_endpoint_logs = FunctionTool(
    _get_endpoint_logs, "Get the logs of an endpoint in the cluster", name="get_endpoint_logs"
)
GetEndpointLogs, GetEndpointLogsConfig = create_typed_fn_tool(
    get_endpoint_logs, "kagent.tools.cilium.GetEndpointLogs", "GetEndpointLogs"
)


def _get_endpoint_health(endpoint_id: Annotated[str, "The ID of the endpoint to get health for"]) -> str:
    return _run_cilium_dbg_command(f"endpoint get {endpoint_id}")

get_endpoint_health = FunctionTool(
    _get_endpoint_health, "Get the health of an endpoint in the cluster", name="get_endpoint_health"
)

GetEndpointHealth, GetEndpointHealthConfig = create_typed_fn_tool(
    get_endpoint_health, "kagent.tools.cilium.GetEndpointHealth", "GetEndpointHealth"
)

def _manage_endpoint_labels(endpoint_id: Annotated[str, "The ID of the endpoint to manage labels for"], labels: Annotated[dict[str, str], "The labels to manage for the endpoint"], action: Annotated[Literal["add", "delete"], "The action to perform on the labels"]) -> str:
    if action == "add":
        return _run_cilium_dbg_command(f"endpoint labels {endpoint_id} --add {labels}")
    elif action == "delete":
        return _run_cilium_dbg_command(f"endpoint labels {endpoint_id} --delete {labels}")

manage_endpoint_labels = FunctionTool(
    _manage_endpoint_labels, "Manage the labels (add or delete) of an endpoint in the cluster", name="manage_endpoint_labels"
)
ManageEndpointLabels, ManageEndpointLabelsConfig = create_typed_fn_tool(
    manage_endpoint_labels, "kagent.tools.cilium.ManageEndpointLabels", "ManageEndpointLabels"
)

def _manage_endpoint_configuration(
        endpoint_id: Annotated[str, "The ID of the endpoint to manage configuration for"],
        config: Annotated[list[str], "The configuration to manage for the endpoint provided as a list of key-value pairs (e.g. ['DropNotification=false', 'TraceNotification=false'])"]) -> str:
    return _run_cilium_dbg_command(f"endpoint config {endpoint_id} {' '.join(config)}")

manage_endpoint_configuration = FunctionTool(
    _manage_endpoint_configuration, "Manage the configuration of an endpoint in the cluster", name="manage_endpoint_configuration"
)
ManageEndpointConfig, ManageEndpointConfigConfig = create_typed_fn_tool(
    manage_endpoint_configuration, "kagent.tools.cilium.ManageEndpointConfig", "ManageEndpointConfig"
)

def _disconnect_endpoint(endpoint_id: Annotated[str, "The ID of the endpoint to disconnect from the network"]) -> str:
    return _run_cilium_dbg_command(f"endpoint disconnect {endpoint_id}")

disconnect_endpoint = FunctionTool(
    _disconnect_endpoint, "Disconnect an endpoint from the network", name="disconnect_endpoint"
)
DisconnectEndpoint, DisconnectEndpointConfig = create_typed_fn_tool(
    disconnect_endpoint, "kagent.tools.cilium.DisconnectEndpoint", "DisconnectEndpoint")

def _get_endpoints_list() -> str:
    return _run_cilium_dbg_command("endpoint list")

get_endpoints_list = FunctionTool(
    _get_endpoints_list,
    "Get the list of all endpoints in the cluster.",
    name="get_endpoints_list",
)

GetEndpointsList, GetEndpointsListConfig = create_typed_fn_tool(
    get_endpoints_list, "kagent.tools.cilium.GetEndpointsList", "GetEndpointsList"
)

def _list_identities() -> str:
    return _run_cilium_dbg_command("identity list")

list_identities = FunctionTool(
    _list_identities, "List all identities in the cluster", name="list_identities"
)
ListIdentities, ListIdentitiesConfig = create_typed_fn_tool(
    list_identities, "kagent.tools.cilium.ListIdentities", "ListIdentities"
)

def _get_identity_details(identity_id: Annotated[str, "The ID of the identity to get details for"]) -> str:
    return _run_cilium_dbg_command(f"identity get {identity_id}")

get_identity_details = FunctionTool(
    _get_identity_details, "Get the details of an identity in the cluster", name="get_identity_details"
)
GetIdentityDetails, GetIdentityDetailsConfig = create_typed_fn_tool(
    get_identity_details, "kagent.tools.cilium.GetIdentityDetails", "GetIdentityDetails"
)

def _show_configuration_options(
        list_all: Annotated[bool, "Whether to list all configuration options"] = False,
        list_read_only: Annotated[bool, "Whether to list read-only configuration options"] = False,
        list_options: Annotated[bool, "Whether to list options"] = False) -> str:
    if list_all:
        return _run_cilium_dbg_command("endpoint config --all")
    elif list_read_only:
        return _run_cilium_dbg_command("endpoint config -r")
    elif list_options:
        return _run_cilium_dbg_command("endpoint config --list-options")
    else:
        return _run_cilium_dbg_command("endpoint config")

show_configuration_options = FunctionTool(
    _show_configuration_options, "Show Cilium configuration options", name="show_configuration_options"
)
ShowConfigurationOptions, ShowConfigurationOptionsConfig = create_typed_fn_tool(
    show_configuration_options, "kagent.tools.cilium.ShowConfigurationOptions", "ShowConfigurationOptions"
)

def _toggle_configuration_option(option: Annotated[str, "The option to toggle"], value: Annotated[bool, "The value to set the option to"]) -> str:
    return _run_cilium_dbg_command(f"endpoint config {option}={'enable' if value else 'disable'}")

toggle_configuration_option = FunctionTool(
    _toggle_configuration_option, "Toggle a Cilium configuration option", name="toggle_configuration_option"
)
ToggleConfigurationOption, ToggleConfigurationOptionConfig = create_typed_fn_tool(
    toggle_configuration_option, "kagent.tools.cilium.ToggleConfigurationOption", "ToggleConfigurationOption"
)

def _request_debugging_information() -> str:
    return _run_cilium_dbg_command("debuginfo")

request_debugging_information = FunctionTool(
    _request_debugging_information, "Request debugging information from Cilium agent", name="request_debugging_information"
)
RequestDebuggingInformation, RequestDebuggingInformationConfig = create_typed_fn_tool(
    request_debugging_information, "kagent.tools.cilium.RequestDebuggingInformation", "RequestDebuggingInformation"
)

def _display_encryption_state() -> str:
    return _run_cilium_dbg_command("encrypt state")

display_encryption_state = FunctionTool(
    _display_encryption_state, "Display the current encryption state", name="display_encryption_state"
)
DisplayEncryptionState, DisplayEncryptionStateConfig = create_typed_fn_tool(
    display_encryption_state, "kagent.tools.cilium.DisplayEncryptionState", "DisplayEncryptionState"
)

def _flush_ipsec_state() -> str:
    return _run_cilium_dbg_command("encrypt flush -f")

flush_ipsec_state = FunctionTool(
    _flush_ipsec_state, "Flush the IPsec state", name="flush_ipsec_state"
)
FlushIPsecState, FlushIPsecStateConfig = create_typed_fn_tool(
    flush_ipsec_state, "kagent.tools.cilium.FlushIPsecState", "FlushIPsecState"
)


def _list_envoy_config(resource_name: Annotated[Literal["certs", "clusters", "config", "listeners", "logging", "metrics", "serverinfo"], "The name of the Envoy config to list"]) -> str:
    return _run_cilium_dbg_command(f"envoy admin {resource_name}")

list_envoy_config = FunctionTool(
    _list_envoy_config, "List the Envoy configuration", name="list_envoy_config"
)
ListEnvoyConfig, ListEnvoyConfigConfig = create_typed_fn_tool(
    list_envoy_config, "kagent.tools.cilium.ListEnvoyConfig", "ListEnvoyConfig"
)

def _fqdn_cache(command: Annotated[Literal["list", "clean"], "The command to execute on the FQDN cache"]) -> str:
    if command == "clean":
        return _run_cilium_dbg_command("fqdn cache clean -f")
    else:
        return _run_cilium_dbg_command(f"fqdn cache {command}")

fqdn_cache = FunctionTool(
    _fqdn_cache, "Manage the FQDN cache", name="fqdn_cache"
)
FQDNCache, FQDNCacheConfig = create_typed_fn_tool(
    fqdn_cache, "kagent.tools.cilium.FQDNCache", "FQDNCache"
)

def _show_dns_names() -> str:
    return _run_cilium_dbg_command("dns names")

show_dns_names = FunctionTool(
    _show_dns_names, "Show the internal state Cilium has for DNS names/regexes", name="show_dns_names"
)
ShowDNSNames, ShowDNSNamesConfig = create_typed_fn_tool(
    show_dns_names, "kagent.tools.cilium.ShowDNSNames", "ShowDNSNames"
)

def _list_ip_addresses() -> str:
    return _run_cilium_dbg_command("ip list")

list_ip_addresses = FunctionTool(
    _list_ip_addresses, "List the IP addresses in the userspace IPCache", name="list_ip_addresses"
)

ListIPAddresses, ListIPAddressesConfig = create_typed_fn_tool(
    list_ip_addresses, "kagent.tools.cilium.ListIPAddresses", "ListIPAddresses"
)

def _show_ip_cache_information(cidr: Annotated[str, "The CIDR to show information for"], labels: Annotated[Optional[str], "The identity labels"]) -> str:
    if labels:
        return _run_cilium_dbg_command(f"ip get --labels {labels}")
    else:
        return _run_cilium_dbg_command(f"ip get {cidr}")

show_ip_cache_information = FunctionTool(
    _show_ip_cache_information, "Show the information of the IP cache", name="show_ip_cache_information"
)
ShowIPCacheInformation, ShowIPCacheInformationConfig = create_typed_fn_tool(
    show_ip_cache_information, "kagent.tools.cilium.ShowIPCacheInformation", "ShowIPCacheInformation"
)

def _delete_key_from_kvstore(key: Annotated[str, "The key to delete from the kvstore"]) -> str:
    return _run_cilium_dbg_command(f"kvstore delete {key}")

delete_key_from_kvstore = FunctionTool(
    _delete_key_from_kvstore, "Delete a key from the kvstore", name="delete_key_from_kvstore"
)
DeleteKeyFromKVStore, DeleteKeyFromKVStoreConfig = create_typed_fn_tool(
    delete_key_from_kvstore, "kagent.tools.cilium.DeleteKeyFromKVStore", "DeleteKeyFromKVStore"
)

def _get_kvstore_key(key: Annotated[str, "The key to get from the kvstore"]) -> str:
    return _run_cilium_dbg_command(f"kvstore get {key}")

get_kvstore_key = FunctionTool(
    _get_kvstore_key, "Get a key from the kvstore", name="get_kvstore_key"
)
GetKVStoreKey, GetKVStoreKeyConfig = create_typed_fn_tool(
    get_kvstore_key, "kagent.tools.cilium.GetKVStoreKey", "GetKVStoreKey"
)

def _set_kvstore_key(key: Annotated[str, "The key to set in the kvstore"], value: Annotated[str, "The value to set the key to"]) -> str:
    return _run_cilium_dbg_command(f"kvstore set {key}={value}")

set_kvstore_key = FunctionTool(
    _set_kvstore_key, "Set a key in the kvstore", name="set_kvstore_key"
)
SetKVStoreKey, SetKVStoreKeyConfig = create_typed_fn_tool(
    set_kvstore_key, "kagent.tools.cilium.SetKVStoreKey", "SetKVStoreKey"
)


def _show_load_information() -> str:
    return _run_cilium_dbg_command("loadinfo")

show_load_information = FunctionTool(
    _show_load_information, "Show the load information", name="show_load_information"
)
ShowLoadInformation, ShowLoadInformationConfig = create_typed_fn_tool(
    show_load_information, "kagent.tools.cilium.ShowLoadInformation", "ShowLoadInformation"
)

def _list_local_redirect_policies() -> str:
    return _run_cilium_dbg_command("lrp list")

list_local_redirect_policies = FunctionTool(
    _list_local_redirect_policies, "List the local redirect policies", name="list_local_redirect_policies"
)
ListLocalRedirectPolicies, ListLocalRedirectPoliciesConfig = create_typed_fn_tool(
    list_local_redirect_policies, "kagent.tools.cilium.ListLocalRedirectPolicies", "ListLocalRedirectPolicies"
)

def _list_bpf_map_events(map_name: Annotated[str, "The name of the BPF map to show events for"]) -> str:
    return _run_cilium_dbg_command(f"bpf map events {map_name}")

list_bpf_map_events = FunctionTool(
    _list_bpf_map_events, "List the events of the BPF maps", name="list_bpf_map_events"
)
ListBPFMapEvents, ListBPFMapEventsConfig = create_typed_fn_tool(
    list_bpf_map_events, "kagent.tools.cilium.ListBPFMapEvents", "ListBPFMapEvents"
)

def _get_bpf_map(map_name: Annotated[str, "The name of the BPF map to get"]) -> str:
    return _run_cilium_dbg_command(f"bpf map get {map_name}")

get_bpf_map = FunctionTool(
    _get_bpf_map, "Get the BPF map", name="get_bpf_map"
)
GetBPFMap, GetBPFMapConfig = create_typed_fn_tool(
    get_bpf_map, "kagent.tools.cilium.GetBPFMap", "GetBPFMap"
)

def _list_bpf_maps() -> str:
    return _run_cilium_dbg_command("bpf map list")

list_bpf_maps = FunctionTool(
    _list_bpf_maps, "List all open BPF maps", name="list_bpf_maps"
)
ListBPFMaps, ListBPFMapsConfig = create_typed_fn_tool(
    list_bpf_maps, "kagent.tools.cilium.ListBPFMaps", "ListBPFMaps"
)

def _list_metrics(match_pattern: Annotated[Optional[str], "The pattern to match in the metrics"]) -> str:
    if match_pattern:
        return _run_cilium_dbg_command(f"metrics list --p {match_pattern}")
    else:
        return _run_cilium_dbg_command("metrics list")

list_metrics = FunctionTool(
    _list_metrics, "List the metrics", name="list_metrics"
)
ListMetrics, ListMetricsConfig = create_typed_fn_tool(
    list_metrics, "kagent.tools.cilium.ListMetrics", "ListMetrics"
)

def _list_cluster_nodes() -> str:
    return _run_cilium_dbg_command("nodes list")

list_cluster_nodes = FunctionTool(
    _list_cluster_nodes, "List the nodes in the cluster", name="list_cluster_nodes"
)
ListClusterNodes, ListClusterNodesConfig = create_typed_fn_tool(
    list_cluster_nodes, "kagent.tools.cilium.ListClusterNodes", "ListClusterNodes"
)

def _list_node_ids() -> str:
    return _run_cilium_dbg_command("nodeid list")

list_node_ids = FunctionTool(
    _list_node_ids, "List the node IDs and the associated IP addresses", name="list_node_ids"
)
ListNodeIds, ListNodeIdsConfig = create_typed_fn_tool(
    list_node_ids, "kagent.tools.cilium.ListNodeIds", "ListNodeIds"
)

def _display_policy_node_information(labels: Annotated[Optional[str], "The labels to display information for"]) -> str:
    if labels:
        return _run_cilium_dbg_command(f"policy get {labels}")
    else:
        return _run_cilium_dbg_command("policy get")

display_policy_node_information = FunctionTool(
    _display_policy_node_information, "Display the policy node information", name="display_policy_node_information"
)
DisplayPolicyNodeInformation, DisplayPolicyNodeInformationConfig = create_typed_fn_tool(
    display_policy_node_information, "kagent.tools.cilium.DisplayPolicyNodeInformation", "DisplayPolicyNodeInformation"
)

def _delete_policy_rules(labels: Annotated[Optional[str], "The labels to delete the policy rules for"], all: Annotated[bool, "Whether to delete all policy rules"] = False) -> str:
    if all:
        return _run_cilium_dbg_command("policy delete --all")
    else:
        return _run_cilium_dbg_command(f"policy delete {labels}")

delete_policy_rules = FunctionTool(
    _delete_policy_rules, "Delete the policy rules", name="delete_policy_rules"
)
DeletePolicyRules, DeletePolicyRulesConfig = create_typed_fn_tool(
    delete_policy_rules, "kagent.tools.cilium.DeletePolicyRules", "DeletePolicyRules"
)

def _display_selectors() -> str:
    return _run_cilium_dbg_command("policy selectors")

display_selectors = FunctionTool(
    _display_selectors, "Display cached information about selectors", name="display_selectors"
)
DisplaySelectors, DisplaySelectorsConfig = create_typed_fn_tool(
    display_selectors, "kagent.tools.cilium.DisplaySelectors", "DisplaySelectors"
)

def _list_xdp_cidr_filters() -> str:
    return _run_cilium_dbg_command("prefilter list")

list_xdp_cidr_filters = FunctionTool(
    _list_xdp_cidr_filters, "List the XDP CIDR filters (prefilter)", name="list_xdp_cidr_filters"
)
ListXDPCIDRFilters, ListXDPCIDRFiltersConfig = create_typed_fn_tool(
    list_xdp_cidr_filters, "kagent.tools.cilium.ListXDPCIDRFilters", "ListXDPCIDRFilters"
)

def _update_xdp_cidr_filters(cidr_prefixes: Annotated[list[str], "The list of CIDR prefixes to block"], revision: Annotated[Optional[int], "The update revision"]) -> str:
    return _run_cilium_dbg_command(f"prefilter update --cidr {' '.join(cidr_prefixes)} --revision {revision}")

update_xdp_cidr_filters = FunctionTool(
    _update_xdp_cidr_filters, "Update the XDP CIDR filters", name="update_xdp_cidr_filters"
)
UpdateXDPCIDRFilters, UpdateXDPCIDRFiltersConfig = create_typed_fn_tool(
    update_xdp_cidr_filters, "kagent.tools.cilium.UpdateXDPCIDRFilters", "UpdateXDPCIDRFilters"
)

def _delete_xdp_cidr_filters(cidr_prefixes: Annotated[list[str], "The list of CIDR prefixes to delete   "], revision: Annotated[Optional[int], "The update revision"]) -> str:
    return _run_cilium_dbg_command(f"prefilter delete --cidr {' '.join(cidr_prefixes)} --revision {revision}")

delete_xdp_cidr_filters = FunctionTool(
    _delete_xdp_cidr_filters, "Delete the XDP CIDR filters", name="delete_xdp_cidr_filters"
)
DeleteXDPCIDRFilters, DeleteXDPCIDRFiltersConfig = create_typed_fn_tool(
    delete_xdp_cidr_filters, "kagent.tools.cilium.DeleteXDPCIDRFilters", "DeleteXDPCIDRFilters"
)


def _validate_cilium_network_policies(enable_k8s: Annotated[bool, "Enable the k8s clientset"] = True, enable_k8s_api_discovery: Annotated[bool, "Enable discovery of Kubernetes API groups and resources with the discovery API"] = True) -> str:
    return _run_cilium_dbg_command(f"preflight validate-cnp {'--enable-k8s' if enable_k8s else ''} {'--enable-k8s-api-discovery' if enable_k8s_api_discovery else ''}")

validate_cilium_network_policies = FunctionTool(
    _validate_cilium_network_policies, "Validate the Cilium network policies. It's recommended to run this before upgrading Cilium to ensure all policies are valid.", name="validate_cilium_network_policies"
)
ValidateCiliumNetworkPolicies, ValidateCiliumNetworkPoliciesConfig = create_typed_fn_tool(
    validate_cilium_network_policies, "kagent.tools.cilium.ValidateCiliumNetworkPolicies", "ValidateCiliumNetworkPolicies"
)


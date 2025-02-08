from enum import Enum
from typing import Annotated

from autogen_core.models import SystemMessage, UserMessage
from autogen_core.tools import FunctionTool
from autogen_ext.models.openai import OpenAIChatCompletionClient

from .prompts import (
    AUTH_POLICY_PROMPT,
    GATEWAY_PROMPT,
    PEER_AUTHENTICATION_PROMPT,
    VIRTUAL_SERVICE_PROMPT,
)
from .prompts.base import IstioResources


def get_model_client():
    # TODO: We should have a way to configure externally somehow.
    return OpenAIChatCompletionClient(
        model="gpt-4o-mini",
    )


async def _generate_crd(system_prompt: str, policy_description: str) -> str:
    """
    Asynchronously generates a Custom Resource Definition (CRD) based on the provided system prompt and policy description.

    Args:
      system_prompt (str): The system prompt to guide the CRD generation.
      policy_description (str): The description of the policy to be included in the CRD.

    Returns:
      str: The generated CRD content or an error message if the generation fails.

    Raises:
      Exception: If there is an error during the CRD generation process.
    """
    try:
        model_client = get_model_client()
        result = await model_client.create(
            messages=[SystemMessage(content=system_prompt), UserMessage(content=policy_description, source="user")],
            json_output=True,
        )
        return result.content
    except Exception as e:
        return f"Error generating policy: {str(e)}"


async def _generate_gateway_crd(
    policy_description: Annotated[str, "Detailed description of the Gateway to generate YAML for"],
) -> str:
    return await _generate_crd(GATEWAY_PROMPT, policy_description)


async def _generate_auth_policy_crd(
    policy_description: Annotated[str, "Detailed description of the AuthorizationPolicy to generate YAML for"],
) -> str:
    return await _generate_crd(AUTH_POLICY_PROMPT, policy_description)


async def _generate_peer_auth_crd(
    policy_description: Annotated[str, "Detailed description of the PeerAuthentication to generate YAML for"],
) -> str:
    return await _generate_crd(PEER_AUTHENTICATION_PROMPT, policy_description)


async def _generate_virtual_service_crd(
    policy_description: Annotated[str, "Detailed description of the VirtualService to generate YAML for"],
) -> str:
    return await _generate_crd(VIRTUAL_SERVICE_PROMPT, policy_description)

async def _generate_istio_resource(
    istio_resource: Annotated[IstioResources, "Type of resources to generate"],
    policy_description: Annotated[str, "Detailed description of the resource to generate YAML for"],
) -> str:
    if istio_resource == IstioResources.AUTH_POLICY:
        return await _generate_auth_policy_crd(policy_description)
    elif istio_resource == IstioResources.GATEWAY:
        return await _generate_gateway_crd(policy_description)
    elif istio_resource == IstioResources.PEER_AUTHENTICATION:
        return await _generate_peer_auth_crd(policy_description)
    elif istio_resource == IstioResources.VIRTUAL_SERVICE:
        return await _generate_virtual_service_crd(policy_description)
    else:
        return "Unsupported Istio resource type"


generate_resource = FunctionTool(
    _generate_istio_resource,
    description="Generates an Istio resource YAML configuration from a detailed description",
    name="generate_istio_resource",
)

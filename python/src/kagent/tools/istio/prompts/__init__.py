from .auth_policy import AUTH_POLICY_PROMPT
from .base import IstioResources
from .gateway import GATEWAY_PROMPT
from .peer_auth import PEER_AUTHENTICATION_PROMPT
from .virtual_service import VIRTUAL_SERVICE_PROMPT

__all__ = [
    "AUTH_POLICY_PROMPT",
    "GATEWAY_PROMPT",
    "PEER_AUTHENTICATION_PROMPT",
    "VIRTUAL_SERVICE_PROMPT",
    "IstioResources"
]

from ._istio_crds import generate_resource
from ._istioctl import (
    ProxyConfig,
    VerifyInstall,
)

__all__ = ["ProxyConfig", "VerifyInstall", "generate_resource"]

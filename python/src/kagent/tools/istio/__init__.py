from ._istio_crds import generate_resource
from ._istioctl import install, proxy_config, uninstall, verify_install

__all__ = ["verify_install", "proxy_config", "generate_resource", "install", "uninstall"]

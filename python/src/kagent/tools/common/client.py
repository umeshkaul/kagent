from datetime import datetime
from typing import Annotated, Optional, Union

import httpx
from pydantic import Field


class HttpClient:
    """HTTP client for making API requests"""

    base_url: str = Field(default="http://localhost:9090/api/v1")
    api_key: str = Field(default="")
    session: Optional[Annotated[httpx.AsyncClient, Field(default=None, exclude=True)]] = None

    def __init__(self, base_url: str = "http://localhost:9090/api/v1", api_key: str = ""):
        self.base_url = base_url
        self.api_key = api_key

    async def __aenter__(self):
        if self.session is None:
            self.session = httpx.AsyncClient()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.aclose()
            self.session = None

    def _format_time(self, time_value: Optional[Union[datetime, float]]) -> Optional[str]:
        if time_value is None:
            return None
        if isinstance(time_value, datetime):
            return str(time_value.timestamp())
        return str(time_value)

    async def _make_request(
        self, method: str, endpoint: str, params: Optional[dict] = None, data: Optional[dict] = None
    ) -> dict:
        if not self.session:
            self.session = httpx.AsyncClient()

        url = f"{self.base_url}/{endpoint.lstrip('/')}"

        # Remove None values from params
        if params:
            params = {k: v for k, v in params.items() if v is not None}

        # Add the api key if needed
        headers = {}
        if self.api_key:
            headers = {
                "authorization": f"Bearer {self.api_key}"
            }

        try:
            response = await self.session.request(method, url, params=params, json=data, headers=headers)

            response_data = response.json()

            if response.status_code >= 400:
                error_msg = response_data.get("error", "Unknown error")
                raise Exception(f"HTTP Client API error: {error_msg}")

            return response_data
        except httpx.RequestError as e:
            raise Exception(f"Failed to connect: {str(e)}") from e

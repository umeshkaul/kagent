from datetime import datetime
from typing import Any, Optional, Union

import aiohttp
from pydantic_core import CoreSchema, core_schema


class PrometheusClient:
    def __init__(self, base_url: str = "http://localhost:9090/api/v1"):
        self.base_url = base_url.rstrip("/")
        self.session = None

    @classmethod
    def __get_pydantic_core_schema__(
        cls,
        _source_type: Any,
        _handler: Any,
    ) -> CoreSchema:
        return core_schema.json_or_python_schema(
            json_schema=core_schema.str_schema(),
            python_schema=core_schema.union_schema([
                core_schema.is_instance_schema(cls),
                core_schema.str_schema(),
            ]),
            serialization=core_schema.plain_serializer_function_ser_schema(
                lambda instance: instance.base_url
            ),
        )

    async def __aenter__(self):
        if self.session is None:
            self.session = aiohttp.ClientSession()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            if not self.session.closed:
                await self.session.close()
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
            self.session = aiohttp.ClientSession()

        url = f"{self.base_url}/{endpoint.lstrip('/')}"

        # Remove None values from params
        if params:
            params = {k: v for k, v in params.items() if v is not None}

        try:
            async with self.session.request(method, url, params=params, json=data) as response:
                response_data = await response.json()

                if response.status >= 400:
                    error_msg = response_data.get("error", "Unknown error")
                    raise Exception(f"Prometheus API error: {error_msg}")

                return response_data
        except aiohttp.ClientError as e:
            raise Exception(f"Failed to connect to Prometheus: {str(e)}") from e


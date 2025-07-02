from typing import Any, Dict

from autogen_core import (
    ComponentModel,
)
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel

from ...datamodel import Tool, ToolServer
from ...toolservermanager import ToolServerManager

router = APIRouter()


class GetServerToolsRequest(BaseModel):
    server: ComponentModel


class GetServerToolsResponse(BaseModel):
    tools: Dict[str, Dict]


@router.post("/")
async def get_server_tools(
    request: GetServerToolsRequest,
) -> GetServerToolsResponse:
    # First check if server exists

    tsm = ToolServerManager()
    tools_dict: Dict[str, Dict] = {}
    try:
        tools = await tsm.discover_tools(request.server)
        for tool in tools:
            # Generate a unique identifier for the tool from its component
            component_data = tool.dump_component().model_dump_json()

            # Check if the tool already exists based on id/name
            component_config = component_data.get("config", {})
            tool_config = component_config.get("tool", {})
            tool_name = tool_config.get("name", None)
            tools_dict[tool_name] = tool

    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Failed to get server tools: {str(e)}") from e

    return GetServerToolsResponse(tools=tools_dict)


@router.post("/{server_id}/refresh")
async def refresh_server_tools(server_id: int, user_id: str, db=Depends(get_db)) -> RefreshServerToolsResponse:
    """Refresh tools for an existing server"""

    server_response = db.get(ToolServer, filters={"id": server_id, "user_id": user_id})
    if not server_response.status or not server_response.data:
        raise HTTPException(status_code=404, detail="Server not found")

    server = server_response.data[0]
    tsm = ToolServerManager()

    try:
        # Use the same discovery logic as the tools endpoint
        tools_components = await tsm.discover_tools(server.component)

        # Update server last_connected timestamp
        from datetime import datetime

        server.last_connected = datetime.now()
        db.upsert(server)

        updated_count = 0
        created_count = 0

        for tool_component in tools_components:
            # Generate a unique identifier for the tool from its component
            component_data = tool_component.dump_component().model_dump()

            # Check if the tool already exists based on id/name
            component_config = component_data.get("config", {})
            tool_config = component_config.get("tool", {})
            tool_name = tool_config.get("name", None)

            # First get all tools for this server and user
            existing_tool_response = db.get(Tool, filters={"server_id": server_id, "user_id": user_id})

            matching_tools = []
            if existing_tool_response.status and existing_tool_response.data:
                for tool in existing_tool_response.data:
                    try:
                        tool_comp = tool.component
                        if tool_comp.get("config", {}).get("tool", {}).get("name") == tool_name:
                            matching_tools.append(tool)
                    except Exception:
                        pass

            # Update existing_tool_response to use our filtered results
            existing_tool_response.data = matching_tools

            if existing_tool_response.status and existing_tool_response.data:
                # Tool exists, update it
                existing_tool = existing_tool_response.data[0]
                existing_tool.component = component_data
                db.upsert(existing_tool)
                updated_count += 1
            else:
                # Tool does not exist, create new
                new_tool = Tool(user_id=user_id, server_id=server_id, component=component_data)
                # print(f"Creating new tool: {new_tool}")
                db.upsert(new_tool)
                created_count += 1

        return RefreshServerToolsResponse(tools={tool.name: tool.component for tool in tools_components})
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Failed to refresh server: {str(e)}") from e

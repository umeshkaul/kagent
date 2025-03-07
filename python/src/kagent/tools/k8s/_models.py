from pydantic import BaseModel, Field

from ._resource_types import ResourceTypes


class GenerateResourceToolConfig(BaseModel):
    model: str = Field(default="gpt-4o-mini", description="The OpenAI model to use for generating the CRD.")
    openai_api_key: str = Field(
        default="",
        description="API key for OpenAI services. If empty, the environment variable 'OPENAI_API_KEY' will be used.",
    )


class GenerateResourceToolInput(BaseModel):
    resource_description: str = Field(description="Detailed description of the resource to generate YAML for")
    resource_type: ResourceTypes = Field(description="Type of resource to generate")

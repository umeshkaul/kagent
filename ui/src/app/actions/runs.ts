"use server";

import { BaseResponse, CreateRunRequest, CreateRunResponse } from "@/lib/types";
import { fetchApi } from "./utils";

export async function createRun(payload: CreateRunRequest): Promise<BaseResponse<CreateRunResponse>> {
  try {
    const response = await fetchApi<CreateRunResponse>(`/runs`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response) {
      throw new Error("Failed to create run");
    }

    return { success: true, data: response };
  } catch (error) {
    console.error("Error creating run:", error);
    return { success: false, error: "Failed to create run. Please try again." };
  }
}

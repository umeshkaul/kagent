"use server";

import { BaseResponse } from "@/lib/types";
import { Team } from "@/types/datamodel";
import { revalidatePath } from "next/cache";
import { fetchApi } from "./utils";

export async function getTeam(teamId: string): Promise<BaseResponse<Team>> {
  try {
    const data = await fetchApi<Team>(`/teams/${teamId}`);
    return { success: true, data };
  } catch (error) {
    console.error("Error getting team:", error);
    return { success: false, error: "Failed to get team. Please try again." };
  }
}

export async function deleteTeam(teamId: string) {
  try {
    await fetchApi(`/teams/${teamId}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
    });

    revalidatePath("/");
    return { success: true };
  } catch (error) {
    console.error("Error deleting team:", error);
    return { success: false, error: "Failed to delete team. Please try again." };
  }
}

export async function createTeam(teamConfig: Team): Promise<BaseResponse<Team>> {
  try {
    const response = await fetchApi<Team>(`/teams`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(teamConfig),
    });

    if (!response) {
      throw new Error("Failed to create team");
    }

    revalidatePath(`/agents/${response.id}/chat`);
    return { success: true, data: response };
  } catch (error) {
    console.error("Error creating team:", error);
    return { success: false, error: "Failed to create team. Please try again." };
  }
}

export async function getTeams(): Promise<BaseResponse<Team[]>> {
  try {
    const data = await fetchApi<Team[]>(`/teams`);
    return { success: true, data };
  } catch (error) {
    console.error("Error getting teams:", error);
    return { success: false, error: `Failed to get teams. Please try again. ${error}` };
  }
}

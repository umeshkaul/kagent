"use client";

import * as React from "react";
import { ChevronsUpDown, Plus } from "lucide-react";

import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuShortcut, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from "@/components/ui/sidebar";
import { useEffect, useState } from "react";
import { getTeams } from "@/app/actions/teams";
import { Agent } from "@/types/datamodel";
import KagentLogo from "../kagent-logo";
import { useRouter } from "next/navigation";
import { getChatData } from "@/app/actions/chat";

interface AgentSwitcherProps {
  agentId: number;
}

export function AgentSwitcher({ agentId }: AgentSwitcherProps) {
  const router = useRouter();
  const { isMobile } = useSidebar();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<Agent | null>(null);

  useEffect(() => {
    const fetchSessions = async () => {
      const data = await getChatData(agentId, null);
      if (data.agent) {
        setSelectedTeam(data.agent);
      }
    };
    fetchSessions();
  }, [agentId]);

  useEffect(() => {
    const fetchTeams = async () => {
      const teamData = await getTeams();

      if (teamData && teamData.data) {
        setAgents(teamData.data);
      }
    };

    fetchTeams();
  }, []);

  if (!selectedTeam) {
    return null;
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton size="lg" className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary/5 text-sidebar-primary-foreground">
                <KagentLogo className="w-4 h-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">{selectedTeam.metadata.name}</span>
                <span className="truncate text-xs">{selectedTeam.spec.modelConfigRef}</span>
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg" align="start" side={isMobile ? "bottom" : "right"} sideOffset={4}>
            <DropdownMenuLabel className="text-xs text-muted-foreground">Agents</DropdownMenuLabel>
            {agents.map((agent, index) => {
              return (
                <DropdownMenuItem
                  key={agent.metadata.name}
                  onClick={() => {
                    router.push(`/agents/${agent.metadata.name}/chat`);
                  }}
                  className="gap-2 p-2"
                >
                  {agent.metadata.name}
                  <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
                </DropdownMenuItem>
              );
            })}
            <DropdownMenuSeparator />
            <DropdownMenuItem className="gap-2 p-2" onClick={() => router.push("/agents/new")}>
              <div className="flex size-6 items-center justify-center rounded-md border bg-background">
                <Plus className="size-4" />
              </div>
              <div className="font-medium text-muted-foreground">New agent</div>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

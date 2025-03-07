"use client";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FunctionSquare, Clipboard } from "lucide-react";
import type { Team, AssistantAgentConfig, ToolConfig, Component } from "@/types/datamodel";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useState } from "react";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog";
import { Textarea } from "@/components/ui/textarea";
import { getToolDescription, getToolDisplayName, getToolIdentifier, isMcpTool } from "@/lib/data";
import { createTeam } from "@/app/actions/teams";
import KagentLogo from "../kagent-logo";
import { findAllAssistantAgents, updateUsersAgent } from "@/lib/agents";
import { SidebarHeader, Sidebar, SidebarContent } from "../ui/sidebar";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";

interface AgentDetailsSidebarProps {
  selectedTeam: Team | null;
}

export function AgentDetailsSidebar({ selectedTeam }: AgentDetailsSidebarProps) {
  const [isSystemPromptOpen, setIsSystemPromptOpen] = useState(false);
  const [currentSystemPrompt, setCurrentSystemPrompt] = useState("");
  const [editedSystemPrompt, setEditedSystemPrompt] = useState("");
  const [isPromptEdited, setIsPromptEdited] = useState(false);
  const [isConfirmationDialogOpen, setIsConfirmationDialogOpen] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);

  // Keep track of the current agent for updating
  const [currentAgentIndex, setCurrentAgentIndex] = useState<number | null>(null);

  const renderAgentTools = (tools: Component<ToolConfig>[] = []) => {
    if (tools.length === 0) {
      return <div className="text-sm italic">No tools available</div>;
    }

    return (
      <ul className="mt-4 flex flex-col gap-2">
        {tools.map((tool) => {
          const toolIdentifier = getToolIdentifier(tool);
          const displayName = getToolDisplayName(tool);
          const displayDescription = getToolDescription(tool);

          return (
            <li key={toolIdentifier} className="text-sm">
              <Tooltip>
                <TooltipTrigger>{displayName}</TooltipTrigger>
                <TooltipContent>
                  <p>{displayDescription}</p>
                </TooltipContent>
              </Tooltip>
              {isMcpTool(tool) && <span className="ml-2 text-xs bg-blue-400/20 text-blue-400 px-2 py-0.5 rounded">MCP</span>}
            </li>
          );
        })}
      </ul>
    );
  };

  const handleOpenSystemPrompt = (systemMessage: string, agentIndex: number) => {
    setCurrentSystemPrompt(systemMessage);
    setEditedSystemPrompt(systemMessage);
    setIsPromptEdited(false);
    setCurrentAgentIndex(agentIndex);
    setIsSystemPromptOpen(true);
  };

  const handleSystemPromptChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setEditedSystemPrompt(e.target.value);
    setIsPromptEdited(e.target.value !== currentSystemPrompt);
  };

  const handleCloseSystemPrompt = () => {
    if (isPromptEdited) {
      setIsConfirmationDialogOpen(true);
    } else {
      setIsSystemPromptOpen(false);
    }
  };

  const handleDiscardChanges = () => {
    setIsConfirmationDialogOpen(false);
    setIsSystemPromptOpen(false);
    setIsPromptEdited(false);
  };

  const handleContinueEditing = () => {
    setIsConfirmationDialogOpen(false);
  };

  const handleUpdateSystemPrompt = async () => {
    if (!selectedTeam || currentAgentIndex === null) return;

    setIsUpdating(true);
    try {
      const updatedTeam = updateUsersAgent(selectedTeam, (agent) => {
        agent.config.system_message = editedSystemPrompt;
      });

      // Remove the created_at and updated_at variables from the selectedTeam - backend fails as the date/time is not in the correct format (or something)
      // in any case, the backend updates the created_at and updated_at fields automatically anyway
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { created_at, updated_at, ...editedTeam } = updatedTeam;
      await createTeam(editedTeam);

      setCurrentSystemPrompt(editedSystemPrompt);
      setIsPromptEdited(true);
      setIsUpdating(false);
      // Close the dialog
      setIsSystemPromptOpen(false);
    } catch (error) {
      console.error("Failed to update instructions:", error);
    } finally {
      setIsUpdating(false);
    }
  };

  const assistantAgents = findAllAssistantAgents(selectedTeam?.component);
  return (
    <>
      <Sidebar side={"right"} collapsible="offcanvas">
        <SidebarHeader>Agent Details</SidebarHeader>
        <SidebarContent className="">
          <ScrollArea className="flex-1 px-6 py-6">
            {assistantAgents.map((participant, index) => {
              const assistantAgent = participant.config as AssistantAgentConfig;
              return (
                <div key={index} className="text-s">
                  <div className="flex items-start  flex-col space-y-2">
                    <div className="flex items-center justify-between w-full">
                      <div className="inline-flex justify-center items-center gap-2">
                        <KagentLogo className="h-5 w-5" />
                        <h5 className="font-semibold text-base">{participant.label}</h5>
                      </div>
                      <div>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button variant="link" size="sm" onClick={async () => await navigator.clipboard.writeText(JSON.stringify(selectedTeam))} className="px-0 ">
                                <Clipboard className="h-4 w-4 mr-2" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>
                              <p>Copy JSON representation</p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </div>
                    </div>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Badge variant="outline" className="text-xs">
                            {assistantAgent.model_client.config.model}
                          </Badge>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>The model agent is using</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                  <p className="mt-4 text-muted-foreground">{assistantAgent.description}</p>
                  <Button variant="link" size="sm" onClick={() => handleOpenSystemPrompt(assistantAgent.system_message ?? "No system prompt available", index)} className="px-0">
                    View instructions &rarr;
                  </Button>

                  <div className="mt-8">
                    <h6 className="text-base font-medium flex items-center gap-2 mb-4">
                      <FunctionSquare className="h-6 w-6" />
                      Available Tools
                    </h6>
                    {renderAgentTools(assistantAgent.tools)}
                  </div>
                </div>
              );
            })}
          </ScrollArea>
        </SidebarContent>
      </Sidebar>

      {/* Agent Instructions Dialog */}
      <Dialog open={isSystemPromptOpen} onOpenChange={handleCloseSystemPrompt}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-secondary-foreground">Agent Instructions</DialogTitle>
            <DialogDescription>You can update the instructions for the agent to follow while interacting with the user</DialogDescription>
          </DialogHeader>

          <Textarea value={editedSystemPrompt} onChange={handleSystemPromptChange} className="min-h-[60vh] font-mono text-sm text-secondary-foreground" placeholder="Enter agent instructions..." />

          <DialogFooter>
            <Button variant="default" onClick={handleUpdateSystemPrompt} disabled={!isPromptEdited || isUpdating} className="">
              {isUpdating ? "Updating..." : "Update"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Confirmation Dialog for Unsaved Changes */}
      <AlertDialog open={isConfirmationDialogOpen} onOpenChange={setIsConfirmationDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="text-secondary-foreground">Unsaved Changes</AlertDialogTitle>
            <AlertDialogDescription>You have unsaved changes to the instructions. Do you want to discard these changes?</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleContinueEditing} className="text-secondary-foreground">
              Continue Editing
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleDiscardChanges} className="bg-red-600 hover:bg-red-700">
              Discard Changes
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

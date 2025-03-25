import type { Agent } from "@/types/datamodel";
import { AgentCard } from "./AgentCard";

interface AgentGridProps {
  agents: Agent[];
}

export function AgentGrid({ agents }: AgentGridProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {agents.map((agent) => (
        <AgentCard key={agent.metadata.name} agent={agent} />
      ))}
    </div>
  );
}
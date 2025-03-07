import { SessionWithRuns } from "@/types/datamodel";
import RunItem from "./RunItem";
import { SidebarGroup, SidebarGroupContent, SidebarGroupLabel } from "../ui/sidebar";

interface SessionGroupProps {
  title: string;
  sessions: SessionWithRuns[];
  onDeleteSession: (sessionId: number) => Promise<void>;
  agentId?: number;
}
const SessionGroup = ({ title, sessions, onDeleteSession, agentId }: SessionGroupProps) => (
  <SidebarGroup>
    <SidebarGroupLabel>{title}</SidebarGroupLabel>
    <SidebarGroupContent>
      {sessions.map((sessionWithRuns) => (
        <div key={sessionWithRuns.session.id} className="py-2.5">
          {sessionWithRuns.runs.map((run) => (
            <RunItem key={run.id} sessionId={sessionWithRuns.session.id!} agentId={agentId} run={run} onDelete={onDeleteSession} />
          ))}
        </div>
      ))}
    </SidebarGroupContent>
  </SidebarGroup>
);

export default SessionGroup;

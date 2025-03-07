"use client";
import { ReactNode, use, useEffect, useState } from "react";
import { Run, Session, SessionWithRuns, Team } from "@/types/datamodel";
import SessionsSidebar from "@/components/sidebars/SessionsSidebar";
import { AgentDetailsSidebar } from "@/components/sidebars/AgentDetailsSidebar";
import { SidebarProvider } from "@/components/ui/sidebar";
import { getChatData } from "@/app/actions/chat";
import { LoadingState } from "@/components/LoadingState";

export default function ChatLayout({ params, children }: { params: Promise<{ id: string }>; children: ReactNode }) {
  const { id } = use(params);
  const [chatData, setChatData] = useState<{
    agent: Team;
    sessions: SessionWithRuns[];
    viewState: {
      session: Session;
      run: Run;
    } | null;
  }>();

  useEffect(() => {
    const fetchData = async () => {
      const data = await getChatData(id, null);
      if (data.agent) {
        setChatData(
          data as {
            agent: Team;
            sessions: SessionWithRuns[];
            viewState: {
              session: Session;
              run: Run;
            } | null;
          }
        );
      }
    };
    fetchData();
  }, [id]);

  if (!chatData) {
    return <LoadingState />;
  }

  return (
    <SidebarProvider>
      <SessionsSidebar agentId={id} />
      <main className="w-full max-w-6xl mx-auto">{children}</main>
      <AgentDetailsSidebar selectedTeam={chatData.agent} />
    </SidebarProvider>
  );
}

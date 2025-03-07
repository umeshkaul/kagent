"use client";
import { use, useEffect, useState } from "react";
import { getChatData } from "@/app/actions/chat";
import { LoadingState } from "@/components/LoadingState";
import { Team, SessionWithRuns, Run, Session } from "@/types/datamodel";
import { useRouter } from "next/navigation";
import useChatStore from "@/lib/useChatStore";
import ChatInterface from "@/components/chat/ChatInterface";

export default function ChatPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { initializeNewChat } = useChatStore();

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

  const onNewSession = async () => {
    try {
      await initializeNewChat(parseInt(id));
      const { session } = useChatStore.getState();

      if (session?.id) {
        router.push(`/agents/${id}/chat/${session.id}`);
      }
    } catch (error) {
      console.error("Error creating new session:", error);
    }
  };

  if (!chatData) {
    return <LoadingState />;
  }

  return <ChatInterface selectedAgentTeam={chatData.agent} onNewSession={onNewSession} selectedRun={chatData.viewState?.run} selectedSession={chatData.viewState?.session} />;
}

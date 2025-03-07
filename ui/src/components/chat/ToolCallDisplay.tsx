import React, { useState, useEffect } from "react";
import { messageUtils } from "@/lib/utils";
import { Message, Run } from "@/types/datamodel";
import ToolDisplay from "@/components/ToolDisplay";

interface ToolCallDisplayProps {
  currentMessage: Message;
  currentRun: Run | null;
}

const ToolCallDisplay = ({ currentMessage, currentRun }: ToolCallDisplayProps) => {
  // Track tool calls and their results
  const [toolState, setToolState] = useState(new Map());

  useEffect(() => {
    const newToolState = new Map();

    if (!currentRun || !currentRun?.messages) return;

    // Get all messages from the same source as the current message
    const sourceMessages = currentRun.messages.filter((msg) => msg.config.source === currentMessage.config.source);

    // First pass: collect all tool calls from this source
    sourceMessages.forEach((message) => {
      if (messageUtils.isToolCallContent(message.config.content)) {
        // Only process tool calls from the current message
        if (message === currentMessage) {
          message.config.content.forEach((call) => {
            newToolState.set(call.id, {
              call,
              result: undefined,
            });
          });
        }
      }
    });

    // Second pass: match results with calls
    sourceMessages.forEach((message) => {
      if (messageUtils.isFunctionExecutionResult(message.config.content)) {
        message.config.content.forEach((result) => {
          if (newToolState.has(result.call_id)) {
            newToolState.set(result.call_id, {
              ...newToolState.get(result.call_id),
              result,
            });
          }
        });
      }
    });

    setToolState(newToolState);
  }, [currentRun, currentRun?.messages, currentMessage]);

  if (!toolState.size) return null;

  return (
    <div className="space-y-2">
      {Array.from(toolState.values()).map(({ call, result }) => (
        <ToolDisplay key={call.id} call={call} result={result} />
      ))}
    </div>
  );
};

export default ToolCallDisplay;

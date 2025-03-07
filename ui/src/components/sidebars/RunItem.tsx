import { Run, RunStatus } from "@/types/datamodel";
import {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogCancel,
  AlertDialogAction,
  AlertDialogHeader,
  AlertDialogFooter,
} from "@/components/ui/alert-dialog";
import { AlertCircle, AlertTriangle, CheckCircle, Loader2, MessageSquare, StopCircle, Trash2 } from "lucide-react";
import { Button } from "../ui/button";
import { getRelativeTimeString } from "@/lib/utils";
import { useRouter } from "next/navigation";

interface StatusIconProps {
  status: RunStatus;
}

const StatusIcon = ({ status }: StatusIconProps) => {
  switch (status) {
    case "complete":
      return <CheckCircle className="h-4 w-4" />;
    case "active":
    case "created":
      return <Loader2 className="h-4 w-4 animate-spin" />;
    case "stopped":
      return <StopCircle className="h-4 w-4" />;
    case "awaiting_input":
      return <MessageSquare className="h-4 w-4" />;
    case "error":
    case "timeout":
      return <AlertTriangle className="h-4 w-4" />;
    default:
      return <AlertCircle className="h-4 w-4" />;
  }
};

interface RunItemProps {
  sessionId: number;
  run: Run;
  onDelete: (sessionId: number, runId: string) => Promise<void>;
  agentId?: number;
}

const RunItem = ({ sessionId, run, agentId, onDelete }: RunItemProps) => {
  const router = useRouter();

  const onViewRun = () => {
    router.push(`/agents/${agentId}/chat/${sessionId}`);
  }

  return (
    <div className="group relative">
      <Button onClick={onViewRun} variant="ghost" className="w-full justify-start gap-2 py-4">
        <StatusIcon status={run.status} />
        <div className="flex flex-col items-start gap-1 min-w-0">
          <span className="truncate max-w-[160px] text-sm ">{String(run.task?.content || "(new chat)")}</span>
          <div className="inline-flex gap-2 items-center">
            <span className="text-xs">{getRelativeTimeString(run.created_at)}</span>
          </div>
        </div>
      </Button>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-transparent"
            onClick={(e) => {
              e.stopPropagation();
            }}
          >
            <Trash2 className="h-4 w-4 text-red-400 hover:text-red-300" />
          </Button>
        </AlertDialogTrigger>

        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="text-secondary-foreground">Delete Chat</AlertDialogTitle>
            <AlertDialogDescription>Are you sure you want to delete this chat? This action cannot be undone.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="text-secondary-foreground">Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={async () => onDelete(sessionId, run.id)} className="bg-red-500 hover:bg-red-600">
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default RunItem;

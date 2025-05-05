package cli

import (
	"context"
	"fmt"
	"github.com/abiosoft/ishell/v2"
	"github.com/google/uuid"
	"github.com/kagent-dev/kagent/go/cli/internal/config"
	"github.com/kagent-dev/kagent/go/controller/utils/a2autils"
	"github.com/spf13/pflag"
	"time"
	"trpc.group/trpc-go/trpc-a2a-go/client"
	"trpc.group/trpc-go/trpc-a2a-go/protocol"
)

func A2ACmd(ctx context.Context) *ishell.Cmd {
	a2aCmd := &ishell.Cmd{
		Name: "a2a",
		Help: "Interact with an Agent over the A2A protocol.",
	}
	a2aCmd.AddCmd(&ishell.Cmd{
		Name: "run",
		Help: "Run a task with an agent using the A2A protocol.",
		LongHelp: `Run a task with an agent using the A2A protocol.
The task is sent to the agent, and the result is printed to the console.

Example:
a2a run [--namespace <agent-namespace>] <agent-name> <task>
`,
		Func: func(c *ishell.Context) {
			if len(c.RawArgs) < 4 {
				c.Println("Usage: a2a run [--namespace <agent-namespace>] <agent-name> <task>")
				return
			}
			flagSet := pflag.NewFlagSet(c.RawArgs[0], pflag.ContinueOnError)
			agentNamespace := flagSet.String("namespace", "kagent", "Agent namespace")
			timeout := flagSet.Duration("timeout", 300*time.Second, "Timeout for the task")
			if err := flagSet.Parse(c.Args); err != nil {
				c.Printf("Failed to parse flags: %v\n", err)
				return
			}
			agentName := flagSet.Arg(0)
			prompt := flagSet.Arg(1)

			result, err := runTask(ctx, *agentNamespace, agentName, prompt, *timeout)
			if err != nil {
				c.Err(err)
				return
			}

			switch result.Status.State {
			case protocol.TaskStateUnknown:
				c.Println("Task state is unknown.")
				if result.Status.Message != nil {
					c.Println("Message:", a2autils.ExtractText(*result.Status.Message))
				} else {
					c.Println("No message provided.")
				}
			case protocol.TaskStateCanceled:
				c.Println("Task was canceled.")
			case protocol.TaskStateFailed:
				c.Println("Task failed.")
				if result.Status.Message != nil {
					c.Println("Error:", a2autils.ExtractText(*result.Status.Message))
				} else {
					c.Println("No error message provided.")
				}
			case protocol.TaskStateCompleted:
				c.Println("Task completed successfully:")
				for _, artifact := range result.Artifacts {
					var text string
					for _, part := range artifact.Parts {
						if textPart, ok := part.(protocol.TextPart); ok {
							text += textPart.Text
						}
					}
					c.Println(text)
				}
			}
		},
	})

	return a2aCmd
}

func runTask(
	ctx context.Context,
	agentNamespace, agentName string,
	userPrompt string,
	timeout time.Duration,
) (*protocol.Task, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}
	a2aURL := fmt.Sprintf("%s/%s/%s", cfg.A2AURL, agentNamespace, agentName)
	a2a, err := client.NewA2AClient(a2aURL)
	if err != nil {
		return nil, err
	}
	task, err := a2a.SendTasks(ctx, protocol.SendTaskParams{
		ID:        "kagent-task-" + uuid.NewString(),
		SessionID: nil,
		Message: protocol.Message{
			Role:  protocol.MessageRoleUser,
			Parts: []protocol.Part{protocol.NewTextPart(userPrompt)},
		},
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Process the task
	return waitForTaskResult(ctx, a2a, task.ID)
}

func waitForTaskResult(ctx context.Context, a2a *client.A2AClient, taskID string) (*protocol.Task, error) {
	// poll task result every 2s
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			task, err := a2a.GetTasks(ctx, protocol.TaskQueryParams{
				ID: taskID,
			})
			if err != nil {
				return nil, err
			}

			switch task.Status.State {
			case protocol.TaskStateSubmitted,
				protocol.TaskStateWorking:
				continue
			}

			return task, nil

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

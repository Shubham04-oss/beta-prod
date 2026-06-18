package workflows

import (
	"fmt"
	"time"

	"github.com/synq/agent-server/internal/llm"
	"go.temporal.io/sdk/workflow"
)

var allowedTeams = map[string]bool{
	"Orders": true,
	"PIM":    true,
	"Support": true,
}

// MasterWorkflowInput represents the incoming event (chat or cron)
type MasterWorkflowInput struct {
	EventSource string // e.g., "user_chat", "cron_daily_sync"
	Payload     string // e.g., "What is the status of order 123?"
}

// MasterWorkflowResult represents the final synthesized output
type MasterWorkflowResult struct {
	Response string
}

// CoreOrchestratorWorkflow is the primary Temporal workflow that acts as the Master Agent.
func CoreOrchestratorWorkflow(ctx workflow.Context, input MasterWorkflowInput) (*MasterWorkflowResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// In a real application, the orchestrator instance would be dependency-injected into the Worker
	// and accessed via activity environments. For simplicity here, we assume the activities are
	// registered on the worker and can access the master orchestrator.

	// 1. SMART ROUTING: Run the Classifier Agent Activity
	var plan llm.ExecutionPlan
	err := workflow.ExecuteActivity(ctx, "AnalyzeIntentActivity", input.Payload).Get(ctx, &plan)
	if err != nil {
		return nil, fmt.Errorf("failed to classify intent: %w", err)
	}

	workflow.GetLogger(ctx).Info("Classification complete", "plan", plan)

	// If no teams are needed, we can just answer directly or sleep
	if len(plan.Teams) == 0 {
		workflow.GetLogger(ctx).Info("No external teams required for this task. Handling directly.")
		var finalResponse string
		err = workflow.ExecuteActivity(ctx, "SynthesizeActivity", input.Payload, "No teams required. Handled by Aris.").Get(ctx, &finalResponse)
		return &MasterWorkflowResult{Response: finalResponse}, err
	}

	// 2. TRANSPARENT DELEGATION (AGUI STREAMING)
	aguiMessage := fmt.Sprintf("Aris delegating task (%s strategy). Rationale: %s", plan.Strategy, plan.Rationale)
	_ = workflow.ExecuteActivity(ctx, "StreamAGUIActivity", "delegation_status", aguiMessage).Get(ctx, nil)

	// 3. TARGETED ADAPTIVE DISPATCH
	var teamResults string

	switch plan.Strategy {
	case "parallel":
		var futures []workflow.Future
		for _, teamName := range plan.Teams {
			if !allowedTeams[teamName] {
				teamResults += fmt.Sprintf("\n[Invalid Team Requested: %s]", teamName)
				continue
			}
			wfName := teamName + "TeamWorkflow"
			f := workflow.ExecuteChildWorkflow(ctx, wfName, input)
			futures = append(futures, f)
		}
		for _, f := range futures {
			var res string
			if err := f.Get(ctx, &res); err != nil {
				teamResults += fmt.Sprintf("\n[Team Failed: %v]", err)
			} else {
				teamResults += fmt.Sprintf("\n[Team Result: %s]", res)
			}
		}

	case "sequential":
		for _, teamName := range plan.Teams {
			if !allowedTeams[teamName] {
				teamResults += fmt.Sprintf("\n[Invalid Team Requested: %s]", teamName)
				break
			}
			wfName := teamName + "TeamWorkflow"
			var res string
			// We pass the previous teamResults into the next team if needed
			if err := workflow.ExecuteChildWorkflow(ctx, wfName, input).Get(ctx, &res); err != nil {
				teamResults += fmt.Sprintf("\n[%s Failed: %v]", teamName, err)
				break // Stop sequence on failure
			}
			teamResults += fmt.Sprintf("\n[%s Result: %s]", teamName, res)
		}

	case "loop":
		// Simplified Loop: Retry the first team until a condition is met (mocked 3 max retries)
		teamName := plan.Teams[0]
		if !allowedTeams[teamName] {
			teamResults += fmt.Sprintf("\n[Invalid Team Requested: %s]", teamName)
			break
		}
		wfName := teamName + "TeamWorkflow"
		for i := 0; i < 3; i++ {
			var res string
			if err := workflow.ExecuteChildWorkflow(ctx, wfName, input).Get(ctx, &res); err == nil {
				teamResults += fmt.Sprintf("\n[%s Loop %d Result: %s]", teamName, i+1, res)
				break // Condition met (simulated by success)
			}
		}
	}

	// 4. SYNTHESIS
	// The Master Agent takes all the team results and synthesizes a final response
	var finalResponse string
	err = workflow.ExecuteActivity(ctx, "SynthesizeActivity", input.Payload, teamResults).Get(ctx, &finalResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize final response: %w", err)
	}

	return &MasterWorkflowResult{Response: finalResponse}, nil
}

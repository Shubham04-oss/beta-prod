package workflows_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	commonpb "go.temporal.io/api/common/v1"

	"github.com/synq/pkg/authcontext"
	"github.com/synq/agent-server/internal/llm"
	"github.com/synq/agent-server/internal/workflows"
)

// Dummy child workflows for testing adaptive routing
func OrdersTeamWorkflow(ctx workflow.Context, input workflows.MasterWorkflowInput) (string, error) {
	return "Orders processed", nil
}

func PIMTeamWorkflow(ctx workflow.Context, input workflows.MasterWorkflowInput) (string, error) {
	return "PIM processed", nil
}

func TestCoreOrchestratorWorkflow_Parallel(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	// Register dummy child workflows
	env.RegisterWorkflowWithOptions(OrdersTeamWorkflow, workflow.RegisterOptions{Name: "OrdersTeamWorkflow"})
	env.RegisterWorkflowWithOptions(PIMTeamWorkflow, workflow.RegisterOptions{Name: "PIMTeamWorkflow"})

	// Register dummy activities before mocking them by string name
	env.RegisterActivityWithOptions(func(ctx context.Context, payload string) (llm.ExecutionPlan, error) {
		return llm.ExecutionPlan{}, nil
	}, activity.RegisterOptions{Name: "AnalyzeIntentActivity"})
	env.RegisterActivityWithOptions(func(ctx context.Context, key, msg string) error { return nil }, activity.RegisterOptions{Name: "StreamAGUIActivity"})
	env.RegisterActivityWithOptions(func(ctx context.Context, payload, results string) (string, error) { return "", nil }, activity.RegisterOptions{Name: "SynthesizeActivity"})

	// Mock the Classifier Activity
	env.OnActivity("AnalyzeIntentActivity", mock.Anything, mock.Anything).Return(func(ctx context.Context, payload string) (llm.ExecutionPlan, error) {
		return llm.ExecutionPlan{
			Teams:     []string{"Orders", "PIM"},
			Strategy:  "parallel",
			Rationale: "Mock parallel execution",
		}, nil
	})

	// Mock AGUI Streaming Activity
	env.OnActivity("StreamAGUIActivity", mock.Anything, "delegation_status", mock.Anything).Return(nil)

	// Mock Synthesizer Activity
	env.OnActivity("SynthesizeActivity", mock.Anything, mock.Anything, mock.Anything).Return("Final unified response", nil)

	input := workflows.MasterWorkflowInput{Payload: "Sync everything"}
	env.ExecuteWorkflow(workflows.CoreOrchestratorWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result workflows.MasterWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Final unified response", result.Response)

	// Verify both child workflows were called
	env.AssertExpectations(t)
}

func TestCoreOrchestratorWorkflow_Sequential(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	// Register dummy child workflows
	env.RegisterWorkflowWithOptions(OrdersTeamWorkflow, workflow.RegisterOptions{Name: "OrdersTeamWorkflow"})
	env.RegisterWorkflowWithOptions(PIMTeamWorkflow, workflow.RegisterOptions{Name: "PIMTeamWorkflow"})

	// Register dummy activities before mocking them by string name
	env.RegisterActivityWithOptions(func(ctx context.Context, payload string) (llm.ExecutionPlan, error) {
		return llm.ExecutionPlan{}, nil
	}, activity.RegisterOptions{Name: "AnalyzeIntentActivity"})
	env.RegisterActivityWithOptions(func(ctx context.Context, key, msg string) error { return nil }, activity.RegisterOptions{Name: "StreamAGUIActivity"})
	env.RegisterActivityWithOptions(func(ctx context.Context, payload, results string) (string, error) { return "", nil }, activity.RegisterOptions{Name: "SynthesizeActivity"})

	// Mock the Classifier Activity to return sequential
	env.OnActivity("AnalyzeIntentActivity", mock.Anything, mock.Anything).Return(func(ctx context.Context, payload string) (llm.ExecutionPlan, error) {
		return llm.ExecutionPlan{
			Teams:     []string{"PIM", "Orders"},
			Strategy:  "sequential",
			Rationale: "PIM then Orders",
		}, nil
	})

	env.OnActivity("StreamAGUIActivity", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("SynthesizeActivity", mock.Anything, mock.Anything, mock.Anything).Return("Sequential finish", nil)

	input := workflows.MasterWorkflowInput{Payload: "Update products then sync"}
	env.ExecuteWorkflow(workflows.CoreOrchestratorWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result workflows.MasterWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Sequential finish", result.Response)
}

func TestCoreOrchestratorWorkflow_SecurityContextPropagation(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	// Register the Propagator we built earlier into the Test Environment
	ts.SetContextPropagators([]workflow.ContextPropagator{
		&authcontext.PillarContextPropagator{},
	})
	
	env := ts.NewTestWorkflowEnvironment()

	// In the child workflow, we assert the 4 Pillars are present
	secureChildWorkflow := func(ctx workflow.Context, input workflows.MasterWorkflowInput) (string, error) {
		tenantID := ctx.Value(authcontext.TenantIDKey)
		require.Equal(t, "tenant-123", tenantID)
		
		role := ctx.Value(authcontext.RoleKey)
		require.Equal(t, "admin", role)
		
		return "Secure execution complete", nil
	}
	
	env.RegisterWorkflowWithOptions(secureChildWorkflow, workflow.RegisterOptions{Name: "OrdersTeamWorkflow"})

	// Setup mock activities
	env.RegisterActivityWithOptions(func(ctx context.Context, payload string) (llm.ExecutionPlan, error) {
		return llm.ExecutionPlan{}, nil
	}, activity.RegisterOptions{Name: "AnalyzeIntentActivity"})
	env.RegisterActivityWithOptions(func(ctx context.Context, key, msg string) error { return nil }, activity.RegisterOptions{Name: "StreamAGUIActivity"})
	env.RegisterActivityWithOptions(func(ctx context.Context, payload, results string) (string, error) { return "", nil }, activity.RegisterOptions{Name: "SynthesizeActivity"})

	env.OnActivity("AnalyzeIntentActivity", mock.Anything, mock.Anything).Return(func(ctx context.Context, payload string) (llm.ExecutionPlan, error) {
		return llm.ExecutionPlan{
			Teams:     []string{"Orders"},
			Strategy:  "parallel",
			Rationale: "Security Test",
		}, nil
	})
	env.OnActivity("StreamAGUIActivity", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("SynthesizeActivity", mock.Anything, mock.Anything, mock.Anything).Return("Secured", nil)

	// Inject the 4 Pillars into the test environment using Temporal Headers
	// This simulates the PillarContextPropagator unpacking headers from an incoming gRPC request
	dc := converter.GetDefaultDataConverter()
	tenantPayload, _ := dc.ToPayload("tenant-123")
	rolePayload, _ := dc.ToPayload("admin")
	
	header := &commonpb.Header{
		Fields: map[string]*commonpb.Payload{
			authcontext.HeaderTenantID: tenantPayload,
			authcontext.HeaderRole:     rolePayload,
		},
	}
	env.SetHeader(header)

	input := workflows.MasterWorkflowInput{Payload: "Test Context Prop"}
	
	// Execute the workflow
	env.ExecuteWorkflow(workflows.CoreOrchestratorWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

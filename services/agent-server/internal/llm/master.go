package llm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// MasterOrchestrator represents the core System-Level Master Agent.
// It manages intent classification and delegates to Temporal workflows.
type MasterOrchestrator struct {
	ClassifierAgent agent.Agent
	SynthesizerAgent agent.Agent
}

// ExecutionPlan dictates which teams to wake up and how to execute them.
type ExecutionPlan struct {
	Teams     []string `json:"teams"`     // e.g. ["Orders", "PIM"]
	Strategy  string   `json:"strategy"`  // "sequential", "parallel", or "loop"
	Rationale string   `json:"rationale"`
}

// SetupArisOrchestrator initializes the Aris Master Agent and its internal ADK sub-agents.
func SetupArisOrchestrator(ctx context.Context, projectID, location string) (*MasterOrchestrator, error) {
	if projectID == "" || location == "" {
		return nil, fmt.Errorf("projectID and location must be provided for Master Orchestrator")
	}

	// The Master Orchestrator exclusively uses the fast, cheap gemini-3.5-flash model.
	masterModel, err := gemini.NewModel(
		ctx,
		"gemini-3.5-flash",
		&genai.ClientConfig{
			Backend:  genai.BackendVertexAI,
			Project:  projectID,
			Location: location,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize master model: %v", err)
	}

	// Internal Sub-Agent 1: The Intent Classifier
	// Its only job is to read the user request or cron ping and output a JSON execution plan.
	classifierSystemPrompt := `You are Aris, the core orchestrator for an Autonomous AI OS.
Analyze the incoming request and determine which Domain Teams (Orders, PIM, Support) are required.

Critically, you must adaptively determine the execution strategy:
- "parallel": Teams can work independently at the same time.
- "sequential": Teams must work one after the other (e.g., PIM must finish before Orders).
- "loop": A team must repeatedly execute until a validation condition is met.

Respond strictly in the following JSON format:
{
  "teams": ["TeamA", "TeamB"],
  "strategy": "parallel | sequential | loop",
  "rationale": "Brief explanation of why these teams and this strategy were chosen"
}`

	classifierAgent, err := llmagent.New(llmagent.Config{
		Name:        "ArisClassifier",
		Description: "Aris's internal sub-agent for adaptive routing.",
		Model:       masterModel,
		Instruction: classifierSystemPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier agent: %v", err)
	}

	// Internal Sub-Agent 2: The Synthesizer
	synthesizerSystemPrompt := `You are Aris, the face of the AI OS. 
You have delegated tasks to internal Domain Teams using adaptive strategies. They have returned their results.
Synthesize these results into a clear, unified, and professional response for the end-user.`

	synthesizerAgent, err := llmagent.New(llmagent.Config{
		Name:        "ArisSynthesizer",
		Description: "Aris's internal sub-agent for unifying team responses.",
		Model:       masterModel,
		Instruction: synthesizerSystemPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create synthesizer agent: %v", err)
	}

	return &MasterOrchestrator{
		ClassifierAgent: classifierAgent,
		SynthesizerAgent: synthesizerAgent,
	}, nil
}

// AnalyzeIntent evaluates the task and determines the adaptive routing strategy.
func (m *MasterOrchestrator) AnalyzeIntent(ctx context.Context, taskDescription string) (*ExecutionPlan, error) {
	// Sanitize input to prevent prompt injection
	if len(taskDescription) > 2000 {
		taskDescription = taskDescription[:2000]
	}
	taskDescription = strings.ReplaceAll(taskDescription, "```", "")
	
	log.Printf("[Aris] Analyzing Intent and Execution Strategy for: %s", taskDescription)

	// Mock routing logic simulating the JSON output of the ArisClassifier
	plan := ExecutionPlan{
		Teams:    []string{"Orders", "PIM"},
		Strategy: "parallel",
		Rationale: "Mock adaptive decision",
	}

	lowerDesc := strings.ToLower(taskDescription)
	if strings.Contains(lowerDesc, "after") || strings.Contains(lowerDesc, "then") {
		plan.Strategy = "sequential"
	} else if strings.Contains(lowerDesc, "until") || strings.Contains(lowerDesc, "retry") {
		plan.Strategy = "loop"
	}

	return &plan, nil
}

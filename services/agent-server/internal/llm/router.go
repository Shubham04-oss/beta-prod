package llm

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// SetupVertexRouter creates a base ADK Agent pre-configured with Vertex AI.
// In this setup, Gemini 3.5 Flash is the primary model, automatically inheriting
// IAM permissions from the Cloud Run service account.
func SetupVertexRouter(ctx context.Context, projectID, location string) (agent.Agent, error) {
	if projectID == "" || location == "" {
		return nil, fmt.Errorf("projectID and location must be provided for Vertex AI routing")
	}

	// Initialize the primary Vertex AI model binding
	primaryModel, err := gemini.NewModel(
		ctx,
		"gemini-3.5-flash",
		&genai.ClientConfig{
			Backend:  genai.BackendVertexAI,
			Project:  projectID,
			Location: location,
		},
	)
	if err != nil {
		return nil, err
	}

	// Build the foundational Router Agent
	routerAgent, err := llmagent.New(llmagent.Config{
		Name:        "VertexRouter",
		Description: "Primary LLM Router using Vertex AI backend",
		Model:       primaryModel,
	})
	if err != nil {
		return nil, err
	}

	return routerAgent, nil
}

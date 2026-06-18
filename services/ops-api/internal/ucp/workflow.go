package ucp

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// GenerateUCPFeedWorkflow orchestrates the extraction of PIM catalog data and uploads it as a UCP JSON-LD feed.
func GenerateUCPFeedWorkflow(ctx workflow.Context, tenantID string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *Activities

	// Step 1: Extract Catalog
	var feed UCPFeed
	err := workflow.ExecuteActivity(ctx, a.ExtractCatalogActivity, tenantID).Get(ctx, &feed)
	if err != nil {
		return "", err
	}

	if len(feed) == 0 {
		return "No products found for UCP feed", nil
	}

	// Step 2: Upload to GCS
	var gcsURL string
	err = workflow.ExecuteActivity(ctx, a.UploadFeedToGCSActivity, feed, tenantID).Get(ctx, &gcsURL)
	if err != nil {
		return "", err
	}

	return gcsURL, nil
}

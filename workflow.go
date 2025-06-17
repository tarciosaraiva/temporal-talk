package temporaltalk

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MainWorkflow(ctx workflow.Context) (string, error) {
	// This is the main workflow that orchestrates the child workflow.
	// It will call the ChildActionWorkflow and handle its result.

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute, // maximum time the activity can run
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second, // amount of time that must elapse before the first retry occurs
			MaximumInterval:    time.Minute, // maximum interval between retries
			BackoffCoefficient: 2,           // how much the retry interval increases
			MaximumAttempts:    5,           // Uncomment this if you want to limit attempts
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var basicActivity *BasicActivity

	var result string
	err := workflow.ExecuteActivity(ctx, basicActivity.RunBasicActivity, "onsite").Get(ctx, &result)
	if err != nil {
		return "", fmt.Errorf("Failed to run basic activity: %s", err)
	}

	return result, nil
}

// func ChildActionWorkflow(ctx workflow.Context) error {
// }

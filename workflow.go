package temporaltalk

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MainWorkflow(ctx workflow.Context, input string) (string, error) {
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
	var remoteServiceActivity *RemoteServiceActivity

	var result string
	err := workflow.ExecuteActivity(ctx, basicActivity.RunBasicActivity, input).Get(ctx, &result)
	if err != nil {
		return "", fmt.Errorf("Failed to run basic activity: %s", err)
	}

	var ip string
	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetIP).Get(ctx, &ip)
	if err != nil {
		return "", fmt.Errorf("Could not get IP: %s", err)
	}

	var location Geopoint
	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetLocationInfo, ip).Get(ctx, &location)
	if err != nil {
		return "", fmt.Errorf("Could not geolocate with IP: %s", err)
	}

	var weather string
	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetWeather, location).Get(ctx, &weather)
	if err != nil {
		return "", fmt.Errorf("Could not retrieve weather: %s", err)
	}

	return weather, nil
}

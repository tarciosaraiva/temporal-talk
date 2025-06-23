package temporaltalk

import (
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MainWorkflow(ctx workflow.Context, input string) (WeatherOutput, error) {
	// This is the main workflow that orchestrates the child workflow.
	// It will call the ChildActionWorkflow and handle its result.
	cwo := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var result WeatherOutput

	err := workflow.ExecuteChildWorkflow(ctx, ChildActionWorkflow, input).Get(ctx, &result)
	if err != nil {
		return result, fmt.Errorf("Failed to run child workflow: %s", err)
	}

	return result, nil
}

func ChildActionWorkflow(ctx workflow.Context, input string) (WeatherOutput, error) {
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

	var result WeatherOutput

	err := workflow.ExecuteActivity(ctx, basicActivity.RunBasicActivity, input).Get(ctx, &result.Name)
	if err != nil {
		return result, fmt.Errorf("Failed to run basic activity: %s", err)
	}

	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetIP).Get(ctx, &result.IpAddress)
	if err != nil {
		return result, fmt.Errorf("Could not get IP: %s", err)
	}

	var geoPoint Geopoint
	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetLocationInfo, result.IpAddress).Get(ctx, &geoPoint)
	if err != nil {
		return result, fmt.Errorf("Could not geolocate with IP: %s", err)
	}

	result.City = geoPoint.City

	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetWeather, geoPoint.Latitude, geoPoint.Longitude).Get(ctx, &result.CurrentForecast)
	if err != nil {
		return result, fmt.Errorf("Could not retrieve weather: %s", err)
	}

	return result, nil
}

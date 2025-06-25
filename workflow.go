package temporaltalk

import (
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	OnTheMoveUpdate  = "OnTheMove"
	StopMovingSignal = "StopMoving"
	TwentySeconds    = 30 * time.Second
)

type StopMovingSignalInput struct {
	IsMoving bool `json:"isMoving"`
}

func MainWorkflow(ctx workflow.Context, input WorkflowInput) (WeatherOutput, error) {
	logger := workflow.GetLogger(ctx)

	isMoving := input.Moving

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

	// This is the main workflow that orchestrates the child workflow.
	// It will call the ChildActionWorkflow and handle its result.
	cwo := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	// gets the current version
	// v := workflow.GetVersion(ctx, "V1", workflow.DefaultVersion, 1)

	var basicActivity *BasicActivity
	var remoteServiceActivity *RemoteServiceActivity

	var weatherOutput WeatherOutput

	err := workflow.ExecuteActivity(ctx, basicActivity.RunBasicActivity, input.Name).Get(ctx, &weatherOutput.Name)
	if err != nil {
		logger.Error("Failed to run basic activity.", "Error", err)
		return weatherOutput, fmt.Errorf("Failed to run basic activity: %s", err)
	}

	// if v == 1 {
	// 	err = workflow.ExecuteActivity(ctx, basicActivity.GetNumber).Get(ctx, &weatherOutput.RandomNumber)
	// 	if err != nil {
	// 		logger.Error("Failed to generate a random number.", "Error", err)
	// 		return weatherOutput, fmt.Errorf("Failed to generate a random number: %s", err)
	// 	}
	// }

	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetIP).Get(ctx, &weatherOutput.IpAddress)
	if err != nil {
		logger.Error("Could not retrieve IP address.", "Error", err)
		return weatherOutput, fmt.Errorf("Could not get IP: %s", err)
	}

	// setup a signal receiver
	var stopMovingSignalInput StopMovingSignalInput
	signalChannel := workflow.GetSignalChannel(ctx, StopMovingSignal)

	for isMoving {
		workflow.SetUpdateHandler(ctx, OnTheMoveUpdate, func(ctx workflow.Context) (WeatherOutput, error) {
			return executeChildWorkflow(ctx, weatherOutput)
		})

		logger.Debug("Setup update handler.")

		// go select concept reimplemented in Temporal
		// allows to wait on multiple communication ops
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &stopMovingSignalInput)
			isMoving = stopMovingSignalInput.IsMoving
		})

		selector.Select(ctx)

		logger.Debug("Selector waiting for signal.")
	}

	return executeChildWorkflow(ctx, weatherOutput)
}

func executeChildWorkflow(ctx workflow.Context, weatherOutput WeatherOutput) (WeatherOutput, error) {
	err := workflow.ExecuteChildWorkflow(ctx, ChildActionWorkflow, weatherOutput).Get(ctx, &weatherOutput)
	if err != nil {
		return weatherOutput, fmt.Errorf("Could not retrieve weather: %s", err)
	}
	return weatherOutput, nil
}

func ChildActionWorkflow(ctx workflow.Context, input WeatherOutput) (WeatherOutput, error) {
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

	var remoteServiceActivity *RemoteServiceActivity

	var geoPoint Geopoint
	err := workflow.ExecuteActivity(ctx, remoteServiceActivity.GetLocationInfo, input.IpAddress).Get(ctx, &geoPoint)
	if err != nil {
		return input, fmt.Errorf("Could not geolocate with IP: %s", err)
	}

	input.City = geoPoint.City

	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetWeather, geoPoint.Latitude, geoPoint.Longitude).Get(ctx, &input.CurrentForecast)
	if err != nil {
		return input, fmt.Errorf("Could not retrieve weather: %s", err)
	}

	return input, nil
}

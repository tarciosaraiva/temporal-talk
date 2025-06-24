package temporaltalk

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	OnTheMoveSignal = "OnTheMove"
	TwentySeconds   = 30 * time.Second
)

type OnTheMoveInput struct {
	IsMoving bool `json:"isMoving"`
}

func MainWorkflow(ctx workflow.Context, input WorkflowInput) error {
	logger := workflow.GetLogger(ctx)

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
		WorkflowID:        "remote-executions-wf",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var basicActivity *BasicActivity
	var remoteServiceActivity *RemoteServiceActivity

	var weatherOutput WeatherOutput

	err := workflow.ExecuteActivity(ctx, basicActivity.RunBasicActivity, input.Name).Get(ctx, &weatherOutput.Name)
	if err != nil {
		return fmt.Errorf("Failed to run basic activity: %s", err)
	}

	err = workflow.ExecuteActivity(ctx, remoteServiceActivity.GetIP).Get(ctx, &weatherOutput.IpAddress)
	if err != nil {
		return fmt.Errorf("Could not get IP: %s", err)
	}

	err = workflow.ExecuteChildWorkflow(ctx, ChildActionWorkflow, weatherOutput).Get(ctx, &weatherOutput)
	if err != nil {
		return fmt.Errorf("Could not retrieve weather: %s", err)
	}

	logger.Info(stringifiedWeather(weatherOutput))

	// setup a signal receiver
	// blocks execution when it is received
	var onTheMoveInput OnTheMoveInput
	signalChannel := workflow.GetSignalChannel(ctx, OnTheMoveSignal)
	signalChannel.Receive(ctx, &onTheMoveInput)

	logger.Info("Received Signal", "Channel", OnTheMoveSignal, "Input", &onTheMoveInput)

	// go select concept reimplemented in Temporal
	// allows to wait on multiple communication ops
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &onTheMoveInput)
	})

	// if it's moving starts a loop and sets up more waiting until another signal is sent
	for onTheMoveInput.IsMoving {
		selector.AddFuture(workflow.NewTimer(ctx, TwentySeconds), func(f workflow.Future) {
			err = workflow.ExecuteChildWorkflow(ctx, ChildActionWorkflow, weatherOutput).Get(ctx, &weatherOutput)
			if err != nil {
				logger.Error("Could not retrieve weather. Bye bye.")
			}

			logger.Info(stringifiedWeather(weatherOutput))
		})

		selector.Select(ctx)
	}

	return nil
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

func stringifiedWeather(input WeatherOutput) string {
	b, _ := json.MarshalIndent(input, "", "  ")
	return string(b)
}

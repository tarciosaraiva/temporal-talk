package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	temporaltalk "temporal-talk"
	"time"

	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()
	if len(os.Args) <= 1 {
		log.Fatalln("Must specify a name as the command-line argument")
	}

	name := os.Args[1]
	clientId := os.Args[2]
	isMoving, _ := strconv.ParseBool(os.Args[3])

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowID := fmt.Sprintf("main-%s-%s", name, clientId)

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: temporaltalk.TaskQueueName,
	}

	wInput := temporaltalk.WorkflowInput{
		Name:   name,
		Moving: isMoving,
	}

	var result temporaltalk.WeatherOutput

	we, err := c.ExecuteWorkflow(ctx, options, temporaltalk.MainWorkflow, wInput)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	if !isMoving {
		err = we.Get(ctx, &result)
		if err != nil {
			log.Fatalln("Unable to retrieve workflow result", err)
		}

		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
	}

	counter := 0

	log.Printf("Sleeping for 2 seconds for the workflow to execute its first tasks.")

	// give time for the workflow to run its first tasks
	time.Sleep(2 * time.Second)

	for isMoving {
		counter++

		updateHandle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
			WorkflowID:   we.GetID(),
			RunID:        we.GetRunID(),
			UpdateName:   temporaltalk.OnTheMoveUpdate,
			WaitForStage: client.WorkflowUpdateStageCompleted,
			Args:         []interface{}{true},
		})

		log.Printf("Setup update handle. Fetching result...")

		if err != nil {
			log.Fatalf("Unable to update workflow: %v", err)
		}

		err = updateHandle.Get(ctx, &result)
		if err != nil {
			log.Fatalf("Unable to get update result: %v", err)
		}

		b, _ := json.Marshal(result)
		fmt.Println(string(b))

		log.Printf("Iteration %d - Sleeping for 10 seconds...", counter)
		time.Sleep(10 * time.Second)

		if counter == 10 {
			log.Println("Signaling workflow to terminate.")
			c.SignalWorkflow(ctx, we.GetID(), we.GetRunID(), temporaltalk.StopMovingSignal, temporaltalk.StopMovingSignalInput{IsMoving: false})
			break
		}
	}
}

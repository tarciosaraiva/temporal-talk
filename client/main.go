package main

import (
	"context"
	"log"
	"os"
	temporaltalk "temporal-talk"

	"go.temporal.io/sdk/client"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalln("Must specify a name as the command-line argument")
	}

	name := os.Args[1]

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowID := "talk-for-onsite"

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: temporaltalk.TaskQueueName,
	}

	wInput := temporaltalk.WorkflowInput{
		Name:   name,
		Moving: false,
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, temporaltalk.MainWorkflow, wInput)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	var result temporaltalk.WeatherOutput
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}

	// b, _ := json.MarshalIndent(result, "", "  ")

	// fmt.Println(string(b))
}

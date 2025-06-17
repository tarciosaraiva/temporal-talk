package main

import (
	"context"
	"fmt"
	"log"
	"os"
	temporaltalk "temporal-talk"

	"go.temporal.io/sdk/client"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalln("Must specify a name as the command-line argument")
	}

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

	we, err := c.ExecuteWorkflow(context.Background(), options, temporaltalk.MainWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}

	fmt.Println(result)
}

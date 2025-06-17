package main

import (
	"log"
	temporaltalk "temporal-talk"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Create the Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Create the Temporal worker
	w := worker.New(c, temporaltalk.TaskQueueName, worker.Options{})

	// inject HTTP client into the Activities Struct
	basicActivity := &temporaltalk.BasicActivity{}

	// Register Workflow and Activities
	w.RegisterWorkflow(temporaltalk.MainWorkflow)
	w.RegisterActivity(basicActivity)

	// Start the Worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Temporal worker", err)
	}
}

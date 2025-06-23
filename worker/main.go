package main

import (
	"log"
	"net/http"
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

	basicActivity := &temporaltalk.BasicActivity{}

	remoteServiceActivity := &temporaltalk.RemoteServiceActivity{
		HTTPClient: http.DefaultClient,
	}

	// Register Workflow and Activities
	w.RegisterWorkflow(temporaltalk.MainWorkflow)
	// w.RegisterWorkflow(temporaltalk.ChildActionWorkflow)
	w.RegisterActivity(basicActivity)
	w.RegisterActivity(remoteServiceActivity)

	// Start the Worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Temporal worker", err)
	}
}

package temporaltalk_test

import (
	temporaltalk "temporal-talk"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

func TestRunBasicActivity(t *testing.T) {
	// set up test environment
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	basicActivity := &temporaltalk.BasicActivity{}

	env.RegisterActivity(basicActivity)

	// Call the RunBasicActivity function
	val, err := env.ExecuteActivity(basicActivity.RunBasicActivity, "tester")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// get the Activity result
	var result string
	val.Get(&result)

	// Validate the returned string
	expectedMessage := "Hello, tester!"
	assert.Equal(t, expectedMessage, result)
}

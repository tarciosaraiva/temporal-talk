package temporaltalk

import (
	"context"
	"fmt"
)

type BasicActivity struct{}

func (a *BasicActivity) RunBasicActivity(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Hello, %s!", input), nil
}

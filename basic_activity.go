package temporaltalk

import (
	"context"
	"fmt"
)

type BasicActivity struct{}

func (a *BasicActivity) RunBasicActivity(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Hello, %s!", input), nil
}

func (a *BasicActivity) GetNumber(ctx context.Context) (int, error) {
	return 100, nil
}

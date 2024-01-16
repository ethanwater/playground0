package main

import (
	"context"

	"vivian.app/app"
)

func main() {
	err := app.Deploy(context.Background())
	if err != nil {
		return
	}
}

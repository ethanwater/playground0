package main

import (
	"context"

	"vivian.app/internal/app"
)

func main() {
	err := app.Deploy(context.Background())
	if err != nil {
		return
	}
}

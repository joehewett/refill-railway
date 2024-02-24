package main

import (
	"fmt"

	refill "github.com/joehewett/refill/internal"
)

func main() {
	api := refill.NewAPIServer()

	err := api.Run()
	if err != nil {
		fmt.Printf("API error occurred: %s\n", err)
	}
}

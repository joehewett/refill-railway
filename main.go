package main

import (
	"os"

	refill "github.com/joehewett/refill/internal"
	"github.com/unidoc/unipdf/v3/common/license"
)

func init() {
	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		panic(err)
	}
}

func main() {
	server := refill.NewAPIServer(":8080")
	server.Run()
}

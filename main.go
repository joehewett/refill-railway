package main

import (
	"os"

	refill "github.com/joehewett/refill/internal"
)

var (
	port = os.Getenv("PORT")
)

// func init() {
// 	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
// 	if err != nil {
// 		panic(err)
// 	}
// }

func main() {
	api := refill.NewAPIServer()

	api.Run()
}

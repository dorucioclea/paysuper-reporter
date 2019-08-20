package main

import "github.com/paysuper/paysuper-reporter/internal"

func main() {
	app := internal.NewApplication()

	defer app.Stop()
	app.Run()
}

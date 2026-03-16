package main

import (
	"fmt"
	"os"

	"scaffold-api/cmd"
	_ "scaffold-api/docs"
)

//go:generate swag init --parseInternal -g main.go -o docs

// @title Scaffold API Users API
// @version 1.0
// @description Online Swagger documentation for the users CRUD service.
// @BasePath /
// @schemes http https
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

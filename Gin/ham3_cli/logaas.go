package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func LogaasCreate(c *cli.Context) error {
	fmt.Println("Creating Logaas cluster")
	return nil
}

func LogaasDelete(c *cli.Context) error {
	fmt.Println("Deleting Logaas cluster")
	return nil
}

package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func CaasCreate(c *cli.Context) error {
	fmt.Println("Creating Caas cluster")
	return nil
}

func CaasDelete(c *cli.Context) error {
	fmt.Println("Deleting Caas cluster")
	return nil
}

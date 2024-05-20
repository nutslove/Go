package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
)

func CaasCreate(c *cli.Context) error {
	s := Spinner("Creating Caas cluster..")
	s.Start()
	fmt.Println("Creating Caas cluster")
	time.Sleep(2 * time.Second)
	s.Stop()
	return nil
}

func CaasDelete(c *cli.Context) error {
	s := Spinner("Deleting Caas cluster..")
	s.Start()
	fmt.Println("Deleting Caas cluster")
	time.Sleep(2 * time.Second)
	s.Stop()
	return nil
}

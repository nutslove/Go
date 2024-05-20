package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
)

func LogaasCreate(c *cli.Context) error {
	s := Spinner("Creating Logaas cluster..")
	s.Start()
	fmt.Println("Creating Logaas cluster")
	time.Sleep(2 * time.Second)
	s.Stop()
	return nil
}

func LogaasDelete(c *cli.Context) error {
	s := Spinner("Deleting Logaas cluster..")
	s.Start()
	fmt.Println("Deleting Logaas cluster")
	time.Sleep(2 * time.Second)
	s.Stop()
	return nil
}

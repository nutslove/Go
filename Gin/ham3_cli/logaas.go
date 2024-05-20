package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func LogaasCreate(c *cli.Context) error {
	s := Spinner("Creating Logaas cluster..")
	s.Start()
	time.Sleep(2 * time.Second)
	s.Stop()
	fmt.Println(color.New(color.FgGreen).Sprint("Logaas cluster created successfully"))

	s = Spinner("Creating Dashboard..")
	s.Start()
	time.Sleep(2 * time.Second)
	s.Stop()
	fmt.Println(color.New(color.FgGreen).Sprint("Dashboard created successfully"))
	return nil
}

func LogaasDelete(c *cli.Context) error {
	s := Spinner("Deleting Logaas cluster..")
	s.Start()
	time.Sleep(2 * time.Second)
	s.Stop()
	fmt.Println(color.New(color.FgGreen).Sprint("Logaas cluster deleted successfully"))

	s = Spinner("Deleting Dashboard..")
	s.Start()
	time.Sleep(2 * time.Second)
	s.Stop()
	fmt.Println(color.New(color.FgGreen).Sprint("Dashboard deleted successfully"))
	return nil
}

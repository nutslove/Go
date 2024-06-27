package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	CaasEndpoint = "http://localhost:8081/api/v1/caas"
)

func CreateCaaS(c *cli.Context) error {
	tenant := c.String("tenant-id")

	s := Spinner("Creating CaaS cluster..")
	s.Start()
	resp, err := http.Post(fmt.Sprintf("%s/%s", CaasEndpoint, tenant), "application/json", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}
	s.Stop()
	fmt.Println("POST Response:", string(body))

	fmt.Println(color.New(color.FgGreen).Sprint("CaaS cluster created successfully"))
	return nil
}

func GetCaaS(c *cli.Context) error {
	tenant := c.String("tenant-id")

	s := Spinner("Getting info about CaaS cluster..")
	s.Start()
	resp, err := http.Get(fmt.Sprintf("%s/%s", CaasEndpoint, tenant))
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}
	s.Stop()
	fmt.Println("GET Response:", string(body))

	fmt.Println(color.New(color.FgGreen).Sprint("CaaS cluster retrieved successfully"))
	return nil
}

func DeleteCaaS(c *cli.Context) error {
	tenant := c.String("tenant-id")

	s := Spinner("Deleting CaaS cluster..")
	s.Start()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", CaasEndpoint, tenant), nil)
	if err != nil {
		fmt.Println("Error creating DELETE request:", err)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()
	s.Stop()
	fmt.Println("DELETE Response:", resp.Status)

	fmt.Println(color.New(color.FgGreen).Sprint("CaaS cluster deleted successfully"))
	return nil
}

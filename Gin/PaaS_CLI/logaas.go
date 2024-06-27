package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type ClusterInfo struct {
	ClusterName string `json:"cluster-name"`
	ClusterType string `json:"cluster-type"`
}

const (
	LoggasEndpoint = "http://localhost:8081/api/v1/logaas"
)

func CreateLOGaaS(c *cli.Context) error {
	clsutername := c.String("cluster-name")
	clustertype := c.String("cluster-type")

	clusterInfo := ClusterInfo{
		ClusterName: clsutername,
		ClusterType: clustertype,
	}

	jsonData, err := json.Marshal(clusterInfo)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	s := Spinner("Creating LOGaaS cluster..")
	s.Start()
	resp, err := http.Post(fmt.Sprintf("%s/%s", LoggasEndpoint, clsutername), "application/json", bytes.NewBuffer(jsonData))
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

	fmt.Println(color.New(color.FgGreen).Sprintf("%s LOGaaS %s cluster created successfully", clsutername, clustertype))

	s = Spinner("Creating Dashboard..")
	s.Start()
	time.Sleep(2 * time.Second)
	s.Stop()
	fmt.Println(color.New(color.FgGreen).Sprint("Dashboard created successfully"))
	return nil
}

func DeleteLOGaaS(c *cli.Context) error {
	clsutername := c.String("cluster-name")
	clustertype := c.String("cluster-type")

	clusterInfo := ClusterInfo{
		ClusterName: clsutername,
		ClusterType: clustertype,
	}

	jsonData, err := json.Marshal(clusterInfo)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	s := Spinner("Deleting LOGaaS cluster..")
	s.Start()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", LoggasEndpoint, clsutername), bytes.NewBuffer(jsonData))
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}
	s.Stop()
	fmt.Println("DELETE Response:", string(body))

	fmt.Println(color.New(color.FgGreen).Sprintf("%s LOGaaS %s cluster deleted successfully", clsutername, clustertype))

	s = Spinner("Deleting Dashboard..")
	s.Start()
	time.Sleep(2 * time.Second)
	s.Stop()
	fmt.Println(color.New(color.FgGreen).Sprint("Dashboard deleted successfully"))
	return nil
}

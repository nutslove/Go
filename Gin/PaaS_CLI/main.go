package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "ham3",
		Usage: "CLI for Ham3",
		Commands: []*cli.Command{
			{
				Name:  "caas",
				Usage: "Container as a Service",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a CaaS cluster",
						Action: CreateCaaS,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "tenant-id",
								Usage:    "ID(Name) of the tenant",
								Required: true,
							},
						},
					},
					{
						Name:   "get",
						Usage:  "Get info about a CaaS cluster",
						Action: GetCaaS,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "tenant-id",
								Usage:    "ID(Name) of the tenant",
								Required: true,
							},
						},
					},
					{
						Name:   "delete",
						Usage:  "Delete a CaaS cluster",
						Action: DeleteCaaS,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "tenant-id",
								Usage:    "ID(Name) of the tenant",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:  "logaas",
				Usage: "Logging as a Service",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a LOGaaS cluster",
						Action: CreateLOGaaS,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster-name",
								Usage:    "Name of the cluster",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "cluster-type",
								Usage:    "Type of the cluster (standard, scalable)",
								Required: true,
							},
						},
					},
					{
						Name:   "delete",
						Usage:  "Delete a LOGaaS cluster",
						Action: DeleteLOGaaS,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster-name",
								Usage:    "Name of the cluster",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "cluster-type",
								Usage:    "Type of the cluster (standard, scalable)",
								Required: true,
							},
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

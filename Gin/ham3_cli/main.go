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
				Name:  "logaas",
				Usage: "Logging as a service",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a logaas cluster",
						Action: LogaasCreate,
					},
					{
						Name:   "delete",
						Usage:  "Delete a logaas cluster",
						Action: LogaasDelete,
					},
				},
			},
			{
				Name:  "caas",
				Usage: "Container as a service",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create a caas cluster",
						Action: CaasCreate,
					},
					{
						Name:   "delete",
						Usage:  "Delete a caas cluster",
						Action: CaasDelete,
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

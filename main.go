package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetFlags(0)

	app := &cli.App{
		Name:  "hcloud-kv",
		Usage: "hetzner cloud key/value store",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "initializes a new database",
				Action: func(cCtx *cli.Context) error {
					database := setupDB(cCtx.String("db"))
					database.Init()
					return nil
				},
			},
			{
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   "sets a key",
				Action: func(cCtx *cli.Context) error {
					database := setupDB(cCtx.String("db"))

					key := cCtx.Args().First()
					val := cCtx.Args().Get(1)

					if len(key) > 63 || len(val) > 63 {
						log.Fatalf("error updating key: max len for key and value is 63")
					}

					database.Set(key, val)
					return nil
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "get a value from given key",
				Action: func(cCtx *cli.Context) error {
					database := setupDB(cCtx.String("db"))
					fmt.Println(database.Get(cCtx.Args().First()))
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list all keys",
				Action: func(cCtx *cli.Context) error {
					database := setupDB(cCtx.String("db"))
					keys := database.List()
					fmt.Println(strings.Join(keys, "\n"))
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "db",
				Value: "0",
				Usage: "database to use",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func setupDB(name string) Database {
	token := os.Getenv("HCLOUD_TOKEN")

	return Database{
		Client:  hcloud.NewClient(hcloud.WithToken(token)),
		Context: context.Background(),
		Name:    fmt.Sprintf("hkv-%s", name),
	}
}

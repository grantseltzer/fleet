package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/kolide/fleet/server/kolide"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	yamlFlagName        = "yaml"
	jsonFlagName        = "json"
	withQueriesFlagName = "with-queries"
)

type specGeneric struct {
	Kind    string      `json:"kind"`
	Version string      `json:"apiVersion"`
	Spec    interface{} `json:"spec"`
}

func defaultTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	return table
}

func yamlFlag() cli.BoolFlag {
	return cli.BoolFlag{
		Name:  yamlFlagName,
		Usage: "Output in yaml format",
	}
}

func jsonFlag() cli.BoolFlag {
	return cli.BoolFlag{
		Name:  jsonFlagName,
		Usage: "Output in JSON format",
	}
}

func printJSON(spec interface{}) error {
	b, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", b)
	return nil
}

func printYaml(spec interface{}) error {
	b, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}
	fmt.Printf("---\n%s", string(b))
	return nil
}

func printLabel(c *cli.Context, label *kolide.LabelSpec) error {
	spec := specGeneric{
		Kind:    "label",
		Version: kolide.ApiVersion,
		Spec:    label,
	}

	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func printQuery(c *cli.Context, query *kolide.QuerySpec) error {
	spec := specGeneric{
		Kind:    "query",
		Version: kolide.ApiVersion,
		Spec:    query,
	}

	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func printPack(c *cli.Context, pack *kolide.PackSpec) error {
	spec := specGeneric{
		Kind:    "pack",
		Version: kolide.ApiVersion,
		Spec:    pack,
	}

	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func printOption(c *cli.Context, option *kolide.OptionsSpec) error {
	spec := specGeneric{
		Kind:    "option",
		Version: kolide.ApiVersion,
		Spec:    option,
	}

	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func printSecret(c *cli.Context, secret *kolide.EnrollSecretSpec) error {
	spec := specGeneric{
		Kind:    "enroll_secret",
		Version: kolide.ApiVersion,
		Spec:    secret,
	}

	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func printHost(c *cli.Context, host *kolide.Host) error {
	spec := specGeneric{
		Kind:    "host",
		Version: kolide.ApiVersion,
		Spec:    host,
	}

	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func printConfig(c *cli.Context, config *kolide.AppConfigPayload) error {
	spec := specGeneric{
		Kind:    "config",
		Version: kolide.ApiVersion,
		Spec:    config,
	}
	var err error

	if c.Bool(jsonFlagName) {
		err = printJSON(spec)
	} else {
		err = printYaml(spec)
	}

	return err
}

func getCommand() cli.Command {
	return cli.Command{
		Name:  "get",
		Usage: "Get/list resources",
		Subcommands: []cli.Command{
			getQueriesCommand(),
			getPacksCommand(),
			getLabelsCommand(),
			getOptionsCommand(),
			getHostsCommand(),
			getEnrollSecretCommand(),
			getAppConfigCommand(),
		},
	}
}

func getQueriesCommand() cli.Command {
	return cli.Command{
		Name:    "queries",
		Aliases: []string{"query", "q"},
		Usage:   "List information about one or more queries",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			name := c.Args().First()

			// if name wasn't provided, list all queries
			if name == "" {
				queries, err := fleet.GetQueries()
				if err != nil {
					return errors.Wrap(err, "could not list queries")
				}

				if len(queries) == 0 {
					fmt.Println("no queries found")
					return nil
				}

				if c.Bool(yamlFlagName) || c.Bool(jsonFlagName) {
					for _, query := range queries {
						if err := printQuery(c, query); err != nil {
							return errors.Wrap(err, "unable to print query")
						}
					}
				} else {
					// Default to printing as a table
					data := [][]string{}

					for _, query := range queries {
						data = append(data, []string{
							query.Name,
							query.Description,
							query.Query,
						})
					}

					table := defaultTable()
					table.SetHeader([]string{"name", "description", "query"})
					table.AppendBulk(data)
					table.Render()
				}
				return nil
			}

			query, err := fleet.GetQuery(name)
			if err != nil {
				return err
			}

			if err := printQuery(c, query); err != nil {
				return errors.Wrap(err, "unable to print query")
			}

			return nil

		},
	}
}

func getPacksCommand() cli.Command {
	return cli.Command{
		Name:    "packs",
		Aliases: []string{"pack", "p"},
		Usage:   "List information about one or more packs",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
			cli.BoolFlag{
				Name:  withQueriesFlagName,
				Usage: "Output queries included in pack(s) too",
			},
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			name := c.Args().First()
			shouldPrintQueries := c.Bool(withQueriesFlagName)
			queriesToPrint := make(map[string]bool)

			addQueries := func(pack *kolide.PackSpec) {
				if shouldPrintQueries {
					for _, q := range pack.Queries {
						queriesToPrint[q.QueryName] = true
					}
				}
			}

			printQueries := func() error {
				if !shouldPrintQueries {
					return nil
				}

				queries, err := fleet.GetQueries()
				if err != nil {
					return errors.Wrap(err, "could not list queries")
				}

				// Getting all queries then filtering is usually faster than getting
				// one query at a time.
				for _, query := range queries {
					if !queriesToPrint[query.Name] {
						continue
					}

					if err := printQuery(c, query); err != nil {
						return errors.Wrap(err, "unable to print query")
					}
				}

				return nil
			}

			// if name wasn't provided, list all packs
			if name == "" {
				packs, err := fleet.GetPacks()
				if err != nil {
					return errors.Wrap(err, "could not list packs")
				}

				if c.Bool(yamlFlagName) {
					for _, pack := range packs {
						if err := printPack(c, pack); err != nil {
							return errors.Wrap(err, "unable to print pack")
						}

						addQueries(pack)
					}

					return printQueries()
				}

				if len(packs) == 0 {
					fmt.Println("no packs found")
					return nil
				}

				data := [][]string{}

				for _, pack := range packs {
					data = append(data, []string{
						pack.Name,
						pack.Platform,
						pack.Description,
					})
				}

				table := defaultTable()
				table.SetHeader([]string{"name", "platform", "description"})
				table.AppendBulk(data)
				table.Render()

				return nil
			}

			// Name was specified
			pack, err := fleet.GetPack(name)
			if err != nil {
				return err
			}

			addQueries(pack)

			if err := printPack(c, pack); err != nil {
				return errors.Wrap(err, "unable to print pack")
			}

			return printQueries()

		},
	}
}

func getLabelsCommand() cli.Command {
	return cli.Command{
		Name:    "labels",
		Aliases: []string{"label", "l"},
		Usage:   "List information about one or more labels",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			name := c.Args().First()

			// if name wasn't provided, list all labels
			if name == "" {
				labels, err := fleet.GetLabels()
				if err != nil {
					return errors.Wrap(err, "could not list labels")
				}

				if c.Bool(yamlFlagName) || c.Bool(jsonFlagName) {
					for _, label := range labels {
						printLabel(c, label)
					}
					return nil
				}

				if len(labels) == 0 {
					fmt.Println("no labels found")
					return nil
				}

				// Default to printing as a table
				data := [][]string{}

				for _, label := range labels {
					data = append(data, []string{
						label.Name,
						label.Platform,
						label.Description,
						label.Query,
					})
				}

				table := defaultTable()
				table.SetHeader([]string{"name", "platform", "description", "query"})
				table.AppendBulk(data)
				table.Render()

				return nil
			}

			// Label name was specified
			label, err := fleet.GetLabel(name)
			if err != nil {
				return err
			}

			printLabel(c, label)
			return nil

		},
	}
}

func getOptionsCommand() cli.Command {
	return cli.Command{
		Name:  "options",
		Usage: "Retrieve the osquery configuration",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			options, err := fleet.GetOptions()
			if err != nil {
				return err
			}

			err = printOption(c, options)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func getEnrollSecretCommand() cli.Command {
	return cli.Command{
		Name:    "enroll_secret",
		Aliases: []string{"enroll_secrets", "enroll-secret", "enroll-secrets"},
		Usage:   "Retrieve the osquery enroll secrets",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			secrets, err := fleet.GetEnrollSecretSpec()
			if err != nil {
				return err
			}

			err = printSecret(c, secrets)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func getAppConfigCommand() cli.Command {
	return cli.Command{
		Name:  "config",
		Usage: "Retrieve the Fleet configuration",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			config, err := fleet.GetAppConfig()
			if err != nil {
				return err
			}

			err = printConfig(c, config)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func getHostsCommand() cli.Command {
	return cli.Command{
		Name:    "hosts",
		Aliases: []string{"host", "h"},
		Usage:   "List information about one or more hosts",
		Flags: []cli.Flag{
			jsonFlag(),
			yamlFlag(),
			configFlag(),
			contextFlag(),
		},
		Action: func(c *cli.Context) error {
			fleet, err := clientFromCLI(c)
			if err != nil {
				return err
			}

			hosts, err := fleet.GetHosts()
			if err != nil {
				return errors.Wrap(err, "could not list hosts")
			}

			if len(hosts) == 0 {
				fmt.Println("no hosts found")
				return nil
			}

			if c.Bool(jsonFlagName) || c.Bool(yamlFlagName) {
				for _, host := range hosts {
					err = printHost(c, &host.Host)
					if err != nil {
						return err
					}
				}
				return nil
			}

			// Default to printing as table
			data := [][]string{}

			for _, host := range hosts {
				data = append(data, []string{
					host.Host.UUID,
					host.DisplayText,
					host.Host.Platform,
					host.Status,
				})
			}

			table := defaultTable()
			table.SetHeader([]string{"uuid", "hostname", "platform", "status"})
			table.AppendBulk(data)
			table.Render()

			return nil
		},
	}
}

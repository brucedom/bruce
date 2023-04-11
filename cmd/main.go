package main

import (
	"cfs/config"
	"cfs/handlers"
	"cfs/system"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var (
	version = "source"
)

func setLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	setLogger()
	err := system.InitializeSysInfo()
	if err != nil {
		log.Error().Err(err).Msg("cannot start with unknown system info")
		os.Exit(1)
	}
	app := &cli.App{
		Name:  "cfs",
		Usage: "Start with: /path/to/cfs https://someinstallhost/installme.yml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "/etc/cfs/config.yml",
				Usage: "See docs for supported endpoints, eg: https://s3.amazonaws.com/somebucket/my_install.yml",
			},
			&cli.StringFlag{
				Name:    "property-file",
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "Loads properties from a file, eg: /etc/cfs/properties.yml to be used as environment variables for operators and templates",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "Enable debug logging",
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.Bool("debug") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if cCtx.Args().First() != "" {
				t, err := config.LoadConfig(cCtx.Args().First())
				if err != nil {
					log.Error().Err(err).Msg("cannot continue without configuration data")
					os.Exit(1)
				}
				handlers.Install(t, cCtx.String("property-file"))
				return nil
			}
			t, err := config.LoadConfig(cCtx.String("config"))
			if err != nil {
				log.Error().Err(err).Msg("cannot continue without configuration data")
				os.Exit(1)
			}
			handlers.Install(t, cCtx.String("property-file"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"setup"},
				Usage:   "this is the default action and will be run if no commands are specified",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					t, err := config.LoadConfig(cCtx.String("config"))
					if err != nil {
						log.Error().Err(err).Msg("cannot continue without configuration data")
						os.Exit(1)
					}
					handlers.Install(t, cCtx.String("property-file"))
					return nil
				},
			},
			{
				Name:    "search",
				Aliases: []string{"find"},
				Usage:   "this will search the ConfigSet repository for a related manifest",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					handlers.Search(cCtx.Args().First(), cCtx.Args().Get(1))
					return nil
				},
			},
			{
				Name:    "view",
				Aliases: []string{"open"},
				Usage:   "this command opens the manifest for you to view in CLI prior to executing install",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					handlers.View(cCtx.Args().First())
					return nil
				},
			},
			{
				Name:  "create",
				Usage: "this command will create a new manifest or template for you to upload directly to your configset.com account and then install, you must have CFS_KEY env variable set to your API key from configset.com",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "kind",
						Aliases: []string{"k"},
						Value:   "template",
						Usage:   "the kind of item to be uploaded, either 'template' or 'manifest'",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Value:   "",
						Usage:   "name for the item to be uploaded",
					},
					&cli.StringFlag{
						Name:    "description",
						Aliases: []string{"d"},
						Value:   "",
						Usage:   "a brief description for the file to be uplaoded, can be edited with more detail later.",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					handlers.Create(cCtx.String("kind"), cCtx.String("name"), cCtx.String("description"), cCtx.Args().First())
					return nil
				},
			},
			{
				Name:  "edit",
				Usage: "this command will edit a manifest or template by re-uploading to your configset.com account, you must have CFS_KEY env variable set to your API key from configset.com, name and description can be edited directly on configset.com",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "kind",
						Aliases: []string{"k"},
						Value:   "template",
						Usage:   "the kind of item to be uploaded, either 'template' or 'manifest'",
					},
					&cli.StringFlag{
						Name:    "id",
						Aliases: []string{"i"},
						Value:   "",
						Usage:   "id for the item to be uploaded",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					handlers.Edit(cCtx.String("kind"), cCtx.String("id"), cCtx.Args().First())
					return nil
				},
			},
		},
	}
	log.Debug().Msgf("Starting Bruce (Version: %s)", version)
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err)
	}

}

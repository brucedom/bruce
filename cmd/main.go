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
	if version == "source" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}
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
		},
		Action: func(cCtx *cli.Context) error {
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
					handlers.Search(cCtx.Args().First())
					return nil
				},
			},
			{
				Name:    "view",
				Aliases: []string{"open"},
				Usage:   "this command opens the manifest for you to view in CLI prior to executing install",
				Action: func(cCtx *cli.Context) error {
					handlers.View(cCtx.Args().First())
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

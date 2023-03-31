package handlers

import (
	"cfs/config"
	"cfs/loader"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

func Install(t *config.TemplateData, propfile string) error {
	log.Debug().Msg("starting install task")
	log.Debug().Msgf("propfile: %s", propfile)
	err := loadPropData(propfile)
	if err != nil {
		log.Error().Err(err).Msg("cannot proceed without the properties file specified.")
		os.Exit(1)
	}
	for idx, step := range t.Steps {
		if step.Action != nil {
			err := step.Action.Execute()
			if err != nil {
				log.Error().Err(err).Msgf("error executing step [%d]", idx+1)
				os.Exit(1)
			}
		}
	}
	return nil
}

// loadPropData loads the property data from property file and does os.SetEnv env for each property value.
func loadPropData(propFile string) error {
	if len(propFile) < 1 {
		return nil
	}
	// read content of property file and unmarshal into map
	d, _, err := loader.ReadRemoteFile(propFile)
	if err != nil {
		return err
	}
	log.Debug().Bytes("rawConfig", d)
	c := make(map[string]string)

	err = yaml.Unmarshal(d, c)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse config file")
	}
	for k, v := range c {
		log.Debug().Msgf("setting env var: %s=%s", k, v)
		os.Setenv(k, v)
	}
	return nil
}

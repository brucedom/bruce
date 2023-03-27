package handlers

import (
	"bruce/config"
	"github.com/rs/zerolog/log"
	"os"
)

func Install(t *config.TemplateData) error {

	log.Debug().Msg("starting install task")

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

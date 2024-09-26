package handlers

import (
	"bruce/config"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func RunServer(svr_config string, portNumber int) error {
	log.Debug().Msg("starting server task")

	// Read server configuration
	sc := &config.ServerConfig{}
	err := config.ReadServerConfig(svr_config, sc)
	if err != nil {
		log.Error().Err(err).Msg("cannot continue without configuration data")
		os.Exit(1)
	}

	// Channel to receive OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// WaitGroup to manage the lifecycle of goroutines
	var wg sync.WaitGroup

	// Context to manage cancellation of goroutines
	ctx, cancel := context.WithCancel(context.Background())

	log.Info().Msg("Starting Bruce in server mode")

	if len(sc.Execution) == 0 {
		log.Error().Msg("no execution targets configured")
		os.Exit(1)
	}

	for _, e := range sc.Execution {
		// Validate execution type
		if e.Type != "event" && e.Type != "cadence" {
			log.Info().Msgf("Skipping invalid execution target: %s must be of type 'event' or 'cadence'", e.Name)
			continue
		}

		// Start CadenceRunner
		if e.Type == "cadence" {
			wg.Add(1)
			go func(e config.Execution) {
				defer wg.Done()
				CadenceRunner(ctx, e.Name, e.Target, e.Cadence)
			}(e)
		}

		// Start SocketRunner
		if e.Type == "event" {
			wg.Add(1)
			go func(e config.Execution) {
				defer wg.Done()
				SocketRunner(ctx, e.Name, e.Target, sc.Endpoint, sc.Key, e.Authorization)
			}(e)
		}
	}

	// Wait for a signal to shut down
	<-sigCh
	log.Info().Msg("Shutting down server...")

	// Cancel all goroutines
	cancel()

	// Wait for all runners to finish
	wg.Wait()

	log.Info().Msg("All runners finished, server shut down successfully.")
	return nil
}

func CadenceRunner(ctx context.Context, name, propfile string, cadence int) {
	log.Debug().Msgf("Starting CadenceRunner[%s] with propfile: %s, every %d minutes", name, propfile, cadence)
	t, err := config.LoadConfig(propfile)
	if err != nil {
		log.Error().Err(err).Msgf("cannot continue without configuration data, runner %s failed", name)
		return
	}

	// Run the task at the specified cadence interval
	ticker := time.NewTicker(time.Duration(cadence) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("CadenceRunner[%s] received cancellation signal, exiting...", name)
			return
		case <-ticker.C:
			log.Debug().Msgf("CadenceRunner[%s] running execution steps", name)
			err = ExecuteSteps(t)
			if err != nil {
				log.Error().Err(err).Msgf("CadenceRunner[%s] failed", name)
				return
			}
			log.Info().Msgf("CadenceRunner[%s] execution succeeded", name)
		}
	}
}

func ExecuteSteps(t *config.TemplateData) error {
	for idx, step := range t.Steps {
		if step.Action != nil {
			err := step.Action.Execute()
			if err != nil {
				log.Error().Err(err).Msgf("Error executing step [%d]", idx+1)
				return err
			}
		}
	}
	return nil
}

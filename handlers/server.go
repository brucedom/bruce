package handlers

import (
	"bruce/config"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"sync"
)

var mutex sync.Mutex
var inProgress bool

func RunServer(t *config.TemplateData, propfile string, portNumber int) error {
	log.Debug().Msg("starting server task")
	if len(t.Variables) > 0 {
		for k, v := range t.Variables {
			log.Debug().Msgf("setting env var: %s=%s", k, v)
			os.Setenv(k, v)
		}
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", portNumber),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			defer mutex.Unlock()
			if inProgress {
				w.WriteHeader(http.StatusTooManyRequests) // 429 status code
				w.Write([]byte("Execution in Progress... please wait"))
				return
			}
			inProgress = true // Set the flag to true to indicate that an execution is starting
			go func() {       // Start the execution in a new goroutine so that it doesn't block incoming requests
				defer func() {
					mutex.Lock()       // Lock the mutex to safely update the inProgress flag
					inProgress = false // Reset the flag once the execution is complete
					mutex.Unlock()     // Unlock the mutex
				}()
				executeRunServer(w, r, propfile, t) // Execute the server logic
			}()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Execution started")) // Immediately return to the requester
		}),
	}

	log.Info().Msgf("Starting web server on port %d", portNumber)
	return srv.ListenAndServe() // Start the server
}

func executeRunServer(w http.ResponseWriter, r *http.Request, propfile string, t *config.TemplateData) {
	log.Debug().Msgf("propfile: %s", propfile)
	err := loadPropData(propfile)
	if err != nil {
		log.Error().Err(err).Msg("cannot proceed without the properties file specified.")
		os.Exit(0)
	}
	for idx, step := range t.Steps {
		if step.Action != nil {
			err := step.Action.Execute()
			if err != nil {
				log.Error().Err(err).Msgf("error executing step [%d]", idx+1)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Execution Failed"))
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Execution Succeeded"))
	log.Info().Msg("execution succeeded...")

}

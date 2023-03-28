package handlers

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type ViewResponse struct {
	Result string
}

func View(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error().Msgf("Something went wrong: " + err.Error())
		return err
	}
	log.Info().Msgf("Viewing: %s", url)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error making request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Info().Msgf(http.StatusText(resp.StatusCode))
		return fmt.Errorf(http.StatusText(resp.StatusCode))
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body")
		return err
	}
	fmt.Println(string(d))
	return nil
}

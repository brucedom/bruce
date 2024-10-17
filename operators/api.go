package operators

import (
	"bruce/exe"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/rs/zerolog/log"
	"net/http"
	"os"

	"strings"
	ttpl "text/template"
)

type API struct {
	Endpoint     string   `yaml:"api"`
	OutputFile   string   `yaml:"outputFile"`
	Method       string   `yaml:"method"`
	Body         string   `yaml:"body"`
	Headers      []string `yaml:"headers"`
	OnlyIf       string   `yaml:"onlyIf"`
	NotIf        string   `yaml:"notIf"`
	EnvId        string   `yaml:"setBodyEnv"`
	JsonEnv      string   `yaml:"setEnv"`
	JsonKey      string   `yaml:"jsonKey"`
	bodyContent  []byte
	bodyTemplate *ttpl.Template
}

// Parse JSON and retrieve value from a nested key
func (api *API) GetJsonMapValue(jsonData, key string) (string, error) {
	// Parse the JSON data into a generic map
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return "", err
	}

	// Split the key into parts for nested access
	keyParts := strings.Split(key, ".")

	// Search for the value using the key parts
	value, err := api.findNestedValue(result, keyParts)
	if err != nil {
		return "", err
	}

	// Return the value as a string if found, otherwise return an error
	if stringValue, ok := value.(string); ok {
		return stringValue, nil
	} else {
		return "", errors.New("value not found or not a string")
	}
}

// Helper function to find a nested value based on key parts
func (api *API) findNestedValue(data map[string]interface{}, keyParts []string) (interface{}, error) {
	// Base case: If there's only one key part left, return the corresponding value
	if len(keyParts) == 1 {
		if val, ok := data[keyParts[0]]; ok {
			return val, nil
		}
		return nil, errors.New("key not found")
	}

	// Recursively look for the next key in the nested map
	nextKey := keyParts[0]
	if nestedMap, ok := data[nextKey].(map[string]interface{}); ok {
		return api.findNestedValue(nestedMap, keyParts[1:])
	}
	return nil, errors.New("key not found in nested map")
}

func (api *API) Setup() {
	api.Endpoint = RenderEnvString(api.Endpoint)
	api.OutputFile = RenderEnvString(api.OutputFile)
	if len(api.Body) == 0 {
		return
	}
	// if api.body starts with file:// or https:// or http:// or s3:// then we use load template from remote, else read body as a const string to template
	if strings.HasPrefix(api.Body, "file://") || strings.HasPrefix(api.Body, "https://") || strings.HasPrefix(api.Body, "http://") || strings.HasPrefix(api.Body, "s3://") {
		t, err := loadTemplateFromRemote(api.Body)
		if err != nil {
			log.Error().Err(err).Msg("failed to load template from remote")
		} else {
			api.bodyTemplate = t
		}
	} else {
		if len(api.Body) > 0 {
			t, err := loadTemplateFromString(api.Body)
			if err != nil {
				log.Error().Err(err).Msg("failed to load template from string")
			} else {
				api.bodyTemplate = t
			}
		}
	}
	var envarMap = make(map[string]string)
	var envars = os.Environ()
	// foreach environment variable we need to load the value
	for _, e := range envars {
		if i := strings.Index(e, "="); i >= 0 {
			envarMap[e[:i]] = e[i+1:]
		}
	}
	var doc bytes.Buffer
	err := api.bodyTemplate.Execute(&doc, envarMap)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute template")
		os.Exit(1)
		return
	}
	api.bodyContent = doc.Bytes()
}

// Execute runs the command.
func (api *API) Execute() error {
	api.Setup()
	/* We do not replace command envars like the other functions, this is intended to be a raw command */
	if len(api.OnlyIf) > 0 {
		pc := exe.Run(api.OnlyIf, "")
		if pc.Failed() || len(pc.Get()) == 0 {
			log.Info().Msgf("skipping on (onlyIf): %s", api.OnlyIf)
			return nil
		}
	}
	// if notIf is set, check if it's return value is empty / false
	if len(api.NotIf) > 0 {
		pc := exe.Run(api.NotIf, "")
		if !pc.Failed() || len(pc.Get()) > 0 {
			log.Info().Msgf("skipping on (notIf): %s", api.NotIf)
			return nil
		}
	}
	if api.Method == "" {
		api.Method = "GET"
	}
	log.Info().Msgf("API request: %s %s", api.Method, api.Endpoint)
	req, err := http.NewRequest(api.Method, api.Endpoint, bytes.NewBuffer(api.bodyContent))
	if err != nil {
		log.Error().Err(err).Msg("failed to create request")
		return err
	}

	for _, h := range api.Headers {
		i := strings.Index(h, ":")
		if i > 0 {
			req.Header.Set(h[:i], h[i+1:])
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to do request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 299 {
		log.Error().Msgf("API request failed with status: %d", resp.StatusCode)
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read response body")
		return err
	}

	if api.OutputFile != "" {
		// create directories first
		err = os.MkdirAll(path.Dir(api.OutputFile), 0755)
		if err != nil {
			log.Error().Err(err).Msg("failed to create directories for api content saving")
			return err
		}
		err = os.WriteFile(api.OutputFile, d, 0644)
		if err != nil {
			log.Error().Err(err).Msg("failed to write output file")
			return err
		}
	}

	if api.EnvId != "" {
		// set env vars
		err = os.Setenv(api.EnvId, string(d))
		if err != nil {
			log.Error().Err(err).Msg("failed to set env var")
			return err
		}
	}

	if len(api.JsonEnv) > 0 && len(api.JsonKey) > 0 {
		val, err := api.GetJsonMapValue(string(d), api.JsonKey)
		if err != nil {
			log.Error().Err(err).Msg("failed to get json map value")
			return err
		}
		err = os.Setenv(api.JsonEnv, val)
		if err != nil {
			log.Error().Err(err).Msg("failed to set json env var")
			return err
		}
	}
	log.Debug().Msgf("API response: %s", string(d))
	if len(api.OutputFile) > 0 {
		log.Info().Msgf("API content saved to: %s", api.OutputFile)
		return nil
	}

	return nil
}

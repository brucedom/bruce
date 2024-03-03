package operators

import (
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Operator interface {
	Execute() error
}

type NullOperator struct {
}

func (n *NullOperator) Execute() error {
	return fmt.Errorf("invalid operator")
}

func GetValueForOSHandler(value string) string {
	log.Debug().Msgf("OS Handler value iteration: %#v", value)
	if system.Get().PackageHandler == "" {
		log.Error().Err(fmt.Errorf("cannot retrieve os handler value without a known package handler"))
		return ""
	}
	log.Debug().Msgf("testing for my package handler: %s", system.Get().PackageHandler)
	if strings.Contains(value, "|") {
		managerList := strings.Split(value, "|")
		var basePackage = ""
		var usablePackage = ""
		for _, mpkg := range managerList {
			log.Debug().Msgf("os handler iteration for manager: %#v", mpkg)
			if strings.Contains(mpkg, "=") {
				pmSplit := strings.Split(mpkg, "=")
				log.Debug().Msgf("handler [%s] specific value: %s", pmSplit[0], pmSplit[1])
				if pmSplit[0] == system.Get().PackageHandler {
					usablePackage = pmSplit[1]
				}
			} else {
				basePackage = mpkg
			}
		}
		if usablePackage != "" {
			log.Debug().Msgf("returning package manager value: %s", usablePackage)
			return usablePackage
		}
		log.Debug().Msgf("returning base value: %s", basePackage)
		return basePackage
	}
	log.Debug().Msgf("returning original value: %s", value)
	return value
}

func RenderEnvString(s string) string {
	log.Debug().Msgf("rendering env string: %s", s)
	envVars := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envVars[pair[0]] = pair[1]
	}

	var envVarRegex *regexp.Regexp
	if system.Get().OSType == "windows" {
		envVarRegex = regexp.MustCompile(`%([^%]+)%`)
	} else {
		envVarRegex = regexp.MustCompile(`(?:\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}|~)`)
	}

	return envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		if strings.HasPrefix(match, "~") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return match
			}
			return filepath.Join(homeDir, match[1:])
		}
		varName := envVarRegex.ReplaceAllString(match, "$1")
		return envVars[varName]
	})
}

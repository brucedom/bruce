package operators

import (
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
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

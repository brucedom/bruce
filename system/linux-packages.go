package system

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

func GetLinuxPackageHandler() string {
	if _, err := os.Stat("/usr/bin/dnf"); !os.IsNotExist(err) {
		log.Debug().Msg("using dnf package handler")
		return "/usr/bin/dnf"
	}
	if _, err := os.Stat("/usr/bin/yum"); !os.IsNotExist(err) {
		log.Debug().Msg("using yum package handler")
		return "/usr/bin/yum"
	}
	if _, err := os.Stat("/usr/bin/apt"); !os.IsNotExist(err) {
		log.Debug().Msg("using apt package handler")
		return "/usr/bin/apt"
	}

	log.Error().Err(fmt.Errorf("no package handler")).Msg("could not find a supported package handler for this system")
	return ""
}

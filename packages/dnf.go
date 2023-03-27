package packages

import (
	"bruce/exe"
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
)

func updateDnf() bool {
	return !exe.Run("/usr/bin/dnf update -y", false).Failed()
}

func installDnfPackage(pkg []string, isInstall bool) bool {
	action := "install"
	if !isInstall {
		action = "remove"
	}
	installCmd := fmt.Sprintf("/usr/bin/dnf %s -y %s", action, strings.Join(pkg, " "))
	log.Debug().Msgf("/usr/bin/dnf install starting with: %s", installCmd)
	install := exe.Run(installCmd, false)
	if install.Failed() {
		if len(install.Get()) > 0 {
			strSplit := strings.Split(install.Get(), "\n")
			log.Error().Err(install.GetErr())
			for _, s := range strSplit {
				log.Info().Msg(s)
			}
		}
		return false
	}
	return true
}

func installDnfRepository(name, location, key string) error {
	installCmd := fmt.Sprintf("/usr/bin/dnf config-manager --add-repo %s", location)
	log.Debug().Msgf("/usr/bin/dnf configuring repo %s", name)
	install := exe.Run(installCmd, false)
	if install.Failed() {
		if len(install.Get()) > 0 {
			strSplit := strings.Split(install.Get(), "\n")
			log.Error().Err(install.GetErr())
			for _, s := range strSplit {
				log.Info().Msg(s)
			}
		}
		return install.GetErr()
	}
	return nil
}

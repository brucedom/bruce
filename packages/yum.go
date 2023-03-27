package packages

import (
	"bruce/exe"
	"bruce/loader"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func updateYum() bool {
	return !exe.Run("/usr/bin/yum update -y", false).Failed()
}

func installYumPackage(pkg []string, isInstall bool) bool {
	action := "install"
	if !isInstall {
		action = "remove"
	}
	installCmd := fmt.Sprintf("/usr/bin/yum %s -y %s", action, strings.Join(pkg, " "))
	log.Debug().Msgf("/usr/bin/yum install starting with: %s", installCmd)
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

func installYumRepository(name, location, key string) error {
	if strings.HasSuffix(location, ".rpm") {
		err := exe.Run(fmt.Sprintf("yum install -y %s", location), false).GetErr()
		if err != nil {
			return err
		}
		return nil
	}
	f, err := os.Create(fmt.Sprintf("/etc/yum.repo.d/%s.repo", name))
	if err != nil {
		return err
	}
	defer f.Close()
	if strings.HasSuffix(location, ".repo") {
		r, _, err := loader.ReadRemoteFile(location)
		if err != nil {
			return err
		}
		l, err := f.Write(r)
		if err != nil {
			return err
		}
		log.Info().Msgf("wrote repo file with %d bytes", l)
		return nil
	}

	return nil
}

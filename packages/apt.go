package packages

import (
	"bruce/exe"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"text/template"
)

const (
	repoTpl        = `deb [arch={{.Arch}} signed-by={{.Key}}] {{.Location}} {{.Release}} stable`
	repoTplWithKey = `deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable`
)

type repoInfo struct {
	Arch     string
	Key      string
	Location string
	Release  string
}

func updateApt() bool {
	return !exe.Run("/usr/bin/apt-get update -y", false).Failed()
}

func installAptPackage(pkg []string, isInstall bool) bool {
	action := "install"
	if !isInstall {
		action = "remove"
	}
	installCmd := fmt.Sprintf("/usr/bin/apt-get %s -y %s", action, strings.Join(pkg, " "))
	log.Debug().Msgf("apt install starting with: %s", installCmd)
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

func installAptRepository(name, location, key string) error {
	os.MkdirAll("/etc/apt/keyrings", 0775)
	if key != "" {
		//curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
		// skip using x/crypto/openpgp for now...
		err := exe.Run(fmt.Sprintf("curl -fsSL %s | gpg --dearmor -o /etc/apt/keyrings/%s.gpg", key, name), false).GetErr()
		if err != nil {
			log.Error().Err(err).Msg("failed to add gpg key")
			return err
		}
		t, err := template.New("aptRepo").Parse(repoTplWithKey)
		if err != nil {
			log.Error().Err(err).Msg("failure with templates... please file an issue")
			return err
		}
		f, err := os.Create(fmt.Sprintf("/etc/apt/sources.list.d/%s.list", name))
		if err != nil {
			return err
		}
		defer f.Close()
		return t.Execute(f, &repoInfo{Arch: system.Get().OSArch, Key: key, Location: location, Release: system.Get().OsName})
	}
	t, err := template.New("aptRepo").Parse(repoTpl)
	if err != nil {
		log.Error().Err(err).Msg("failure with templates... please file an issue")
		return err
	}
	f, err := os.Create(fmt.Sprintf("/etc/apt/sources.list.d/%s.list", name))
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, &repoInfo{Arch: system.Get().OSArch, Key: key, Location: location, Release: system.Get().OsName})
}

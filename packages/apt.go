package packages

import (
	"cfs/exe"
	"cfs/loader"
	"cfs/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"text/template"
)

const (
	repoTpl        = `deb [arch={{.Arch}}] {{.Location}} {{.Release}} stable`
	repoTplWithKey = `deb [arch={{.Arch}} signed-by={{.Key}}] {{.Location}} {{.Release}} stable`
)

type repoInfo struct {
	Arch     string
	Key      string
	Location string
	Release  string
}

func updateApt() bool {
	return !exe.Run("sudo /usr/bin/apt-get update -y", "").Failed()
}

func installAptPackage(pkg []string, isInstall bool) bool {
	action := "install"
	if !isInstall {
		action = "remove"
	}
	installCmd := fmt.Sprintf("/usr/bin/apt-get %s -y %s", action, strings.Join(pkg, " "))
	log.Debug().Msgf("apt install starting with: %s", installCmd)
	install := exe.Run(installCmd, "")
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

func installAptKey(key, name string) error {
	if key != "" {
		// check first if key exists and if it does return early
		if _, err := os.Stat(fmt.Sprintf("/etc/apt/keyrings/%s.gpg", name)); err == nil {
			log.Info().Msgf("Apt key already exists: /etc/apt/keyrings/%s.gpg", name)
			return nil
		}
		tempKey := fmt.Sprintf("/tmp/%s.gpg", name)
		os.MkdirAll("/etc/apt/keyrings", 0775)
		err := loader.CopyFile(key, fmt.Sprintf(tempKey, name), 0775, true)
		if err != nil {
			log.Error().Err(err).Msg("failed to copy gpg key")
			return err
		}
		err = exe.Run(fmt.Sprintf("gpg --dearmor -o /etc/apt/keyrings/%s.gpg %s", name, tempKey), "").GetErr()
		if err != nil {
			log.Error().Err(err).Msg("failed to add gpg key")
			os.Remove(tempKey)
			return err
		}
		os.Remove(tempKey)
		log.Debug().Msgf("added key: %s", key)
	}
	return nil
}

func installAptRepository(name, location, key string, isList bool) error {
	log.Info().Msgf("Creating Apt Repository: %s", name)
	err := installAptKey(key, name)
	if err != nil {
		log.Error().Err(err).Msg("failed to add gpg key")
		return err
	}
	if isList {
		// This is a list file so we can just copy it in place and be done
		err := loader.CopyFile(location, fmt.Sprintf("/etc/apt/sources.list.d/%s.list", name), 0775, true)
		if err != nil {
			log.Error().Err(err).Msg("failed to copy repo file")
			return err
		}
		return nil
	}

	if key != "" {
		log.Info().Msgf("adding key: %s", key)
		os.MkdirAll("/etc/apt/keyrings", 0775)
		t, err := template.New("aptRepo").Parse(repoTplWithKey)
		f, err := os.Create(fmt.Sprintf("/etc/apt/sources.list.d/%s.list", name))
		if err != nil {
			log.Error().Err(err).Msg("failed to create repo file")
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
	err = t.Execute(f, &repoInfo{Arch: system.Get().OSArch, Key: key, Location: location, Release: system.Get().OsName})
	if err != nil {
		log.Error().Err(err).Msg("failed to create repo file")
		return err
	}
	// now we run apt-get update to make sure the new repo is available
	return exe.Run("/usr/bin/apt-get update -y", "").GetErr()
}

package operators

import (
	"bruce/exe"
	"bruce/packages"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
)

type Packages struct {
	PackageList []string `yaml:"packageList"`
	Action      string   `yaml:"action"`
	OsLimits    string   `yaml:"osLimits"`
	OnlyIf      string   `yaml:"onlyIf"`
	NotIf       string   `yaml:"notIf"`
}

func (p *Packages) Setup() {
	newList := make([]string, 0)
	for _, pkg := range p.PackageList {
		newList = append(newList, RenderEnvString(pkg))
	}
	p.PackageList = newList
}

func (p *Packages) Execute() error {
	p.Setup()
	if system.Get().CanExecOnOs(p.OsLimits) {
		if len(p.OnlyIf) > 0 {
			pc := exe.Run(p.OnlyIf, "")
			if pc.Failed() || len(pc.Get()) == 0 {
				log.Info().Msgf("skipping on (onlyIf): %s", p.OnlyIf)
				return nil
			}
		}
		// if notIf is set, check if it's return value is empty / false
		if len(p.NotIf) > 0 {
			pc := exe.Run(p.NotIf, "")
			if !pc.Failed() || len(pc.Get()) > 0 {
				log.Info().Msgf("skipping on (notIf): %s", p.NotIf)
				return nil
			}
		}
		log.Info().Msgf("starting package installs for %s", system.Get().PackageHandler)
		isInstall := true
		if p.Action == "remove" {
			isInstall = false
		}
		success := packages.InstallOSPackage(p.PackageList, system.Get().PackageHandler, isInstall)
		if !success {
			err := fmt.Errorf("cannot install packages: %s", p.PackageList)
			log.Error().Err(err).Msg("package install failed")
			return err
		}
		return nil
	} else {
		si := system.Get()
		log.Debug().Msgf("System (%s|%s) limited execution of installs for: %s", si.OSID, si.OSVersionID, p.OsLimits)
	}
	return nil
}

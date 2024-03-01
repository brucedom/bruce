package operators

import (
	"cfs/exe"
	"cfs/packages"
	"cfs/system"
	"fmt"
	"github.com/rs/zerolog/log"
)

type PackageRepo struct {
	Name       string `yaml:"repoName"`
	Location   string `yaml:"repoLocation"`
	RType      string `yaml:"repoType"`
	Key        string `yaml:"repoKey"`
	IsRepoFile bool   `yaml:"isRepoFile"`
	OsLimits   string `yaml:"osLimits"`
	OnlyIf     string `yaml:"onlyIf"`
	NotIf      string `yaml:"notIf"`
}

func (p *PackageRepo) Setup() {
	p.Location = RenderEnvString(p.Location)
	p.Key = RenderEnvString(p.Key)
}

func (p *PackageRepo) Execute() error {
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
		log.Debug().Msgf("starting package repo configuration for %s", p.RType)
		err := p.InstallPreReq()
		if err != nil {
			return err
		}
		return packages.InstallRepository(p.RType, p.Name, p.Location, p.Key, p.IsRepoFile)
	} else {
		si := system.Get()
		log.Debug().Msgf("System (%s|%s) limited execution of installs for: %s", si.OSID, si.OSVersionID, p.OsLimits)
	}
	return nil
}

func (p *PackageRepo) InstallPreReq() error {
	preReq := ""
	switch p.RType {
	case "dnf":
		preReq = "dnf-plugins-core"
	default:
		preReq = ""
	}

	if preReq == "" {
		return nil
	}
	success := packages.InstallOSPackage([]string{preReq}, p.RType, true)
	if !success {
		err := fmt.Errorf("cannot install pre-requisite package: %s", preReq)
		log.Error().Err(err).Msg("failed repository pre-requisites")
		return err
	}
	return nil
}

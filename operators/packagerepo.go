package operators

import (
	"bruce/packages"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
)

type PackageRepo struct {
	Name     string `yaml:"repoName"`
	Location string `yaml:"repoLocation"`
	RType    string `yaml:"repoType"`
	Key      string `yaml:"repoKey"`
	OsLimits string `yaml:"osLimits"`
}

func (p *PackageRepo) Execute() error {
	if system.Get().CanExecOnOs(p.OsLimits) {
		log.Info().Msgf("starting package repo configuration for %s", p.RType)
		err := p.InstallPreReq()
		if err != nil {
			return err
		}
		return packages.InstallRepository(p.RType, p.Name, p.Location, p.Key)
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

	success := packages.InstallOSPackage([]string{preReq}, p.RType, true)
	if !success {
		err := fmt.Errorf("cannot install pre-requisite package: %s", preReq)
		log.Error().Err(err).Msg("failed repository pre-requisites")
		return err
	}
	return nil
}

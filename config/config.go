package config

import (
	"bruce/loader"
	"bruce/operators"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

// TemplateData will be marshalled from the provided config file that exists.
type TemplateData struct {
	Steps     []Steps `yaml:"steps"`
	BackupDir string
}

// Steps include multiple action operators to be executed per step
type Steps struct {
	Name   string             `yaml:"name"`
	Action operators.Operator `yaml:"action"`
}

// TODO: Add UnmarshallJSON

// UnmarshalYAML Implements the Unmarshaler interface of the yaml pkg.
func (e *Steps) UnmarshalYAML(nd *yaml.Node) error {
	// TODO: Fix this is the near future. (maybe plugin based?)

	crn := &operators.Cron{}
	if err := nd.Decode(crn); err == nil && len(crn.Schedule) > 0 {
		log.Debug().Msg("matching cron operator")
		e.Action = crn
		return nil
	}

	co := &operators.Command{}
	if err := nd.Decode(co); err == nil && len(co.Cmd) > 0 {
		log.Debug().Msg("matching command operator")
		e.Action = co
		return nil
	}

	tb := &operators.Tarball{}
	if err := nd.Decode(tb); err == nil && len(tb.Src) > 0 {
		log.Debug().Msg("matching tarball operator")
		e.Action = tb
		return nil
	}

	cp := &operators.Copy{}
	if err := nd.Decode(cp); err == nil && len(cp.Src) > 0 {
		log.Debug().Msg("matching copy operator")
		e.Action = cp
		return nil
	}

	to := &operators.Template{}
	if err := nd.Decode(to); err == nil && len(to.Template) > 0 {
		log.Debug().Msg("matching template operator")
		e.Action = to
		return nil
	}

	pr := &operators.PackageRepo{}
	if err := nd.Decode(pr); err == nil && len(pr.Location) > 0 {
		log.Debug().Msg("matching package repository operator")
		e.Action = pr
		return nil
	}

	pl := &operators.Packages{}
	if err := nd.Decode(pl); err == nil && len(pl.PackageList) > 0 {
		log.Debug().Msg("matching package operator")
		e.Action = pl
		return nil
	}

	svc := &operators.Services{}
	if err := nd.Decode(svc); err == nil && len(svc.Service) > 0 {
		log.Debug().Msg("matching service operator")
		e.Action = svc
		return nil
	}
	e.Action = &operators.NullOperator{}
	return nil
}

// LoadConfig attempts to load the user provided manifest.
func LoadConfig(fileName string) (*TemplateData, error) {
	d, _, err := loader.ReadRemoteFile(fileName)
	if err != nil {
		log.Error().Err(err).Msg("cannot proceed without a config file and specified config cannot be read.")
		os.Exit(1)
	}
	log.Debug().Bytes("rawConfig", d)
	c := &TemplateData{}

	err = yaml.Unmarshal(d, c)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse config file")
	}
	return c, nil
}

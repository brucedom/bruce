package operators

import (
	"cfs/exe"
	"cfs/loader"
	"github.com/rs/zerolog/log"
	"io/fs"
)

type Copy struct {
	Src    string      `yaml:"copy"`
	Dest   string      `yaml:"dest"`
	Perm   fs.FileMode `yaml:"perm"`
	OnlyIf string      `yaml:"onlyIf"`
	NotIf  string      `yaml:"notIf"`
}

func (c *Copy) Setup() {
	c.Src = RenderEnvString(c.Src)
	c.Dest = RenderEnvString(c.Dest)
}

func (c *Copy) Execute() error {
	c.Setup()
	if len(c.OnlyIf) > 0 {
		pc := exe.Run(c.OnlyIf, "")
		if pc.Failed() || len(pc.Get()) == 0 {
			log.Info().Msgf("skipping on (onlyIf): %s", c.OnlyIf)
			return nil
		}
	}
	// if notIf is set, check if it's return value is empty / false
	if len(c.NotIf) > 0 {
		pc := exe.Run(c.NotIf, "")
		if !pc.Failed() || len(pc.Get()) > 0 {
			log.Info().Msgf("skipping on (notIf): %s", c.NotIf)
			return nil
		}
	}
	err := loader.CopyFile(c.Src, c.Dest, c.Perm, true)
	log.Info().Msgf("copy: %s => %s", c.Src, c.Dest)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
		return err
	}
	return nil
}

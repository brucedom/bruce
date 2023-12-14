package operators

import (
	"cfs/loader"
	"github.com/rs/zerolog/log"
	"io/fs"
)

type Copy struct {
	Src  string      `yaml:"copy"`
	Dest string      `yaml:"dest"`
	Perm fs.FileMode `yaml:"perm"`
}

func (c *Copy) Setup() {
	c.Src = RenderEnvString(c.Src)
	c.Dest = RenderEnvString(c.Dest)
}

func (c *Copy) Execute() error {
	c.Setup()
	err := loader.CopyFile(c.Src, c.Dest, c.Perm, true)
	log.Info().Msgf("copy: %s => %s", c.Src, c.Dest)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
		return err
	}
	return nil
}

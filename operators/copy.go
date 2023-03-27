package operators

import (
	"bruce/exe"
	"bruce/loader"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"os"
	"path"
)

type Copy struct {
	Src  string      `yaml:"copy"`
	Dest string      `yaml:"dest"`
	Perm fs.FileMode `yaml:"perm"`
}

func (c *Copy) Execute() error {
	if len(c.Src) < 1 {
		return fmt.Errorf("source is too short")
	}
	source, _, err := loader.GetRemoteReader(c.Src)
	if err != nil {
		log.Error().Err(err).Msg("cannot open source file")
		return err
	}
	defer source.Close()

	if exe.FileExists(c.Dest) {
		exe.DeleteFile(c.Dest)
	} else {
		// check if the directories exist to render the file
		if !exe.FileExists(path.Dir(c.Dest)) {
			os.MkdirAll(path.Dir(c.Dest), 0775)
		}
	}

	destination, err := os.OpenFile(c.Dest, os.O_RDWR|os.O_CREATE, c.Perm)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file for writing copy: %s", c.Dest)
		return err
	}
	defer destination.Close()

	log.Debug().Str("copy", c.Src).Msg("preparing to execute")

	len, err := io.Copy(destination, source)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
	}
	log.Info().Msgf("copied %d bytes", len)
	return nil
}

package operators

import (
	"cfs/loader"
	"github.com/rs/zerolog/log"
	"io/fs"
	"os"
)

type RecursiveCopy struct {
	Src       string      `yaml:"copyRecursive"`
	Dest      string      `yaml:"dest"`
	Mode      string      `yaml:"mode"`
	ParentDir string      `yaml:"parentDir"`
	Perm      fs.FileMode `yaml:"perm"`
}

func (c *RecursiveCopy) Setup() {
	c.Src = RenderEnvString(c.Src)
	c.Dest = RenderEnvString(c.Dest)
	c.ParentDir = RenderEnvString(c.ParentDir)
	// Check if parent directory exists and create it if it doesn't
	if _, err := os.Stat(c.ParentDir); os.IsNotExist(err) {
		err = os.MkdirAll(c.ParentDir, 0755)
		if err != nil {
			log.Error().Err(err).Msg("failed to create parent directory for recursive copy")
		}
	}

}

func (c *RecursiveCopy) Execute() error {
	c.Setup()
	err := loader.CopyFile(c.Src, c.Dest, c.Perm, true)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
		return err
	}
	return nil
}

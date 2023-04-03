package operators

import (
	"cfs/loader"
	"github.com/rs/zerolog/log"
	"os"
)

type RecursiveCopy struct {
	Src      string   `yaml:"copyRecursive"`
	Dest     string   `yaml:"dest"`
	Ignores  []string `yaml:"ignoreFiles"`
	FlatCopy bool     `yaml:"flatCopy"`
	MaxDepth int      `yaml:"maxDepth"`
}

func (c *RecursiveCopy) Setup() {
	c.Dest = RenderEnvString(c.Dest)
	// Check if parent directory exists and create it if it doesn't
	if _, err := os.Stat(c.Dest); os.IsNotExist(err) {
		err = os.MkdirAll(c.Dest, 0755)
		if err != nil {
			log.Error().Err(err).Msg("failed to create parent directory for recursive copy")
		}
	}

}

func (c *RecursiveCopy) Execute() error {
	c.Setup()
	log.Info().Msgf("recursive copying %s to %s", c.Src, c.Dest)
	err := loader.RecursiveCopy(c.Src, c.Dest, c.Dest, true, c.Ignores, c.FlatCopy, c.MaxDepth)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
		return err
	}
	return nil
}

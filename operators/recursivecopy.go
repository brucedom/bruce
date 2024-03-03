package operators

import (
	"bruce/exe"
	"bruce/loader"
	"github.com/rs/zerolog/log"
	"os"
)

type RecursiveCopy struct {
	Src           string   `yaml:"copyRecursive"`
	Dest          string   `yaml:"dest"`
	Ignores       []string `yaml:"ignoreFiles"`
	FlatCopy      bool     `yaml:"flatCopy"`
	MaxDepth      int      `yaml:"maxDepth"`
	MaxConcurrent int      `yaml:"maxConcurrent"`
	OnlyIf        string   `yaml:"onlyIf"`
	NotIf         string   `yaml:"notIf"`
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
	if c.MaxConcurrent == 0 {
		c.MaxConcurrent = 5
	}
}

func (c *RecursiveCopy) Execute() error {
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
	log.Info().Msgf("rcopy (%d files at a time) with a maxDepth of: %d", c.MaxConcurrent, c.MaxDepth)
	log.Info().Msgf("  %s => %s", c.Src, c.Dest)
	err := loader.RecursiveCopy(c.Src, c.Dest, c.Dest, true, c.Ignores, c.FlatCopy, c.MaxDepth, c.MaxConcurrent)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
		return err
	}
	return nil
}

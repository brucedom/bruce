package operators

import (
	"bruce/exe"
	"bruce/mutation"
	"fmt"
	"github.com/rs/zerolog/log"
)

type Tarball struct {
	Name   string `yaml:"name"`
	Src    string `yaml:"tarball"`
	Dest   string `yaml:"dest"`
	Force  bool   `yaml:"force"`
	Strip  bool   `yaml:"stripRoot"`
	OnlyIf string `yaml:"onlyIf"`
	NotIf  string `yaml:"notIf"`
}

func (t *Tarball) Setup() {
	t.Src = RenderEnvString(t.Src)
	t.Dest = RenderEnvString(t.Dest)
}

func (t *Tarball) Execute() error {
	t.Setup()
	if len(t.OnlyIf) > 0 {
		pc := exe.Run(t.OnlyIf, "")
		if pc.Failed() || len(pc.Get()) == 0 {
			log.Info().Msgf("skipping on (onlyIf): %s", t.OnlyIf)
			return nil
		}
	}
	// if notIf is set, check if it's return value is empty / false
	if len(t.NotIf) > 0 {
		pc := exe.Run(t.NotIf, "")
		if !pc.Failed() || len(pc.Get()) > 0 {
			log.Info().Msgf("skipping on (notIf): %s", t.NotIf)
			return nil
		}
	}
	if len(t.Src) < 1 {
		return fmt.Errorf("source is too short")
	}
	log.Info().Msgf("tarball: %s => %s", t.Src, t.Dest)
	return mutation.ExtractTarball(t.Src, t.Dest, t.Force, t.Strip)
}

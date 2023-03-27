package operators

import (
	"bruce/mutation"
	"fmt"
)

type Tarball struct {
	Name  string `yaml:"name"`
	Src   string `yaml:"tarball"`
	Dest  string `yaml:"dest"`
	Force bool   `yaml:"force"`
	Strip bool   `yaml:"stripRoot"`
}

func (t *Tarball) Execute() error {
	if len(t.Src) < 1 {
		return fmt.Errorf("source is too short")
	}

	return mutation.ExtractTarball(t.Src, t.Dest, t.Force, t.Strip)
}

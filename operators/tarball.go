package operators

import (
	"cfs/mutation"
	"fmt"
)

type Tarball struct {
	Name  string `yaml:"name"`
	Src   string `yaml:"tarball"`
	Dest  string `yaml:"dest"`
	Force bool   `yaml:"force"`
	Strip bool   `yaml:"stripRoot"`
}

func (t *Tarball) Setup() {
	t.Src = RenderEnvString(t.Src)
	t.Dest = RenderEnvString(t.Dest)
}

func (t *Tarball) Execute() error {
	t.Setup()
	if len(t.Src) < 1 {
		return fmt.Errorf("source is too short")
	}
	return mutation.ExtractTarball(t.Src, t.Dest, t.Force, t.Strip)
}

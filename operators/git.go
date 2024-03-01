package operators

import (
	"cfs/exe"
	"cfs/system"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"
	"os"
	"path"
)

type Git struct {
	Repo     string `yaml:"gitRepo"`
	Location string `yaml:"dest"`
	OsLimits string `yaml:"osLimits"`
	OnlyIf   string `yaml:"onlyIf"`
	NotIf    string `yaml:"notIf"`
}

func (g *Git) Setup() {
	g.Repo = RenderEnvString(g.Repo)
	g.Location = RenderEnvString(g.Location)
	// make the destination directory without the last path element
	target := path.Dir(g.Location)
	err := os.MkdirAll(target, 0755)
	if err != nil {
		log.Error().Err(err).Msg("failed to create git destination directory for git clone")
	}
}

// Execute runs the command.
func (g *Git) Execute() error {
	g.Setup()
	/* We do not replace command envars like the other functions, this is intended to be a raw command */
	if system.Get().CanExecOnOs(g.OsLimits) {
		if len(g.OnlyIf) > 0 {
			pc := exe.Run(g.OnlyIf, "")
			if pc.Failed() || len(pc.Get()) == 0 {
				log.Info().Msgf("skipping on (onlyIf): %s", g.OnlyIf)
				return nil
			}
		}
		// if notIf is set, check if it's return value is empty / false
		if len(g.NotIf) > 0 {
			pc := exe.Run(g.NotIf, "")
			if !pc.Failed() || len(pc.Get()) > 0 {
				log.Info().Msgf("skipping on (notIf): %s", g.NotIf)
				return nil
			}
		}
		// if directory exists and it contains a .git directory, just return
		if _, err := os.Stat(path.Join(g.Location, ".git")); err == nil {
			log.Info().Msgf("git repo already exists: %s", g.Location)
			return nil
		}

		_, err := git.PlainClone(g.Location, false, &git.CloneOptions{
			URL:      g.Repo,
			Progress: os.Stdout,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to clone repo")
			return err
		}
		log.Info().Msgf("git cloned: %s to %s", g.Repo, g.Location)
	} else {
		log.Info().Str("git", g.Repo).Msgf("skipped due to os limit: %s", g.OsLimits)
	}
	return nil
}

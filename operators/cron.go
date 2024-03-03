package operators

import (
	"bruce/exe"
	"bruce/mutation"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"runtime"
)

// Cron provides a means to set the ownership of files or directories as needed.
type Cron struct {
	Name     string `yaml:"cron"`
	Schedule string `yaml:"schedule"`
	User     string `yaml:"username"`
	Exec     string `yaml:"cmd"`
	OnlyIf   string `yaml:"onlyIf"`
	NotIf    string `yaml:"notIf"`
}

func (c *Cron) Setup() {
	c.Exec = RenderEnvString(c.Exec)
	c.User = RenderEnvString(c.User)
}

func (c *Cron) Execute() error {
	c.Setup()
	if runtime.GOOS == "linux" {
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
		jobName := mutation.StripNonAlnum(c.Name)
		log.Info().Msgf("cron: /etc/cron.d/%s", jobName)
		c.Schedule = mutation.StripExtraWhitespaceFB(c.Schedule)
		c.User = mutation.StripNonAlnum(c.User)
		log.Debug().Msgf("starting cronjob: %s", jobName)
		if c.User == "" {
			c.User = system.Get().CurrentUser.Username
		}
		return mutation.WriteInlineTemplate(fmt.Sprintf("/etc/cron.d/%s", jobName), "{{.Schedule}} {{.User}} {{.Exec}}", c)
	}
	return fmt.Errorf("not supported")
}

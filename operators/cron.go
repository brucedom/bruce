package operators

import (
	"cfs/mutation"
	"cfs/system"
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
}

func (c *Cron) Setup() {
	c.Exec = RenderEnvString(c.Exec)
	c.User = RenderEnvString(c.User)
}

func (c *Cron) Execute() error {
	c.Setup()
	if runtime.GOOS == "linux" {
		jobName := mutation.StripNonAlnum(c.Name)
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

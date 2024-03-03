package operators

import (
	"bruce/exe"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type Command struct {
	Cmd        string `yaml:"cmd"`
	WorkingDir string `yaml:"dir"`
	OsLimits   string `yaml:"osLimits"`
	SetEnv     string `yaml:"setEnv"`
	OnlyIf     string `yaml:"onlyIf"`
	NotIf      string `yaml:"notIf"`
	EnvCmd     string
}

func (c *Command) Setup() {
	c.WorkingDir = RenderEnvString(c.WorkingDir)
	c.EnvCmd = RenderEnvString(c.Cmd)
}

// Execute runs the command.
func (c *Command) Execute() error {
	c.Setup()
	/* We do not replace command envars like the other functions, this is intended to be a raw command */
	if system.Get().CanExecOnOs(c.OsLimits) {
		// if onlyIf is set, check if it's return value is not empty / true
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
		if len(c.EnvCmd) < 1 {
			return fmt.Errorf("no command to execute")
		}
		log.Info().Msgf("cmd: %s", c.EnvCmd)
		fileName := exe.EchoToFile(c.EnvCmd, os.TempDir())
		// change directory to the working directory if specified
		err := os.Chmod(fileName, 0775)
		if err != nil {
			log.Error().Err(err).Msg("temp file must exist to continue")
			return err
		}
		log.Debug().Str("command", c.EnvCmd).Msgf("executing local file: %s", fileName)
		pc := exe.Run(fileName, c.WorkingDir)
		if pc.Failed() {
			log.Error().Err(pc.GetErr()).Msg(pc.Get())
			return pc.GetErr()
		} else {
			log.Debug().Str("cmd", c.EnvCmd).Msgf("completed executing: %s", fileName)
			log.Debug().Msgf("Output: %s", pc.Get())
			if len(c.SetEnv) > 0 {
				log.Debug().Str("cmd", c.EnvCmd).Msgf("setting env var: %s=%s", c.SetEnv, pc.Get())
				log.Error().Err(os.Setenv(c.SetEnv, pc.Get()))
			}
			log.Error().Err(os.Remove(fileName))
		}
	} else {
		log.Info().Str("cmd", c.EnvCmd).Msgf("skipped due to os limit: %s", c.OsLimits)
	}
	return nil
}

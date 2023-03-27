package operators

import (
	"bruce/exe"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type Command struct {
	Cmd      string `yaml:"cmd"`
	OsLimits string `yaml:"osLimits"`
}

// Execute runs the command.
func (c *Command) Execute() error {
	if system.Get().CanExecOnOs(c.OsLimits) {
		if len(c.Cmd) < 1 {
			return fmt.Errorf("no command to execute")
		}
		log.Debug().Str("cmd", c.Cmd).Msg("preparing to execute")
		fileName := exe.EchoToFile(c.Cmd, os.TempDir())
		err := os.Chmod(fileName, 0775)
		if err != nil {
			log.Error().Err(err).Msg("temp file must exist to continue")
			return err
		}
		log.Debug().Str("command", c.Cmd).Msgf("executing local file: %s", fileName)
		pc := exe.Run(fileName, false)
		if pc.Failed() {
			log.Error().Err(pc.GetErr()).Msg(pc.Get())
			return pc.GetErr()
		} else {
			log.Info().Str("cmd", c.Cmd).Msgf("completed executing: %s", fileName)
			log.Debug().Msgf("Output: %s", pc.Get())
			os.Remove(fileName)
		}
	} else {
		log.Info().Str("cmd", c.Cmd).Msgf("skipped due to os limit: %s", c.OsLimits)
	}
	return nil
}

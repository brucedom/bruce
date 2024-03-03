package operators

import (
	"cfs/exe"
	"cfs/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type Loop struct {
	LoopScript string `yaml:"loopScript"`
	Count      int    `yaml:"count"`
	Variable   string `yaml:"var"`
	OsLimits   string `yaml:"osLimits"`
	OnlyIf     string `yaml:"onlyIf"`
	NotIf      string `yaml:"notIf"`
}

func (lp *Loop) Setup() {

}

// Execute runs the command.
func (lp *Loop) Execute() error {
	lp.Setup()
	/* We do not replace command envars like the other functions, this is intended to be a raw command */
	if system.Get().CanExecOnOs(lp.OsLimits) {
		// if onlyIf is set, check if it's return value is not empty / true
		if len(lp.OnlyIf) > 0 {
			pc := exe.Run(lp.OnlyIf, "")
			if pc.Failed() || len(pc.Get()) == 0 {
				log.Info().Msgf("skipping on (onlyIf): %s", lp.OnlyIf)
				return nil
			}
		}
		// if notIf is set, check if it's return value is empty / false
		if len(lp.NotIf) > 0 {
			pc := exe.Run(lp.NotIf, "")
			if !pc.Failed() || len(pc.Get()) > 0 {
				log.Info().Msgf("skipping on (notIf): %s", lp.NotIf)
				return nil
			}
		}
		for i := 0; i < lp.Count; i++ {
			log.Info().Str("loop", lp.LoopScript).Msgf("executing: %s with variable: %s and value: %d", lp.LoopScript, lp.Variable, i)
			log.Error().Err(os.Setenv(lp.Variable, fmt.Sprintf("%d", i)))
			// get current running file and append the loop script as the first argument
			execCmd := fmt.Sprintf("%s %s", os.Args[0], lp.LoopScript)
			pc := exe.Run(execCmd, "")
			if pc.Failed() {
				log.Error().Err(pc.GetErr()).Msg(pc.Get())
				return pc.GetErr()
			}
		}
	} else {
		log.Info().Str("loop", lp.LoopScript).Msgf("skipped due to os limit: %s", lp.OsLimits)
	}
	return nil
}

package operators

import (
	"bytes"
	"cfs/exe"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Signals struct {
	PidFile string `yaml:"pidFile"`
	Signal  string `yaml:"signal"`
	OnlyIf  string `yaml:"onlyIf"`
	NotIf   string `yaml:"notIf"`
}

func (s *Signals) Execute() error {
	if len(s.OnlyIf) > 0 {
		pc := exe.Run(s.OnlyIf, "")
		if pc.Failed() || len(pc.Get()) == 0 {
			log.Info().Msgf("skipping on (onlyIf): %s", s.OnlyIf)
			return nil
		}
	}
	// if notIf is set, check if it's return value is empty / false
	if len(s.NotIf) > 0 {
		pc := exe.Run(s.NotIf, "")
		if !pc.Failed() || len(pc.Get()) > 0 {
			log.Info().Msgf("skipping on (notIf): %s", s.NotIf)
			return nil
		}
	}
	if _, err := os.Stat(s.PidFile); os.IsNotExist(err) {
		err = fmt.Errorf("pidfile does not exist at: %s", s.PidFile)
		return err
	}
	d, err := os.ReadFile(s.PidFile)
	if err != nil {
		log.Error().Err(err).Msg("pid file read error")
		return err
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(d)))
	if err != nil {
		log.Error().Err(err).Msgf("could not reading pid file: %s", s.PidFile)
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		log.Error().Err(err).Msgf("could not find process for pid: %d", pid)
		return err
	}
	switch strings.ToUpper(s.Signal) {
	case "SIGINT":
		p.Signal(syscall.SIGINT)
		return nil
	case "SIGHUP":
		p.Signal(syscall.SIGHUP)
		return nil
	default:
		p.Signal(syscall.SIGHUP)
		return nil
	}
	return nil
}

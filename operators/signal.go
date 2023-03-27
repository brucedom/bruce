package operators

import (
	"bytes"
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
}

func (s *Signals) Execute() error {
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

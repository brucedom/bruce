package operators

import (
	"bruce/rssh"
	"github.com/rs/zerolog/log"
	"os"
	"os/user"
	"strings"
)

type RemoteExec struct {
	ExecCmd string `yaml:"remoteCmd"`
	RemHost string `yaml:"host"`
	SetEnv  string `yaml:"setEnv"`
	PrivKey string `yaml:"key"`
	OnlyIf  string `yaml:"onlyIf"`
	NotIf   string `yaml:"notIf"`
}

func (re *RemoteExec) Setup() {
	re.ExecCmd = RenderEnvString(re.ExecCmd)
	re.RemHost = RenderEnvString(re.RemHost)
	re.OnlyIf = RenderEnvString(re.OnlyIf)
	re.NotIf = RenderEnvString(re.NotIf)
}

func (re *RemoteExec) Execute() error {
	re.Setup()
	usr, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current user")
		return err
	}
	uname := usr.Username
	hostname := re.RemHost
	if strings.Contains(re.RemHost, "@") {
		uname = strings.Split(re.RemHost, "@")[0]
		hostname = strings.Split(re.RemHost, "@")[1]
	}
	rs, err := rssh.NewRSSH(hostname, uname, re.PrivKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create RSSH")
		return err
	}
	defer rs.Close()
	if len(re.OnlyIf) > 0 {
		oif, err := rs.ExecCommand(re.OnlyIf)
		if err != nil || len(oif) == 0 {
			log.Info().Msgf("remoteCmd skipping on (onlyIf): %s", re.ExecCmd)
			return nil
		}
	}
	// if notIf is set, check if it's return value is empty / false
	if len(re.NotIf) > 0 {
		nif, err := rs.ExecCommand(re.NotIf)
		if err == nil || len(nif) > 0 {
			log.Info().Msgf("remoteCmd skipping on (notIf): %s", re.ExecCmd)
			return nil
		}
	}
	log.Info().Msgf("remoteCmd: %s", re.ExecCmd)
	output, err := rs.ExecCommand(re.ExecCmd)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to execute %s", re.ExecCmd)
		return err
	} else {
		log.Debug().Str("cmd", re.ExecCmd).Msgf("completed executing on [%s]", re.RemHost)
		log.Debug().Msgf("Output: %s", output)
		if len(re.SetEnv) > 0 {
			log.Debug().Str("remoteCmd", re.SetEnv).Msgf("setting env var: %s=%s", re.SetEnv, re.ExecCmd)
			log.Error().Err(os.Setenv(re.SetEnv, re.ExecCmd))
		}
	}
	return nil
}

package operators

import (
	"bruce/exe"
	"bruce/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
)

//

type Services struct {
	Service         string   `yaml:"service"`
	Enabled         bool     `yaml:"setEnabled"`
	State           string   `yaml:"state"`
	RestartOnUpdate []string `yaml:"restartTrigger"`
	RestartAlways   bool     `yaml:"restartAlways"`
	OsLimits        string   `yaml:"osLimits"`
}

func (s *Services) Execute() error {
	si := system.Get()
	if si.OSType == "linux" {
		if si.CanExecOnOs(s.OsLimits) {
			return s.ExecuteLinux(si)
		}
	}
	return nil
}

func (s *Services) ExecuteLinux(si *system.SystemInfo) error {
	doDaemonReload := false
	for _, tpl := range si.ModifiedTemplates {
		if strings.Contains(tpl, "systemd") {
			doDaemonReload = true
		}
	}
	if doDaemonReload {
		log.Debug().Msgf("daemon reload required due to service change")
		exe.Run("systemctl daemon-reload", false)
	}
	// We only support sytemd / systemctrl for right now...

	status := exe.Run(fmt.Sprintf("systemctl is-active %s", s.Service), false).Get()
	if strings.Contains(strings.ToLower(status), "could not be found") {
		err := fmt.Errorf("%s service not found", s.Service)
		log.Error().Err(err).Msg("service does not exist cannot manage state")
		return err
	}
	if s.Enabled {
		// test if not enabled
		curState := exe.Run(fmt.Sprintf("systemctl is-enabled %s", s.Service), false).Get()
		if curState != "enabled" {
			eno := exe.Run(fmt.Sprintf("systemctl enable %s --now", s.Service), false).Get()
			log.Info().Str("output", eno).Msgf("set enabled for %s", s.Service)
		}
	}

	if s.State == "started" {
		if status != "active" {
			out := exe.Run(fmt.Sprintf("systemctl restart %s", s.Service), false).Get()
			log.Info().Str("output", out).Msgf("issued restart to inactive service: %s", s.Service)
		}
	}
	if s.State == "stopped" {
		if status != "inactive" {
			out := exe.Run(fmt.Sprintf("systemctl stop %s", s.Service), false).Get()
			log.Info().Str("output", out).Msgf("issued stop to active service: %s", s.Service)
		}
	}
	if s.RestartAlways {
		out := exe.Run(fmt.Sprintf("systemctl restart %s", s.Service), false).Get()
		log.Info().Str("output", out).Msgf("issued restart (always) to service: %s", s.Service)
	} else {
		for _, resTemp := range s.RestartOnUpdate {
			shouldRestart := false
			for _, modT := range si.ModifiedTemplates {
				if resTemp == modT {
					shouldRestart = true
				}
			}
			if shouldRestart {
				out := exe.Run(fmt.Sprintf("systemctl restart %s", s.Service), false).Get()
				log.Info().Str("output", out).Msgf("issued restart (modified by template) to service: %s", s.Service)
			}
		}
	}
	// finally we recheck to see if it is started as we may have to revert
	status = exe.Run(fmt.Sprintf("systemctl is-active %s", s.Service), false).Get()
	if s.State == "started" {
		if status != "active" {
			return fmt.Errorf("service [%s] is in an invalid state", s.Service)
		}
	}
	return nil
}

package services

/*

func StartOSServiceReloads() []string {
	var failedSvcs []string
	// TODO: Execute services and aggregate the list of ones that fail here
	cfg := config.Get()
	if cfg.SystemType == "linux" {
		// We only support sytemd / systemctrl for right now...
		for _, svc := range cfg.Template.Reloadables {
			if strings.ToLower(svc.RType) == "systemd" {
				out := exe.Run(fmt.Sprintf("systemctl restart %s", svc.Name), cfg.TrySudo).Get()
				log.Info().Str("output", out).Msgf("issued restart (update event) to service: %s", svc.Name)
				status := exe.Run(fmt.Sprintf("systemctl is-active %s", svc.Name), cfg.TrySudo).Get()
				if strings.Contains(strings.ToLower(status), "could not be found") {
					log.Error().Err(fmt.Errorf("%s service not found", svc.Name)).Msg("service does not exist cannot manage state")
					continue
				}
				if status != "active" {
					log.Error().Msgf("failed to restart: %s for updates", svc.Name)
				}
			}
			if strings.ToLower(svc.RType) == "signal" {
				// Realistically we just send a signal should we validate this somehow later?
				SendSignal(svc)
			}
		}
	}
	return failedSvcs
}

func SendSignal(s config.Reloads) error {
	d, err := os.ReadFile(s.Pid)
	if err != nil {
		log.Error().Err(err).Msg("pid file error")
		return err
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(d)))
	if err != nil {
		log.Error().Err(err).Msgf("could not reading pid file: %s", s.Pid)
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
	default:
		p.Signal(syscall.SIGHUP)
		return nil
	}
}

func RestoreFailedServices(svcs []string) error {
	for _, svc := range svcs {
		for _, cs := range config.Get().Template.Services {
			if svc == cs.Name {
				for _, srcName := range cs.RestartOnUpdate {
					log.Info().Msgf("restoring template %s", srcName)
					err := templates.RestoreBackupFile(srcName)
					if err != nil {
						log.Error().Err(err).Msg("could not restore template")
					}
				}
			}
		}
	}
	StartOSServiceExecution()
	return nil
}
*/

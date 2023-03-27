package packages

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
)

func InstallOSPackage(pkgs []string, packageHandler string, isInstall bool) bool {
	if len(pkgs) < 1 {
		log.Error().Err(fmt.Errorf("can't install nothing"))
		return false
	}
	switch packageHandler {
	case "apt":
		return installAptPackage(GetManagerPackages(pkgs, "apt"), isInstall)
	case "yum":
		return installYumPackage(GetManagerPackages(pkgs, "yum"), isInstall)
	case "dnf":
		return installDnfPackage(GetManagerPackages(pkgs, "dnf"), isInstall)
	}
	log.Info().Msg("no package manager to check for installed package")
	return false
}

func InstallRepository(rType, name, loc, key string) error {
	var err error
	switch rType {
	case "dnf":
		err = installDnfRepository(name, loc, key)
	case "apt":
		err = installAptRepository(name, loc, key)
	case "yum":
		err = installYumRepository(name, loc, key)
	default:
		err = fmt.Errorf("no supported package manager")
	}
	if err != nil {
		return err
	}
	DoPackageManagerUpdate(rType)
	return nil
}

func GetManagerPackages(pkgs []string, manager string) []string {
	// TODO: Do we want to honor manager to install since we have os limits now?
	var newList []string
	for _, pkg := range pkgs {
		log.Debug().Msgf("package iteration: %#v", pkg)
		if strings.Contains(pkg, "|") {
			managerList := strings.Split(pkg, "|")
			var basePackage = ""
			var usablePackage = ""
			for _, mpkg := range managerList {
				log.Debug().Msgf("package iteration for manager: %#v", mpkg)
				if strings.Contains(mpkg, "=") {
					pmSplit := strings.Split(mpkg, "=")
					if pmSplit[0] == manager {
						usablePackage = pmSplit[1]
					}
				} else {
					basePackage = mpkg
				}
			}
			if usablePackage != "" {
				newList = append(newList, usablePackage)
			} else {
				newList = append(newList, basePackage)
			}
		}
		// no manager substitutes so just add it
		newList = append(newList, pkg)
	}
	return newList
}

func DoPackageManagerUpdate(packageHandler string) bool {
	updateComplete := false
	switch packageHandler {
	case "apt":
		updateComplete = updateApt()
		break
	case "yum":
		updateComplete = updateYum()
		break
	case "dnf":
		updateComplete = updateDnf()
		break
	}
	if !updateComplete {
		log.Info().Msg("no package manager to check for installed package, during packaging update")
		return false
	}
	return true
}

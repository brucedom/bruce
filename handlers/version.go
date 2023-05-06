package handlers

import "fmt"

func Version(currentVersion string) error {
	fmt.Println("CFS version: " + currentVersion)
	tag, err := getLatestTag("configset", "cfs")
	if err != nil {
		return err
	}
	fmt.Println("Latest version: " + tag)
	return nil
}

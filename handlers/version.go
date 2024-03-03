package handlers

import "fmt"

func Version(currentVersion string) error {
	fmt.Println("BRUCE version: " + currentVersion)
	tag, err := getLatestTag("brucedom", "bruce")
	if err != nil {
		return err
	}
	fmt.Println("Latest version: " + tag)
	return nil
}

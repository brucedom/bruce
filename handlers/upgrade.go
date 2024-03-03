package handlers

import (
	"bytes"
	"cfs/loader"
	"cfs/mutation"
	"encoding/json"
	"fmt"
	"github.com/minio/selfupdate"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"sort"
)

type RepositoryTag struct {
	Name string `json:"name"`
}

func Upgrade(currentVersion string) error {
	// this gets stripped by goreleaser
	currentVersion = "v" + currentVersion
	// then we need to get the latest version
	fmt.Println("Checking for updates... (current version is " + currentVersion + ")")
	fmt.Println("System information: " + runtime.GOOS + "/" + runtime.GOARCH)
	latestTag, err := getLatestTag("configset", "cfs")
	if err != nil {
		log.Fatalf("Error fetching latest tag: %s", err)
	}
	if latestTag == currentVersion {
		fmt.Println("You are already on the latest version!")
		return nil
	}
	fmt.Println("There is a new version available: " + latestTag)

	url := fmt.Sprintf("https://github.com/configset/cfs/releases/download/%s/cfs_%s_%s_%s.tar.gz", latestTag, latestTag[1:], runtime.GOOS, runtime.GOARCH)
	updateDir := path.Join(os.TempDir(), "cfs-update", fmt.Sprintf("%c", os.PathSeparator))
	fName := "cfs"
	if runtime.GOOS == "windows" {
		fName += ".exe"
	}
	err = mutation.ExtractTarball(url, updateDir, true, true)
	if err != nil {
		log.Fatalf("Error downloading tarball: %s", err)
	}
	rd, _, err := loader.GetRemoteData(path.Join(updateDir, fName))
	if err != nil {
		log.Fatalf("Error getting reader: %s", err)
	}
	reader := bytes.NewReader(rd)
	err = selfupdate.Apply(reader, selfupdate.Options{})
	if err != nil {
		log.Fatalf("Error updating binary (You may need to run upgrade with sudo / admin privileges): %s", err)
	}
	return err
}

func getLatestTag(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tags []RepositoryTag
	if err := json.Unmarshal(body, &tags); err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found for the repository")
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name > tags[j].Name
	})

	return tags[0].Name, nil
}

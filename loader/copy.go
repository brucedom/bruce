package loader

import (
	"cfs/exe"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CopyFile(src, dest string, perm os.FileMode, overwrite bool) error {
	source, _, err := GetRemoteReader(src)
	if err != nil {
		log.Error().Err(err).Msg("cannot open source file")
		return err
	}
	defer source.Close()

	if exe.FileExists(dest) {
		if overwrite {
			exe.DeleteFile(dest)
		} else {
			log.Error().Msgf("file %s already exists", dest)
			return fmt.Errorf("file %s already exists", dest)
		}
	} else {
		// check if the directories exist to render the file
		if !exe.FileExists(path.Dir(dest)) {
			os.MkdirAll(path.Dir(dest), perm)
		}
	}

	destination, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file for writing copy: %s", dest)
		return err
	}
	defer destination.Close()

	log.Debug().Str("copy", src).Msg("preparing to execute")

	len, err := io.Copy(destination, source)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
	}
	log.Debug().Msgf("copied %d bytes", len)
	return nil
}

func ReadRemoteIndex(remoteLoc string) ([]string, error) {
	resp, err := http.Get(remoteLoc)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	var links []string
	z := html.NewTokenizer(resp.Body)
	baseUrl, err := url.Parse(remoteLoc)
	if err != nil {
		return nil, err
	}

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return links, nil
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						link, err := baseUrl.Parse(attr.Val)
						if err != nil {
							log.Error().Err(err).Str("link", attr.Val).Msg("failed to parse link")
							continue
						}
						if link.Scheme == "http" || link.Scheme == "https" {
							continue
						}
						links = append(links, link.String())
					}
				}
			}
		}
	}
}

// for future use
func RecursiveHTTPCopy(src string, dest string, overwrite bool) error {
	// Read the remote index
	remoteIndex, err := ReadRemoteIndex(src)
	if err != nil {
		return err
	}

	// Make the destination directory if it doesn't exist
	if !exe.FileExists(dest) {
		err := os.MkdirAll(dest, 0775)
		if err != nil {
			log.Error().Err(err).Msgf("could not create directory: %s", dest)
			return err
		}
	}

	// Copy files and recurse into directories
	for _, url := range remoteIndex {
		destPath := filepath.Join(dest, path.Base(url))
		if IsDirectory(url) {
			err := RecursiveHTTPCopy(url, destPath, overwrite)
			if err != nil {
				return err
			}
		} else {
			err := CopyFile(url, destPath, 0664, overwrite)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// IsDirectory checks if the given URL is a directory
func IsDirectory(url string) bool {
	return strings.HasSuffix(url, "/")
}

package loader

import (
	"cfs/exe"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type PageLink struct {
	Target string
	Text   string
}

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

	sln, err := io.Copy(destination, source)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
	}
	log.Debug().Msgf("copied %d bytes", sln)
	return nil
}

func ReadRemoteIndex(remoteLoc string) ([]PageLink, error) {
	resp, err := http.Get(remoteLoc)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Debug().Msgf("received status code %d", resp.StatusCode)
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var links []PageLink
	// The regex pattern to match <a href="...">...</a> tags
	pattern := regexp.MustCompile(`<a\s+href="([^"]+)">([^<]+)<\/a>`)

	matches := pattern.FindAllSubmatch(content, -1)
	for _, match := range matches {
		link := PageLink{Target: string(match[1]), Text: string(match[2])}
		links = append(links, link)
	}

	return links, nil

}

func RecursiveCopy(src string, baseDir, dest string, overwrite bool, ignores []string, isFlatCopy bool, maxDepth int) error {
	return recursiveCopyInternal(src, baseDir, dest, overwrite, ignores, isFlatCopy, maxDepth, 0)
}

// for future use
func recursiveCopyInternal(src string, baseDir, dest string, overwrite bool, ignores []string, isFlatCopy bool, maxDepth, currentDepth int) error {
	if dest == "" {
		dest = baseDir
	}

	log.Debug().Str("src", src).Str("dest", dest).Msg("recursively copying")
	// Read the remote index
	remoteIndex, err := ReadRemoteIndex(src)
	if err != nil {
		log.Error().Err(err).Msg("could not read remote index")
		return err
	}
	log.Debug().Interface("remoteIndex", remoteIndex).Msg("remote index")

	// Make the destination directory if it doesn't exist
	if !exe.FileExists(dest) {
		err := os.MkdirAll(dest, 0775)
		if err != nil {
			log.Error().Err(err).Msgf("could not create directory: %s", dest)
			return err
		}
	}

	// Copy files and recurse into directories
	for _, link := range remoteIndex {
		if link.Target == "../" || link.Target == "./" || len(link.Target) == 0 {
			continue
		}
		log.Debug().Interface("url", src).Msg("src")
		srcUrl, err := joinURL(src, link.Target)
		if err != nil {
			log.Error().Err(err).Msg("could not join url")
			return err
		}
		log.Debug().Str("srcUrl", srcUrl).Msg("srcUrl")

		if link.Target[len(link.Target)-1:] == "/" {
			// This is a directory
			destPath := filepath.Join(dest, path.Base(link.Target))
			log.Debug().Str("url", link.Target).Str("destPath", destPath).Msg("copying directory")
			targetUrl, err := joinURL(src, link.Target)
			if err != nil {
				log.Error().Err(err).Msg("could not join url")
				return err
			}

			if maxDepth == 0 || currentDepth+1 < maxDepth {
				if !isFlatCopy {
					err = recursiveCopyInternal(targetUrl, baseDir, destPath, overwrite, ignores, isFlatCopy, maxDepth, currentDepth+1)
					if err != nil {
						return err
					}
				} else {
					err = recursiveCopyInternal(targetUrl, baseDir, baseDir, overwrite, ignores, isFlatCopy, maxDepth, currentDepth+1)
					if err != nil {
						return err
					}
				}
			}
		} else {
			// This is a file
			destPath := filepath.Join(isFlatCopyDest(baseDir, dest, isFlatCopy), path.Base(link.Target))
			shouldIgnore := false
			for _, ignore := range ignores {
				if strings.Contains(destPath, ignore) {
					shouldIgnore = true
				}
			}
			if shouldIgnore {
				log.Info().Msgf("skipping ignored file: %s", srcUrl)
			} else {
				log.Info().Msgf("saving file: %s", destPath)
				err := CopyFile(srcUrl, destPath, 0664, overwrite)
				if err != nil {
					return err
				}
			}

		}

		log.Debug().Str("url", link.Target).Msg("copying file")
	}
	return nil
}

func isFlatCopyDest(baseDir, dest string, isFlatCopy bool) string {
	if isFlatCopy {
		return baseDir
	}
	return dest
}

func joinURL(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	pathURL, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	joinedURL := u.ResolveReference(pathURL)
	return joinedURL.String(), nil
}

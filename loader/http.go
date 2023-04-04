package loader

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
)

func ReaderFromHttp(fileName string) (io.ReadCloser, string, error) {
	req, err := http.NewRequest("GET", fileName, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	fn := path.Base(resp.Request.URL.String())
	return resp.Body, fn, nil
}

func ReadRemoteHttpIndex(remoteLoc string) ([]PageLink, error) {
	log.Debug().Msgf("reading remote http index: %s", remoteLoc)
	resp, err := http.Get(remoteLoc)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Debug().Msgf("received status code %d", resp.StatusCode)
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}
	log.Debug().Msg("received okay status code, proceeding to read body")
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("could not read response body")
		return nil, err
	}
	log.Debug().Msgf("read %d bytes", len(content))
	var links []PageLink
	// The regex pattern to match <a href="...">...</a> tags
	pattern := regexp.MustCompile(`<a\s+href="([^"]+)">([^<]+)<\/a>`)

	matches := pattern.FindAllSubmatch(content, -1)
	for _, match := range matches {
		link := PageLink{Target: string(match[1]), Text: string(match[2])}
		links = append(links, link)
	}
	log.Debug().Msgf("found %d links", len(links))
	return links, nil
}

func GetHttpRecursiveList(src string, maxDepth int) ([]string, error) {
	return getHttpRecursiveListWithDepth(src, maxDepth, 0)
}

func getHttpRecursiveListWithDepth(src string, maxDepth, currentDepth int) ([]string, error) {
	var files []string
	if strings.HasSuffix(src, "/") {
		log.Debug().Msgf("List remote http index: %s", src)
		links, err := ReadRemoteHttpIndex(src)
		if err != nil {
			log.Debug().Err(err).Msg("could not read remote http index")
			return nil, err
		}
		for _, link := range links {
			if link.Target == "../" || link.Target == "./" || len(link.Target) == 0 {
				continue
			}
			nurl, err := joinURL(src, link.Target)
			if strings.HasSuffix(link.Target, "/") {
				log.Debug().Msgf("found directory: %s", link.Target)
				// This is a directory, we need to recursively list it

				if err != nil {
					log.Debug().Err(err).Msg("could not join url")
					return nil, err
				}
				if maxDepth == 0 || currentDepth < maxDepth {
					dfiles, err := getHttpRecursiveListWithDepth(nurl, maxDepth, currentDepth+1)
					if err != nil {
						return nil, err
					}
					files = append(files, dfiles...)
				}
			} else {
				files = append(files, nurl)
				log.Debug().Msgf("found link: %s", nurl)
			}
		}
	} else {
		log.Debug().Msgf("adding file to list: %s", src)
		files = append(files, src)
	}
	return files, nil
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

func downloadFile(src, dest string, overwrite bool, wg *sync.WaitGroup, semaphore chan struct{}) {
	defer wg.Done()

	semaphore <- struct{}{}
	err := CopyFile(src, dest, 0664, overwrite)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
	}
	<-semaphore
}

func recursiveHttpCopy(src string, baseDir, dest string, overwrite bool, ignores []string, isFlatCopy bool, maxDepth, maxConcurrent int) error {
	if dest == "" {
		dest = baseDir
	}
	log.Debug().Str("src", src).Str("dest", dest).Msg("recursively copying")
	list, err := GetHttpRecursiveList(src, maxDepth)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent)
	for _, file := range list {
		lFile := strings.Replace(file, src, "", 1)
		aDest := isFlatCopyDest(lFile, baseDir, dest, isFlatCopy)

		// Check if the file or directory should be ignored
		shouldIgnore := false
		for _, ignore := range ignores {
			if strings.Contains(lFile, ignore) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		wg.Add(1)
		go downloadFile(file, aDest, overwrite, &wg, semaphore)
	}

	wg.Wait()
	return nil
}

package loader

import (
	"bytes"
	"cfs/exe"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path"
)

type PageLink struct {
	Target string
	Text   string
}

func CopyFile(src, dest string, perm os.FileMode, overwrite bool) error {
	// if filemode is 0, set it to 0644
	if perm == 0 {
		perm = 0644
	}
	sd, _, err := GetRemoteData(src)
	if err != nil {
		log.Error().Err(err).Msg("cannot open source file")
		return err
	}
	// create a io.reader from sd
	source := bytes.NewReader(sd)
	if exe.FileExists(dest) {
		if overwrite {
			log.Err(exe.DeleteFile(dest))
		} else {
			log.Error().Msgf("file %s already exists", dest)
			return fmt.Errorf("file %s already exists", dest)
		}
	} else {
		// check if the directories exist to render the file
		if !exe.FileExists(path.Dir(dest)) {
			log.Err(os.MkdirAll(path.Dir(dest), perm))
		}
	}

	destination, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file for writing copy: %s", dest)
		return err
	}
	defer log.Err(destination.Close())
	printSrc := src
	if len(src) > 32 {
		printSrc = "..." + src[len(src)-32:]
	}
	printDest := dest
	if len(dest) > 32 {
		printDest = "..." + dest[len(dest)-32:]
	}
	log.Info().Msgf("copying %s ==> %s", printSrc, printDest)

	sln, err := io.Copy(destination, source)
	if err != nil {
		log.Error().Err(err).Msg("could not copy file")
	}
	log.Debug().Msgf("copied %d bytes", sln)
	return nil
}

func RecursiveCopy(src string, baseDir, dest string, overwrite bool, ignores []string, isFlatCopy bool, maxDepth, maxConcurrent int) error {
	if src[0:4] == "http" {
		// This is a remote http copy
		return recursiveHttpCopy(src, baseDir, dest, overwrite, ignores, isFlatCopy, maxDepth, maxConcurrent)
	}
	if src[0:5] == "s3://" {
		// This is a remote s3 copy
		return recursiveS3Copy(src, baseDir, dest, overwrite, ignores, isFlatCopy, maxDepth, maxConcurrent)
	}
	return recursiveNotSupported(src, baseDir, dest, overwrite, ignores, isFlatCopy, maxDepth)
}

func recursiveNotSupported(_ string, _, _ string, _ bool, _ []string, _ bool, _ int) error {
	log.Error().Msg("recursive copy not supported for this source")
	return fmt.Errorf("recursive copy not supported for this source")
}

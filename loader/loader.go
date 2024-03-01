package loader

import (
	"strings"
)

func ReadRemoteFile(remoteLoc string) ([]byte, string, error) {
	d, fn, err := GetRemoteData(remoteLoc)
	if err != nil {
		return nil, fn, err
	}
	return d, fn, err
}

// GetRemoteData returns a ReadCloser with a filename and error if exists.
func GetRemoteData(remoteLoc string) ([]byte, string, error) {
	if strings.ToLower(remoteLoc[0:4]) == "http" {
		return ReadFromHttp(remoteLoc)
	}
	if strings.ToLower(remoteLoc[0:5]) == "s3://" {
		return ReadFromS3(remoteLoc)
	}
	// if no remote handlers can handle the reading of the file, lets try local
	return ReadFromLocal(remoteLoc)
}

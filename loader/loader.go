package loader

import (
	"io"
	"strings"
)

func ReadRemoteFile(remoteLoc string) ([]byte, string, error) {
	r, fn, err := GetRemoteReader(remoteLoc)
	if err != nil {
		return nil, fn, err
	}
	defer r.Close()
	d, err := io.ReadAll(r)
	return d, fn, err
}

// GetRemoteReader returns a readcloser with a filename and error if exists.
func GetRemoteReader(remoteLoc string) (io.ReadCloser, string, error) {
	if strings.ToLower(remoteLoc[0:4]) == "http" {
		return ReaderFromHttp(remoteLoc)
	}
	if strings.ToLower(remoteLoc[0:5]) == "s3://" {
		return ReaderFromS3(remoteLoc)
	}
	// if no remote handlers can handle the reading of the file, lets try local
	return ReaderFromLocal(remoteLoc)
}

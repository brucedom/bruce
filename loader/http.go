package loader

import (
	"io"
	"net/http"
	"path"
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

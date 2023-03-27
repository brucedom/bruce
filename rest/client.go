package rest

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

// RESTClient is a default http rest client and holds the connection info and HttpClient
type RESTClient struct {
	Host       string
	HTTPClient *http.Client
}

// NewRestClient is a constructor for setting up a new RESTClient
func NewRestClient(host string, tlsSkipVerify bool) (*RESTClient, error) {
	if host == "" {
		return nil, fmt.Errorf("Host cannot be blank.")
	}
	restClient := &RESTClient{Host: host}
	if tlsSkipVerify {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		restClient.HTTPClient = &http.Client{Transport: tr}
	}
	return restClient, nil
}

// Get runs a http GET against the specified endpoint and encodes the result or returns error.
func (restClient *RESTClient) Get(endpoint string, headers map[string]string, obj interface{}) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	req.Close = true

	resp, err := restClient.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&obj); err != nil {
		return err
	}

	return nil
}

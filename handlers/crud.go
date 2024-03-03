package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ContentUpload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type ContentItem struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ItemResponse struct {
	Result  []ContentItem `json:"result"`
	Message string        `json:"message"`
}

func readContent(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read up to 8192 bytes to check for binary content
	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Check for null byte, which indicates binary content
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return nil, fmt.Errorf("binary files are not supported")
		}
	}

	// Rewind the file to the beginning
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// Read the entire file
	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func GetItem(kind, id string) *ItemResponse {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://brucedom.com/api/%ss", kind), nil)
	if err != nil {
		fmt.Println("Something went wrong: " + err.Error())
		return nil
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Sprintf("Error: %s", resp.Status))
		return nil
	}
	bd, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	d := &ItemResponse{}
	err = json.Unmarshal(bd, d)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return d
}

func Edit(kind, id, filename string) {
	key := os.Getenv("BRUCE_KEY")
	if key == "" {
		fmt.Println("BRUCE_KEY environment variable not set, please set this to your API key from brucedom.com")
		return
	}
	if kind == "" {
		fmt.Println("Kind not set, please set this to either 'template' or 'manifest'")
		return
	}
	if id == "" {
		fmt.Println("ID not set, please set this to the ID of the item you wish to edit")
		return
	}
	if filename == "" {
		fmt.Println("Filename not set, please set this to the path of the file you wish to upload")
		return
	}
	fmt.Println(fmt.Sprintf("Editing [%s] with ID: %s from: %s", kind, id, filename))
	item := GetItem(kind, id)
	if item == nil {
		fmt.Println("Error getting item")
		return
	}
	data, err := readContent(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	var content ContentUpload
	for _, i := range item.Result {
		if i.Id == id {
			content.Name = i.Name
			content.Description = i.Description
			content.Content = string(data)
		}
	}
	// encode content to be set as the body of the request
	body, err := json.Marshal(content)
	if err != nil {
		fmt.Println(err)
		return
	}
	// make an io.reader from the body
	bodyReader := io.NopCloser(bytes.NewReader(body))
	// make a request
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://brucedom.com/api/%ss/%s", kind, id), bodyReader)
	if err != nil {
		fmt.Println(err)
		return
	}
	// set the content type
	req.Header.Set("Content-Type", "application/json")
	// set the api key
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	// make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Sprintf("Error: %s", resp.Status))
		return
	}

	fmt.Println(fmt.Sprintf("Success: %s", resp.Status))
}

func Create(kind, name, description, filename string) {
	key := os.Getenv("BRUCE_KEY")
	if key == "" {
		fmt.Println("BRUCE_KEY environment variable not set, please set this to your API key from brucedom.com")
		return
	}
	if kind == "" {
		fmt.Println("Kind not set, please set this to either 'template' or 'manifest'")
		return
	}
	if name == "" {
		fmt.Println("Name not set, please set this to a name for your item")
		return
	}
	if description == "" {
		fmt.Println("Description not set, please set this to a brief description for your item")
		return
	}
	if filename == "" {
		fmt.Println("Filename not set, please set this to the path of the file you wish to upload")
		return
	}
	// if the kind is manifest, then we post a json object containing keys "name, description, content" to /api/manifests

	fmt.Println(fmt.Sprintf("Creating [%s] with name: %s from: %s", kind, name, filename))
	// read the file and check that it is non binary:
	data, err := readContent(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("Creating [%s]: %s", kind, name))
	content := ContentUpload{
		Name:        name,
		Description: description,
		Content:     string(data),
	}
	// encode content to be set as the body of the request
	body, err := json.Marshal(content)
	if err != nil {
		fmt.Println(err)
		return
	}
	// make an io.reader from the body
	bodyReader := io.NopCloser(bytes.NewReader(body))

	req, err := http.NewRequest("POST", fmt.Sprintf("https://brucedom.com/api/%ss", kind), bodyReader)
	// now submit the request
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Sprintf("Error: %s", resp.Status))
		return
	}
	fmt.Println(fmt.Sprintf("Success: %s", resp.Status))
}

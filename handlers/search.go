package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SearchResponse struct {
	Result  []Manifest `json:"result"`
	Message string     `json:"message"`
}

type Manifest struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Author      string    `json:"author"`
	AuthorID    string    `json:"author_id"`
	Description string    `json:"description"`
	Likes       int       `json:"likes"`
	Dislikes    int       `json:"dislikes"`
	IsFlagged   bool      `json:"is_flagged"`
	Format      string    `json:"format"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

func Search(q string) error {
	// do an http client lookup for the q string against the cfs server and print the results
	manifestUrl := "https://configset.dev/api/manifests"

	req, err := http.NewRequest("GET", manifestUrl+"?search="+q, nil)
	if err != nil {
		fmt.Println("Something went wrong: " + err.Error())
		return err
	}
	fmt.Printf("Searching for: %s\n", q)
	fmt.Println("=====================================")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: status code %d\n", resp.StatusCode)
		return fmt.Errorf("Error: status code %d", resp.StatusCode)
	}

	var sr SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		fmt.Println("Error decoding JSON response: " + err.Error())
		return err
	}

	for _, manifest := range sr.Result {
		fmt.Printf("Name: %s\n  Description: %s\n  URL: %s\n", manifest.Name, manifest.Description, fmt.Sprintf("%s/%s/data", manifestUrl, manifest.Id))
	}

	return nil
}

package handlers

import (
	"context"
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

func Search(query, q2 string) error {
	searchMode := "manifests"
	q := query
	if q2 != "" {
		q = q2
		if query == "manifests" || query == "manifest" || query == "m" {
			searchMode = "manifests"
		} else if query == "templates" || query == "template" || query == "t" {
			searchMode = "templates"
		} else {
			fmt.Printf("Invalid object search: \"%s\" Choose one of (manifests/manifest/m) or (templates/template/t)\n", query)
			return fmt.Errorf("invalid search mode: %s only (manifests/manifest/m) or (templates/template/t) is supported", q)
		}
	}

	searchUrl := "https://configset.com/api/" + searchMode

	req, err := http.NewRequest("GET", searchUrl+"?search="+q, nil)
	if err != nil {
		fmt.Println("Something went wrong: " + err.Error())
		return err
	}
	fmt.Printf("Searching for (%s): %s\n", searchMode, q)
	fmt.Println("=====================================")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}

	// Create a context with a 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add the context to the request
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		// Check if the error is due to a timeout
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Request timed out. Please check your firewall settings or try again later.")
			return err
		}

		fmt.Println("Error making request: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: status code %d\n", resp.StatusCode)
		return fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	var sr SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		fmt.Println("Error decoding JSON response: " + err.Error())
		return err
	}

	for _, searchItem := range sr.Result {
		fmt.Printf("Name: %s\n  Description: %s\n  URL: %s\n", searchItem.Name, searchItem.Description, fmt.Sprintf("%s/%s/data", searchUrl, searchItem.Id))
	}

	return nil
}

// wikipedia/client.go
package wikipedia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Article struct {
	Title   string `json:"title"`
	Extract string `json:"extract"`
}

type Client struct {
	BaseURL string
}

func NewClient() *Client {
	return &Client{
		BaseURL: "https://en.wikipedia.org/api/rest_v1/page/summary/",
	}
}

func (c *Client) GetArticle(title string) (*Article, error) {
	// URL encode the title
	encodedTitle := url.PathEscape(title)
	
	resp, err := http.Get(c.BaseURL + encodedTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikipedia API returned status: %s", resp.Status)
	}

	var article Article
	if err := json.NewDecoder(resp.Body).Decode(&article); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &article, nil
}
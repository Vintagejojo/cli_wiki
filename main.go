package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Wikipedia API response structures
type SearchResponse struct {
	Pages []struct {
		Title       string `json:"title"`
		Key         string `json:"key"`
		Description string `json:"description"`
		Thumbnail   struct {
			URL string `json:"url"`
		} `json:"thumbnail"`
		Excerpt string `json:"excerpt"`
	} `json:"pages"`
}

type LanguageLinksResponse []struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// Model for Bubble Tea TUI
type Model struct {
	query          string
	language       string
	results        []string
	selectedResult int
	languageLinks  []string
	viewMode       string // "search" or "languages"
}

func initialModel(query, language string) Model {
	return Model{
		query:    query,
		language: language,
		viewMode: "search",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.selectedResult > 0 {
				m.selectedResult--
			}
		case "down":
			if m.selectedResult < len(m.results)-1 {
				m.selectedResult++
			}
		case "enter":
			if len(m.results) > 0 && m.viewMode == "search" {
				title := url.PathEscape(m.results[m.selectedResult])
				links, err := getLanguageLinks(title, m.language)
				if err != nil {
					m.results = []string{fmt.Sprintf("Error: %v", err)}
					return m, nil
				}
				m.languageLinks = links
				m.viewMode = "languages"
			}
		case "esc":
			if m.viewMode == "languages" {
				m.viewMode = "search"
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := fmt.Sprintf("Wikipedia Search (%s): %s\n\n", m.language, m.query)

	if m.viewMode == "search" {
		for i, result := range m.results {
			cursor := " "
			if i == m.selectedResult {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, result)
		}
		s += "\n↑/↓: navigate • Enter: select • q: quit"
	} else {
		s += "Available in other languages:\n\n"
		for _, link := range m.languageLinks {
			s += fmt.Sprintf("• %s\n", link)
		}
		s += "\nEsc: back to results • q: quit"
	}

	return s
}

// Wikipedia API functions
func searchWikipedia(query, language string, limit int) ([]string, error) {
	baseURL := fmt.Sprintf("https://api.wikimedia.org/core/v1/wikipedia/%s/search/page", language)
	
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", "cli_wiki (your-email@example.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var titles []string
	for _, page := range result.Pages {
		desc := page.Description
		if desc == "" {
			desc = "No description available"
		}
		titles = append(titles, fmt.Sprintf("%s - %s", page.Title, desc))
	}

	return titles, nil
}

func getLanguageLinks(title, language string) ([]string, error) {
	baseURL := fmt.Sprintf("https://api.wikimedia.org/core/v1/wikipedia/%s/page/%s/links/language", language, url.PathEscape(title))
	
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "cli_wiki (your-email@example.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LanguageLinksResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var links []string
	for _, lang := range result {
		links = append(links, fmt.Sprintf("%s: https://%s.wikipedia.org/wiki/%s", 
			lang.Name, lang.Code, url.PathEscape(lang.Key)))
	}

	return links, nil
}

func main() {
	query := flag.String("query", "Solar System", "Search query")
	language := flag.String("lang", "en", "Wikipedia language code")
	limit := flag.Int("limit", 5, "Number of results to return")
	flag.Parse()

	results, err := searchWikipedia(*query, *language, *limit)
	if err != nil {
		fmt.Printf("Error searching Wikipedia: %v\n", err)
		os.Exit(1)
	}

	model := initialModel(*query, *language)
	model.results = results

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
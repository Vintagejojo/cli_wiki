// ui/article.go
package ui

import (
	"cli_wiki/wikipedia"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


type ArticleModel struct {
	article   *wikipedia.Article
	viewport  viewport.Model
	ready     bool
}

func NewArticleModel(article *wikipedia.Article) ArticleModel {
	return ArticleModel{
		article: article,
	}
}

func (m ArticleModel) Init() tea.Cmd {
	return nil
}

func (m ArticleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			// Initialize viewport
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.SetContent(m.article.Extract)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	// Handle viewport keybindings
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ArticleModel) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	// Create a nice title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render(m.article.Title)

	// Combine title and content
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.viewport.View(),
	)
}
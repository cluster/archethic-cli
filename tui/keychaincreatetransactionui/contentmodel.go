package keychaincreatetransactionui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type ContentModel struct {
	contentTextAreaInput textarea.Model
}

type UpdateContent struct {
	Content []byte
}

func NewContentModel() ContentModel {
	m := ContentModel{}
	m.contentTextAreaInput = textarea.New()
	// passing 0 or a negative number here doesn't seem to work...
	m.contentTextAreaInput.CharLimit = 100000000000
	m.contentTextAreaInput.SetWidth(150)

	return m
}

func (m ContentModel) Init() tea.Cmd {
	return nil
}

func (m ContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "esc":

			if m.contentTextAreaInput.Focused() {
				m.contentTextAreaInput.Blur()
			}
			return m, nil

		default:

			if !m.contentTextAreaInput.Focused() {
				m.contentTextAreaInput.Focus()
			}
			m.contentTextAreaInput, _ = m.contentTextAreaInput.Update(msg)
			return m, func() tea.Msg {
				return UpdateContent{Content: []byte(m.contentTextAreaInput.Value())}
			}

		}
	}
	return m, nil
}

func (m *ContentModel) SwitchTab() (ContentModel, []tea.Cmd) {
	return *m, nil
}

func (m ContentModel) View() string {
	var b strings.Builder

	b.WriteString(m.contentTextAreaInput.View())

	if m.contentTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}

	return b.String()
}

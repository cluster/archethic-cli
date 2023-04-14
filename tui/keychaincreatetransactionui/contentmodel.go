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

			m.contentTextAreaInput, _ = m.contentTextAreaInput.Update(msg)
			return m, func() tea.Msg {
				return UpdateContent{Content: []byte(m.contentTextAreaInput.Value())}
			}

		}
	}
	m, cmds := updateContentFocus(m)
	cmds = append(cmds, m.updateContentInputs(msg)...)

	return m, tea.Batch(cmds...)
}

func (m *ContentModel) updateContentInputs(msg tea.Msg) []tea.Cmd {

	cmds := make([]tea.Cmd, 1)
	m.contentTextAreaInput, cmds[0] = m.contentTextAreaInput.Update(msg)

	return cmds
}

func updateContentFocus(m ContentModel) (ContentModel, []tea.Cmd) {

	cmds := make([]tea.Cmd, 1)
	cmds[0] = m.contentTextAreaInput.Focus()

	return m, cmds
}

func (m *ContentModel) SwitchTab() (ContentModel, []tea.Cmd) {
	m.contentTextAreaInput.Focus()
	m2, cmds := updateContentFocus(*m)
	return m2, cmds
}

func (m ContentModel) View() string {
	var b strings.Builder

	b.WriteString(m.contentTextAreaInput.View())

	if m.contentTextAreaInput.Focused() {
		b.WriteString(helpStyle.Render("\npress 'esc' to exit edit mode "))
	}

	return b.String()
}

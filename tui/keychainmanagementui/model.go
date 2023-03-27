package keychainmanagementui

import (
	"fmt"
	"sort"
	"strings"

	archethic "github.com/archethic-foundation/libgo"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BackMsg change state back to project view
type BackMsg bool

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func GetNodeEndpointURL(name string) string {
	switch name {
	case "L", "l":
		return "http://localhost:4000"
	case "T", "t":
		return "https://testnet.archethic.org"
	case "M", "m":
		return "https://mainnet.archethic.org"
	default:
		return name
	}
}

type Model struct {
	focusIndex   int
	inputs       []textinput.Model
	keychain     *archethic.Keychain
	serviceNames []string
}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle

		switch i {
		case 0:
			t.Prompt = "> Node endpoint:\n"
			t.Focus()
			t.Placeholder = "(default local)"
		case 1:
			t.Prompt = "> Access seed\n"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return BackMsg(true)
			}

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				var tmpUrl string
				if m.inputs[0].Value() == "" {
					tmpUrl = "l"
				} else {
					tmpUrl = m.inputs[0].Value()
				}
				url := GetNodeEndpointURL(tmpUrl)
				seed := []byte(m.inputs[1].Value())
				client := archethic.NewAPIClient(url, "")
				m.keychain = archethic.GetKeychain(seed, *client)
				m.serviceNames = make([]string, 0, len(m.keychain.Services))
				for k := range m.keychain.Services {
					m.serviceNames = append(m.serviceNames, k)
				}

				sort.Strings(m.serviceNames)

			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					// m.inputs[i].PromptStyle = focusedStyle
					// m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i == 0 {
			b.WriteString("\n\tL for local\n\tM for mainnet\n\tT for testnet\n\tor enter a custom endpoint\n")
		}
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.keychain != nil {
		for _, k := range m.serviceNames {
			b.WriteString(k + " : " + m.keychain.Services[k].DerivationPath + "\n")
		}
	}

	b.WriteString(helpStyle.Render("press 'esc' to go back "))

	return b.String()
}

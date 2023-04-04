package keychainmanagementui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/archethic-foundation/archethic-cli/tui/keychaincreatetransactionui"
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

	accessFocusedButton = focusedStyle.Copy().Render("[ Access Keychain ]")
	accessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Access Keychain"))

	createTransactionFocusedButton       = focusedStyle.Copy().Render("[ Create Transaction ]")
	createTransactionaccessBlurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Create Transaction"))
	urlType                              = []string{"Local", "Testnet", "Mainnet", "Custom"}
	urls                                 = map[string]string{
		"Local":   "http://localhost:4000",
		"Testnet": "https://testnet.archethic.net",
		"Mainnet": "https://mainnet.archethic.net",
		"Custom":  ""}
)

type Model struct {
	focusIndex      int
	inputs          []textinput.Model
	keychain        *archethic.Keychain
	serviceNames    []string
	selectedUrl     string
	selectedService int
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
			t.Prompt = ""
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
			if s == "enter" && m.focusIndex < 4 {
				u := urlType[m.focusIndex]
				m.inputs[0].SetValue(urls[u])
				m.selectedUrl = u
				m.focusIndex = 3
			}

			if s == "enter" && m.focusIndex == len(m.inputs)+4 {
				url := m.inputs[0].Value()
				seed := archethic.MaybeConvertToHex(m.inputs[1].Value())
				client := archethic.NewAPIClient(url)
				m.keychain = archethic.GetKeychain(seed, *client)
				m.serviceNames = make([]string, 0, len(m.keychain.Services))
				for k := range m.keychain.Services {
					m.serviceNames = append(m.serviceNames, k)
				}

				sort.Strings(m.serviceNames)

			}

			if s == "enter" && m.focusIndex > len(m.inputs)+4 {
				m.selectedService = m.focusIndex - len(m.inputs) - 5
			}

			if s == "enter" && m.focusIndex == len(m.inputs)+5+len(m.serviceNames) {
				return m, func() tea.Msg {
					return keychaincreatetransactionui.CreateTransactionMsg{
						ServiceName: m.serviceNames[m.selectedService-1],
						Url:         m.inputs[0].Value(),
						Seed:        m.inputs[1].Value(),
					}
				}
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			serviceSize := len(m.serviceNames)
			// if there is at least one service, there is a button to create a transaction
			if serviceSize > 0 {
				serviceSize++
			}
			if m.focusIndex > len(m.inputs)+4+serviceSize {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 4 + serviceSize
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex-4 {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
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

	b.WriteString("> Node endpoint:\n")
	b.WriteString(urlView(m))
	for i := range m.inputs {
		b.WriteRune('\n')
		b.WriteString(m.inputs[i].View())
	}

	button := &accessBlurredButton
	if m.focusIndex == len(m.inputs)+4 {
		button = &accessFocusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.keychain != nil {
		for i, k := range m.serviceNames {

			var u string
			if m.selectedService == i {
				u = "(•) "
			} else {
				u = "( ) "
			}
			u += k + " : " + m.keychain.Services[k].DerivationPath + "\n"
			if m.focusIndex == i+len(m.inputs)+5 {
				b.WriteString(focusedStyle.Render(u))
			} else {
				b.WriteString(u)
			}
			b.WriteString("\n")
		}

		button := &createTransactionaccessBlurredButton
		if m.focusIndex == len(m.inputs)+len(m.serviceNames)+5 {
			button = &createTransactionFocusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	}

	b.WriteString(helpStyle.Render("press 'esc' to go back "))

	return b.String()
}

func urlView(m Model) string {
	s := strings.Builder{}

	for i := 0; i < len(urlType); i++ {
		var u string
		if m.selectedUrl == urlType[i] {
			u = "(•) "
		} else {
			u = "( ) "
		}
		u += urlType[i]
		if i == m.focusIndex {
			s.WriteString(focusedStyle.Render(u))
		} else {
			s.WriteString(u)
		}
		s.WriteString("\n")
	}

	return s.String()
}
